package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/relseah/parken/scraping"
	"github.com/relseah/parken/server"
)

func initializeDatabase(config *config) {

}

func startServer(config *config) {
	var db *sql.DB
	if config.Database.DataSourceName != "" {
		var err error
		db, err = sql.Open("mysql", config.Database.DataSourceName)
		if err != nil {
			log.Fatalln("opening database:", err)
		}
		if err = db.Ping(); err != nil {
			log.Fatalln("establishing database connection:", err)
		}
	}
	httpServer := &http.Server{Addr: config.Server.Address}
	server, err := server.NewServer(httpServer, &scraping.Scraper{}, 5*time.Second, log.Default(), db)
	if err != nil {
		log.Fatalln("initializing server:", err)
	}
	log.Fatalln(server.ListenAndServe())
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "configuration", "config.json", "path to configuration")
	var initialize bool
	flag.BoolVar(&initialize, "initialize", false, "initialize database")
	flag.Parse()

	config, err := readConfig(configPath)
	if err != nil {
		log.Fatalln("reading configuration:", err)
	}

	if initialize {
		initializeDatabase(config)
	} else {
		startServer(config)
	}
}
