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
	"github.com/relseah/parken/scraping"
)

const timeLayout = "2006-01-02 15:04:05"

type Server struct {
	*http.Server
	Scraper *scraping.Scraper
	Logger  *log.Logger

	cache []byte

	parkings    []parken.Parking
	updated     time.Time
	coordinates map[int]parken.Coordinates

	db              *sql.DB
	dbMutex         sync.Mutex
	insertSpotsStmt *sql.Stmt

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

func (s *Server) scrape() error {
	res, err := s.Scraper.Scrape(s.updated)
	if err != nil {
		if err == scraping.ErrNoUpdate {
			return nil
		}
		return err
	}
	var ok bool
	for i := 0; i < len(res.Parkings); i++ {
		res.Parkings[i].Coordinates, ok = s.coordinates[res.Parkings[i].ID]
		if !ok {
			return fmt.Errorf("missing coordinates for parking with ID %d", res.Parkings[i].ID)
		}
	}

	cache, err := json.Marshal(res.Parkings)
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
	if s.DB() != nil && !timeDB.IsZero() && s.updated.After(timeDB) {
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
	closeStmt := func() {
		if s.DB() != nil {
			s.insertSpotsStmt.Close()
		}
	}
	s.dbMutex.Lock()
	defer s.dbMutex.Unlock()
	if db == nil {
		closeStmt()
		s.db = nil
		return nil
	}
	insertSpotsStmt, err := db.Prepare("INSERT INTO spots (parking_id, time, free) VALUES (?, ?, ?);")
	if err != nil {
		return err
	}
	closeStmt()
	s.db, s.insertSpotsStmt = db, insertSpotsStmt
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

func NewServer(httpServer *http.Server, scraper *scraping.Scraper, interval time.Duration, coordinates map[int]parken.Coordinates, db *sql.DB, logger *log.Logger) (*Server, error) {
	if httpServer == nil {
		httpServer = &http.Server{}
	}
	mux := http.NewServeMux()
	httpServer.Handler = mux
	server := &Server{Server: httpServer, Scraper: scraper, Logger: logger, coordinates: coordinates}

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
