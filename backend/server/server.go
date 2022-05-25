package server

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

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

type Server struct {
	*http.Server
	Scraper *scraping.Scraper
	Logger  *log.Logger

	parkings []parken.Parking
	updated  time.Time

	db               *sql.DB
	dbMutex          sync.Mutex
	insertStatusStmt *sql.Stmt

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
	payload, err := json.Marshal(s.parkings)
	if err != nil {
		httpError(w, http.StatusInternalServerError)
		s.logln("parkings handler:", err)
		return
	}
	w.Write(payload)
}

func (s *Server) scrape() error {
	res, err := s.Scraper.Scrape(s.updated)
	if err != nil {
		if err == scraping.ErrNoUpdate {
			return nil
		}
		return err
	}
	s.updated, s.parkings = res.Updated, res.Parkings

	s.dbMutex.Lock()
	if s.db != nil {
		defer s.dbMutex.Unlock()
		rows, err := s.db.Query("SELECT 1 FROM spots WHERE time = ?", formatTime(res.Updated))
		if err != nil {
			return err
		}
		if rows.Next() {
			return nil
		}
		rows.Close()
		for _, p := range res.Parkings {
			_, err := s.insertStatusStmt.Exec(p.ID,
				formatTime(res.Updated), p.Spots)
			if err != nil {
				return err
			}
		}
	} else {
		s.dbMutex.Unlock()
	}
	return nil
}

func (s *Server) DB() *sql.DB {
	return s.db
}

func (s *Server) SetDB(db *sql.DB) error {
	clostStmt := func() {
		if s.insertStatusStmt != nil {
			s.insertStatusStmt.Close()
		}
	}
	if s.db != nil {
		s.dbMutex.Lock()
		defer s.dbMutex.Unlock()
	}
	if db == nil {
		clostStmt()
		s.db = nil
		return nil
	}
	stmt, err := db.Prepare("INSERT INTO spots (parking_id, time, free) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	clostStmt()
	s.insertStatusStmt = stmt
	s.db = db
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

func NewServer(httpServer *http.Server, scraper *scraping.Scraper, interval time.Duration, logger *log.Logger, db *sql.DB) (*Server, error) {
	if httpServer == nil {
		httpServer = &http.Server{}
	}
	mux := http.NewServeMux()
	httpServer.Handler = mux
	server := &Server{Server: httpServer, Scraper: scraper, Logger: logger}

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
