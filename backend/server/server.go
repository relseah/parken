package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"github.com/relseah/parken"
	"github.com/relseah/parken/scraping"
)

type Server struct {
	*http.Server
	Scraper *scraping.Scraper
	DB      *sql.DB
	Logger  *log.Logger

	parkings []parken.Parking

	ticker *time.Ticker
	done   chan struct{}
}

func (s *Server) logln(v ...any) {
	if s.Logger != nil {
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
		s.logln("parkings handler: ", err)
		return
	}
	w.Write(payload)
}

func (s *Server) scrape() error {
	parkings, err := s.Scraper.Scrape()
	if err != nil {
		return err
	}
	s.parkings = parkings
	return nil
}

func (s *Server) ScheduleScraping(intervall time.Duration) {
	if intervall == 0 {
		if s.ticker != nil {
			s.ticker.Stop()
			s.ticker = nil

			s.done <- struct{}{}
			s.done = nil
		}
		return
	}
	if s.ticker != nil {
		s.ticker.Reset(intervall)
		return
	}

	s.ticker = time.NewTicker(intervall)
	s.done = make(chan struct{})
	go func() {
		for {
			select {
			case <-s.ticker.C:
				err := s.scrape()
				if err != nil {
					s.logln("scraping: ", err)
				}
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

func NewServer(httpServer *http.Server, scraper *scraping.Scraper, intervall time.Duration, logger *log.Logger, db *sql.DB) (*Server, error) {
	if httpServer == nil {
		httpServer = &http.Server{}
	}
	mux := http.NewServeMux()
	httpServer.Handler = mux
	server := &Server{Server: httpServer, Scraper: scraper, Logger: logger, DB: db}

	err := server.scrape()
	if err != nil {
		return nil, err
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

	server.ScheduleScraping(intervall)

	return server, nil
}
