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

func openDatabase(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return db, err
	}
	return db, db.Ping()
}

func initializeDatabase(config *config) {
	if config.Database.DataSourceName == "" {
		log.Fatalln("no data source name specified")
	}
	db, err := openDatabase(config.Database.DataSourceName)
	if err != nil {
		log.Fatalln("opening database:", err)
	}
	_, err = db.Exec("CREATE DATABASE parken")
	if err != nil {
		log.Fatalln("creating database:", err)
	}
	_, err = db.Exec(`CREATE TABLE parken.status (
parking_id INT NOT NULL,
time DATETIME NOT NULL,
spots INT,
PRIMARY KEY (parking_id, time))`)
	if err != nil {
		log.Fatalln("creating table:", err)
	}
}

func startServer(config *config) {
	var db *sql.DB
	if config.Database.DataSourceName != "" {
		var err error
		db, err = openDatabase(config.Database.DataSourceName)
		if err != nil {
			log.Fatalln("opening database:", err)
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
