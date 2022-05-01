package main

import (
	"log"
	"net/http"

	"github.com/relseah/parken/scraping"
	"github.com/relseah/parken/server"
)

func main() {
	scraper := &scraping.Scraper{}
	parkings, err := scraper.Scrape()
	if err != nil {
		log.Fatal("scraping parkings: ", err)
	}
	httpServer := &http.Server{Addr: ":8000"}
	server := server.NewServer(httpServer, parkings)
	log.Fatal(server.ListenAndServe())
}
