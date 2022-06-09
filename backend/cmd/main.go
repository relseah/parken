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

	"github.com/relseah/parken/nominatim"
	"github.com/relseah/parken/scraping"
	"github.com/relseah/parken/web"
)

func selectDefaultDatabase(db *sql.DB) error {
	if _, err := db.Exec("USE parken"); err != nil {
		return fmt.Errorf("selecting default database: %w", err)
	}
	return nil
}

func initializeDatabase(db *sql.DB) error {
	if _, err := db.Exec("CREATE DATABASE parken;"); err != nil {
		return fmt.Errorf("creating database: %w", err)
	}
	if err := selectDefaultDatabase(db); err != nil {
		return err
	}
	query := `CREATE TABLE spots (
parking_id INT NOT NULL,
time DATETIME NOT NULL,
free INT,
PRIMARY KEY (parking_id, time));`
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("creating spots table: %w", err)
	}
	query = `CREATE TABLE coordinates (
parking_id INT NOT NULL,
latitude DOUBLE NOT NULL,
longitude DOUBLE NOT NULL,
PRIMARY KEY (parking_id));`
	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("creating coordinates table: %w", err)
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
		db, err = sql.Open("mysql", config.Database.DataSourceName)
		if err != nil {
			return fmt.Errorf("opening database: %w", err)
		}
		defer close(db)
		err = db.Ping()
		if err != nil {
			return fmt.Errorf("pinging database: %w", err)
		}
		rows, err := db.Query("SELECT 1 FROM information_schema.schemata WHERE schema_name = 'parken';")
		if err != nil {
			return err
		}
		initialized := rows.Next()
		if err := rows.Err(); err != nil {
			return err
		}
		if err != nil {
			if err != sql.ErrNoRows {
				return err
			}
			initialized = false
		}
		if !initialized {
			if err = initializeDatabase(db); err != nil {
				return err
			}
		} else {
			if err := selectDefaultDatabase(db); err != nil {
				return err
			}
		}
	}
	var addr string
	if config.Web.Address == "" {
		addr = ":80"
	} else {
		addr = config.Web.Address
	}
	httpServer := &http.Server{Addr: addr}
	scraper := new(scraping.Scraper)
	nominatimClient := new(nominatim.Client)
	nominatimClient.SetRate(config.Scraping.Nominatim.RateLimiting.Rate, time.Duration(config.Scraping.Nominatim.RateLimiting.Interval))

	log.Println("Initializing server...")
	server, err := web.NewServer(httpServer, scraper, time.Duration(config.Scraping.Interval), nominatimClient, log.Default(), db)
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
		err = server.SetDB(nil)
		if err != nil {
			log.Println(err)
		}
		err = db.Close()
		if err != nil {
			log.Println(err)
		}
		log.Println("Shutting down server...")
		err = server.Shutdown(context.Background())
		if err != nil {
			log.Println(err)
		}
		return nil
	}
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "configuration", "config.json", "path to configuration")

	config, err := readConfig(configPath)
	if err != nil {
		log.Fatalln("reading configuration:", err)
	}

	err = runServer(config)
	if err != nil {
		log.Fatalln(err)
	}
}
