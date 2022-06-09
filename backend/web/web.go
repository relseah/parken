package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
	Scraper         *scraping.Scraper
	NominatimClient *nominatim.Client
	Logger          *log.Logger

	cache []byte

	parkings []parken.Parking
	updated  time.Time

	db                    *sql.DB
	dbMutex               sync.Mutex
	insertSpotsStmt       *sql.Stmt
	insertCoordinatesStmt *sql.Stmt

	ticker *time.Ticker
	done   chan struct{}
}

func (s *Server) logln(v ...any) {
	logger := s.Logger
	if logger != nil {
		s.Logger.Println(v...)
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
	w.Write(s.cache)
}

func (s *Server) selectCoordinates() (map[int]parken.Coordinates, error) {
	s.dbMutex.Lock()
	if s.db != nil {
		rows, err := s.db.Query("SELECT parking_id, latitude, longitude FROM coordinates;")
		s.dbMutex.Unlock()
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		res := make(map[int]parken.Coordinates)
		var id int
		var coordinates parken.Coordinates
		for rows.Next() {
			rows.Scan(&id, &coordinates.Latitude, &coordinates.Longitude)
			res[id] = coordinates
		}
		return res, rows.Err()
	}
	s.dbMutex.Unlock()
	return nil, nil
}

var coordinatesExceptions = map[int]parken.Coordinates{
	25: {Latitude: 49.418786850000004, Longitude: 8.675542779096492},
}

func (s *Server) scrape() error {
	res, err := s.Scraper.Scrape(s.updated)
	if err != nil {
		if err == scraping.ErrNoUpdate {
			return nil
		}
		return err
	}
	var coordinatesDB map[int]parken.Coordinates
	var queried bool
	for i := 0; i < len(res.Parkings); i++ {
		parking := &res.Parkings[i]

		var ok bool
		if parking.Coordinates, ok = coordinatesExceptions[parking.ID]; ok {
			continue
		}

		if s.parkings != nil {
			if s.parkings[i].ID == parking.ID {
				parking.Coordinates = s.parkings[i].Coordinates
				continue
			}
			found := false
			for j := 0; j < len(s.parkings); j++ {
				if s.parkings[j].ID == parking.ID {
					parking.ID = s.parkings[j].ID
					found = true
					break
				}
			}
			if found {
				continue
			}
		}

		if !queried {
			coordinatesDB, err = s.selectCoordinates()
			if err != nil {
				return err
			}
			queried = true
		}
		if parking.Coordinates, ok = coordinatesDB[parking.ID]; ok {
			continue
		}

		parking.Coordinates, err = s.NominatimClient.FetchCoordinates(parking.ID)
		if err != nil {
			return fmt.Errorf("fetching coordinates for parking with ID %d: %w", parking.ID, err)
		}

		s.dbMutex.Lock()
		if s.db != nil {
			_, err = s.insertCoordinatesStmt.Exec(parking.ID, parking.Coordinates.Latitude, parking.Coordinates.Longitude)
			s.dbMutex.Unlock()
			if err != nil {
				return err
			}
		} else {
			s.dbMutex.Unlock()
		}
	}

	cache, err := json.Marshal(res.Parkings)
	if err != nil {
		return err
	}
	var timeDB time.Time
	s.dbMutex.Lock()
	if s.db != nil && s.updated.IsZero() {
		row := s.db.QueryRow("SELECT time FROM spots ORDER BY time DESC LIMIT 1;")
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
	if s.db != nil && !timeDB.IsZero() && s.updated.After(timeDB) {
		for _, p := range res.Parkings {
			if _, err := s.insertSpotsStmt.Exec(p.ID,
				res.Updated.Format(timeLayout), p.Spots); err != nil {
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
		if s.db != nil {
			s.insertSpotsStmt.Close()
			s.insertCoordinatesStmt.Close()
		}
	}
	s.dbMutex.Lock()
	defer s.dbMutex.Unlock()
	if db == nil {
		closeStatements()
		s.db = nil
		return nil
	}
	insertSpotsStmt, err := db.Prepare("INSERT INTO spots (parking_id, time, free) VALUES (?, ?, ?);")
	if err != nil {
		return err
	}
	insertCoordinatesStmt, err := db.Prepare("INSERT INTO coordinates (parking_id, latitude, longitude) VALUES (?, ?, ?);")
	if err != nil {
		return err
	}
	closeStatements()
	s.db, s.insertSpotsStmt, s.insertCoordinatesStmt = db, insertSpotsStmt, insertCoordinatesStmt
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

func NewServer(httpServer *http.Server, scraper *scraping.Scraper, interval time.Duration, nominatimClient *nominatim.Client, logger *log.Logger, db *sql.DB) (*Server, error) {
	if httpServer == nil {
		httpServer = &http.Server{}
	}
	mux := http.NewServeMux()
	httpServer.Handler = mux
	server := &Server{Server: httpServer, Scraper: scraper, NominatimClient: nominatimClient, Logger: logger}

	if db != nil {
		if err := server.SetDB(db); err != nil {
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
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend"))))

	server.ScheduleScraping(interval)

	return server, nil
}
