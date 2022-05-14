package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/relseah/parken/scraping"
	"github.com/relseah/parken/server"
)

func initializeDatabase() {

}

func startServer() {
	httpServer := &http.Server{Addr: ":8000"}
	server, err := server.NewServer(httpServer, &scraping.Scraper{}, 5*time.Second, log.Default(), nil)
	if err != nil {
		log.Fatalln("initializing server: ", err)
	}
	log.Fatalln(server.ListenAndServe())
}

func main() {
	var initialize bool
	flag.BoolVar(&initialize, "initialize", false, "initialize database")
	flag.Parse()
	if initialize {
		initializeDatabase()
	} else {
		startServer()
	}
}
