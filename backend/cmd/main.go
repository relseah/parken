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
	"strings"
	"time"

	"github.com/relseah/parken"
	"github.com/relseah/parken/nominatim"
	"github.com/relseah/parken/scraping"
	"github.com/relseah/parken/web"
)

func openDB(config *config) (*sql.DB, error) {
	db, err := sql.Open("mysql", config.Database.DataSourceName)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	defer func() {
		if err != nil {
			db.Close()
		}
	}()
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}
	rows, err := db.Query("SELECT 1 FROM information_schema.schemata WHERE schema_name = 'parken';")
	if err != nil {
		return nil, err
	}
	initialized := rows.Next()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		initialized = false
	}
	if !initialized {
		if err = initializeDB(db); err != nil {
			return nil, err
		}
	} else {
		if err := selectDefaultDatabase(db); err != nil {
			return nil, err
		}
	}
	return db, nil
}

func selectDefaultDatabase(db *sql.DB) error {
	if _, err := db.Exec("USE parken"); err != nil {
		return fmt.Errorf("selecting default database: %w", err)
	}
	return nil
}

func initializeDB(db *sql.DB) error {
	log.Println("Initializing database...")
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
	return nil
}

func queryCoordinates(db *sql.DB) (map[int]parken.Coordinates, error) {
	rows, err := db.Query("SELECT parking_id, latitude, longitude FROM coordinates;")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	coordinates := make(map[int]parken.Coordinates)
	var id int
	var current parken.Coordinates
	for rows.Next() {
		rows.Scan(&id, &current.Latitude, &current.Longitude)
		coordinates[id] = current
	}
	return coordinates, rows.Err()
}

func obtainCoordinates(config *config, scraper *scraping.Scraper, db *sql.DB) (map[int]parken.Coordinates, error) {
	log.Println("Obtaining coordinates...")
	res, err := scraper.Scrape(time.Time{})
	if err != nil {
		return nil, err
	}
	parkings := res.Parkings
	// Consider issuing only one query.
	insertStmt, err := db.Prepare("INSERT INTO coordinates (parking_id, latitude, longitude) VALUES (?, ?, ?);")
	if err != nil {
		return nil, err
	}
	defer insertStmt.Close()
	coordinatesDB, err := queryCoordinates(db)
	if err != nil {
		return nil, err
	}
	client := nominatim.NewClient(config.Coordinates.Nominatim.RateLimiting.Rate, time.Duration(config.Coordinates.Nominatim.RateLimiting.Interval))
	coordinates := make(map[int]parken.Coordinates)
	ambiguous := false
	for i := 0; i < len(parkings); i++ {
		p := &parkings[i]
		if preset, ok := config.Coordinates.Presets[p.ID]; ok {
			coordinates[p.ID] = preset
		} else {
			if c, ok := coordinatesDB[p.ID]; ok {
				coordinates[p.ID] = c
			} else {
				results, err := client.Search(p)
				if err != nil {
					return coordinates, fmt.Errorf("searching for coordinates of parking with ID %d: %w", p.ID, err)
				}
				if len(results) == 0 {
					ambiguous = true
					log.Printf("No results for parking P%d %s.\n", p.ID, p.Name)
				} else if len(results) > 1 {
					ambiguous = true
					var b strings.Builder
					fmt.Fprintf(&b, "Multiple results for parking P%d %s.\n", p.ID, p.Name)
					for i, c := range results {
						fmt.Fprintf(&b, "%d. Latitude: %f°, longitude: %f°\n", i, c.Latitude, c.Longitude)
					}
					log.Print(b.String())
				} else {
					coordinates[p.ID] = results[0]
					_, err = insertStmt.Exec(p.ID, coordinates[p.ID].Latitude, coordinates[p.ID].Longitude)
					if err != nil {
						return coordinates, err
					}
				}
			}
		}
	}
	if ambiguous {
		log.Println("Please specify missing or ambiguous coordinates in the configuration.")
		return nil, nil
	}
	return coordinates, nil
}

func runServer(config *config) error {
	interrupted := false
	close := func(c io.Closer) {
		if !interrupted {
			c.Close()
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

	db, err := openDB(config)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer close(db)

	coordinates, err := obtainCoordinates(config, scraper, db)
	if err != nil {
		return fmt.Errorf("obtaining coordinates: %w", err)
	}
	if coordinates == nil {
		return nil
	}

	log.Println("Initializing server...")
	server, err := web.NewServer(httpServer, scraper, time.Duration(config.Scraping.Interval), coordinates, db, log.Default())
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
	flag.Parse()

	config, err := readConfig(configPath)
	if err != nil {
		log.Fatalln("reading configuration:", err)
	}

	if err = runServer(config); err != nil {
		log.Fatalln(err)
	}
}
