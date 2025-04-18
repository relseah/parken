package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/relseah/parken"
	"github.com/relseah/parken/nominatim"
	"github.com/relseah/parken/scraping"
)

const timeLayout = "2006-01-02 15:04:05"

type Server struct {
	*http.Server
	Scraper *scraping.Scraper
	Client  *nominatim.Client
	Logger  *log.Logger

	cache []byte

	parkings      []parken.Parking
	updated       time.Time
	coordinates   map[int]parken.Coordinates
	presets       map[int]parken.Coordinates
	coordinatesDB map[int]parken.Coordinates

	db                    *sql.DB
	dbMutex               sync.Mutex
	insertCoordinatesStmt *sql.Stmt
	insertSpotsStmt       *sql.Stmt

	ticker *time.Ticker
	done   chan struct{}
}

func (s *Server) logf(format string, v ...any) {
	logger := s.Logger
	if logger != nil {
		logger.Printf(format, v...)
	}
}

func (s *Server) logln(v ...any) {
	logger := s.Logger
	if logger != nil {
		logger.Println(v...)
	}
}

var errorMessages = map[int]string{
	http.StatusNotFound:            "Not Found",
	http.StatusInternalServerError: "Internal Server Error",
}

func httpError(w http.ResponseWriter, code int) {
	http.Error(w, fmt.Sprintf("%d %s", code, errorMessages[code]), code)
}

func (s *Server) parkingsHandler(w http.ResponseWriter, r *http.Request) {
	// The correct Content-Type is not detected.
	w.Header().Set("Content-Type", "application/json")
	w.Write(s.cache)
}

func compressedFileServer(root http.FileSystem, extensions []string) http.Handler {
	handler := http.FileServer(root)
	// Unclear, whether the mime package caches the types.
	types := make(map[string]string)
	for _, extension := range extensions {
		types[extension] = mime.TypeByExtension(extension)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, extension := range extensions {
			if strings.HasSuffix(r.URL.Path, extension) {
				accepted := strings.Split(r.Header.Get("Accept-Encoding"), ", ")
				for _, encoding := range accepted {
					if encoding == "gzip" {
						r.URL.Path += ".gz"
						w.Header().Add("Content-Encoding", "gzip")
						w.Header().Add("Content-Type", types[extension])
						break
					}
				}
				break
			}
		}
		handler.ServeHTTP(w, r)
	})
}

func (s *Server) queryCoordinates() error {
	s.dbMutex.Lock()
	defer s.dbMutex.Unlock()
	rows, err := s.DB().Query("SELECT parking_id, latitude, longitude FROM coordinates;")
	if err != nil {
		return err
	}
	defer rows.Close()
	coordinates := make(map[int]parken.Coordinates)
	var id int
	var current parken.Coordinates
	for rows.Next() {
		rows.Scan(&id, &current.Latitude, &current.Longitude)
		coordinates[id] = current
	}
	if err := rows.Err(); err != nil {
		return err
	}
	s.coordinatesDB = coordinates
	return nil
}

func (s *Server) obtainCoordinates(p *parken.Parking) (parken.Coordinates, error) {
	if preset, ok := s.presets[p.ID]; ok {
		return preset, nil
	}
	if coordinates, ok := s.coordinatesDB[p.ID]; ok {
		return coordinates, nil
	}
	results, err := s.Client.Search(p)
	if err != nil {
		return parken.Coordinates{}, fmt.Errorf("searching for coordinates of parking with ID %d: %w", p.ID, err)
	}
	if len(results) == 0 {
		s.logf("No results for parking P%d %s.\n", p.ID, p.Name)
	} else if len(results) > 1 {
		var b strings.Builder
		fmt.Fprintf(&b, "Multiple results for parking P%d %s.\n", p.ID, p.Name)
		for i, c := range results {
			fmt.Fprintf(&b, "%d. Latitude: %f°, longitude: %f°\n", i, c.Latitude, c.Longitude)
		}
		logger := s.Logger
		if logger != nil {
			logger.Print(b.String())
		}
	} else {
		coordinates := results[0]
		_, err = s.insertCoordinatesStmt.Exec(p.ID, coordinates.Latitude, coordinates.Longitude)
		return coordinates, err
	}
	return parken.Coordinates{}, nil
}

func (s *Server) scrape() error {
	res, err := s.Scraper.Scrape(s.updated)
	if err != nil {
		if err == scraping.ErrNoUpdate {
			return nil
		}
		return err
	}
	for i := 0; i < len(res.Parkings); i++ {
		p := &res.Parkings[i]
		if coordinates, ok := s.coordinates[p.ID]; ok {
			p.Coordinates = coordinates
		} else {
			coordinates, err := s.obtainCoordinates(p)
			if err != nil {
				return err
			}
			s.coordinates[p.ID], p.Coordinates = coordinates, coordinates
		}
	}

	cache, err := json.Marshal(res)
	if err != nil {
		return err
	}
	var timeDB time.Time
	s.dbMutex.Lock()
	if s.DB() != nil && s.updated.IsZero() {
		row := s.DB().QueryRow("SELECT time FROM spots ORDER BY time DESC LIMIT 1;")
		s.dbMutex.Unlock()
		var updated string
		err = row.Scan(&updated)
		if err != sql.ErrNoRows {
			if err != nil {
				return err
			}
			timeDB, err = time.Parse(timeLayout, updated)
			if err != nil {
				return err
			}
		}
	} else {
		s.dbMutex.Unlock()
	}

	s.updated, s.parkings, s.cache = res.Updated, res.Parkings, cache
	s.dbMutex.Lock()
	defer s.dbMutex.Unlock()
	if s.DB() != nil && timeDB.IsZero() || s.updated.After(timeDB) {
		for i := 0; i < len(res.Parkings); i++ {
			if _, err := s.insertSpotsStmt.Exec(res.Parkings[i].ID,
				res.Updated.Format(timeLayout), res.Parkings[i].Spots); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Server) DB() *sql.DB {
	return s.db
}

func (s *Server) SetDB(db *sql.DB) error {
	closeStatements := func() {
		if s.DB() != nil {
			s.insertCoordinatesStmt.Close()
			s.insertSpotsStmt.Close()
		}
	}
	s.dbMutex.Lock()
	defer s.dbMutex.Unlock()
	if db == nil {
		closeStatements()
		s.db = nil
		return nil
	}
	insertCoordinatesStmt, err := db.Prepare("INSERT INTO coordinates (parking_id, latitude, longitude) VALUES (?, ?, ?);")
	if err != nil {
		return err
	}
	insertSpotsStmt, err := db.Prepare("INSERT INTO spots (parking_id, time, free) VALUES (?, ?, ?);")
	if err != nil {
		return err
	}
	closeStatements()
	s.db, s.insertCoordinatesStmt, s.insertSpotsStmt = db, insertCoordinatesStmt, insertSpotsStmt
	return nil
}

func (s *Server) ScheduleScraping(interval time.Duration) {
	if interval == 0 {
		if s.ticker != nil {
			s.ticker.Stop()
			s.ticker = nil

			s.done <- struct{}{}
			s.done = nil
		}
		return
	}
	if s.ticker != nil {
		s.ticker.Reset(interval)
		return
	}

	s.ticker = time.NewTicker(interval)
	s.done = make(chan struct{})
	go func() {
		for {
			select {
			case <-s.ticker.C:
				go func() {
					err := s.scrape()
					if err != nil {
						s.logln("scraping:", err)
					}
				}()
			case <-s.done:
				return
			}
		}
	}()
}

func (s *Server) Close() error {
	s.ScheduleScraping(0)
	return s.Server.Close()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.ScheduleScraping(0)
	return s.Server.Shutdown(ctx)
}

func NewServer(httpServer *http.Server, scraper *scraping.Scraper, scrapingInterval time.Duration, presets map[int]parken.Coordinates, client *nominatim.Client, db *sql.DB, logger *log.Logger) (*Server, error) {
	if httpServer == nil {
		httpServer = &http.Server{}
	}
	mux := http.NewServeMux()
	httpServer.Handler = mux
	if client == nil {
		client = &nominatim.Client{}
	}
	server := &Server{Server: httpServer, Scraper: scraper, coordinates: make(map[int]parken.Coordinates), presets: presets, Client: client, Logger: logger}

	if db != nil {
		if err := server.SetDB(db); err != nil {
			return nil, err
		}
		if err := server.queryCoordinates(); err != nil {
			return nil, err
		}
	}

	if err := server.scrape(); err != nil {
		return nil, fmt.Errorf("scraping: %w", err)
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			httpError(w, http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, "frontend/index.html")
	})
	mux.HandleFunc("/api/parkings", server.parkingsHandler)
	// dirty
	mime.AddExtensionType(".ttf", "font/ttf")
	mux.Handle("/static/", http.StripPrefix("/static/", compressedFileServer(http.Dir("frontend"), []string{".html", ".css", ".js", ".ttf"})))
	mux.Handle("/tiles/", http.StripPrefix("/tiles/", http.FileServer(http.Dir("tiles"))))

	server.ScheduleScraping(scrapingInterval)

	return server, nil
}
