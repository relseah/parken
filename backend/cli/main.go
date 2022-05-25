package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
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

func initializeDatabase(db *sql.DB) error {
	if _, err := db.Exec("CREATE DATABASE parken"); err != nil {
		return fmt.Errorf("creating database: %w", err)
	}
	_, err := db.Exec(`CREATE TABLE parken.spots (
parking_id INT NOT NULL,
time DATETIME NOT NULL,
free INT,
PRIMARY KEY (parking_id, time))`)
	if err != nil {
		return fmt.Errorf("creating table: %w", err)
	}
	log.Println("Database initialized.")
	return nil
}

func runServer(config *config) error {
	var interrupted bool
	close := func(c io.Closer) {
		if !interrupted {
			c.Close()
		}
	}
	var db *sql.DB
	if config.Database.DataSourceName != "" {
		var err error
		db, err = openDatabase(config.Database.DataSourceName)
		if err != nil {
			return fmt.Errorf("opening database connection: %w", err)
		}
		defer close(db)
		rows, err := db.Query("SELECT 1 FROM information_schema.schemata WHERE schema_name = 'parken'")
		if err != nil {
			return err
		}
		initialized := rows.Next()
		rows.Close()
		if !initialized {
			initializeDatabase(db)
		}
		if _, err = db.Exec("USE parken"); err != nil {
			return err
		}
	}
	var addr string
	if config.Server.Address == "" {
		addr = ":80"
	} else {
		addr = config.Server.Address
	}
	httpServer := &http.Server{Addr: addr}
	server, err := server.NewServer(httpServer, &scraping.Scraper{}, time.Duration(config.Scraping.Interval), log.Default(), db)
	if err != nil {
		return fmt.Errorf("initializing server: %w", err)
	}
	defer close(server)

	e := make(chan error)
	go func() {
		e <- server.ListenAndServe()
	}()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	log.Println("Server running.")

	select {
	case err = <-e:
		return err
	case <-interrupt:
		interrupted = true
		log.Println("Closing database connection...")
		server.SetDB(nil)
		db.Close()
		log.Println("Shutting down server...")
		server.Shutdown(context.Background())
		return nil
	}
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "configuration", "config.json", "path to configuration")
	flag.Parse()

	config, err := readConfig(configPath)
	if err != nil {
		log.Fatalln("reading configuration:", err)
	}

	err = runServer(config)
	if err != nil {
		log.Fatalln(err)
	}
}
