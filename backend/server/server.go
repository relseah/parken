package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/relseah/parken"
)

type Server struct {
	Parkings []parken.Parking
	*http.Server
}

var errorMessages = map[int]string{
	http.StatusNotFound:            "Not Found",
	http.StatusInternalServerError: "Internal Server Error",
}

func httpError(w http.ResponseWriter, code int) {
	http.Error(w, fmt.Sprintf("%d %s", code, errorMessages[code]), code)
}

func (s *Server) parkings(w http.ResponseWriter, r *http.Request) {
	payload, err := json.Marshal(s.Parkings)
	if err != nil {
		httpError(w, http.StatusInternalServerError)
		log.Println(err)
		return
	}
	_, err = w.Write(payload)
	if err != nil {
		log.Println(err)
	}
}

func NewServer(httpServer *http.Server, parkings []parken.Parking) *Server {
	if httpServer == nil {
		httpServer = &http.Server{}
	}
	mux := http.NewServeMux()
	httpServer.Handler = mux
	server := &Server{Parkings: parkings, Server: httpServer}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			httpError(w, http.StatusNotFound)
			return
		}
		http.ServeFile(w, r, "frontend/index.html")
	})
	mux.HandleFunc("/api", server.parkings)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("frontend"))))
	return server
}
