package main

import (
	"encoding/json"
	"os"
	"time"
)

type duration time.Duration

func (d *duration) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	val, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = duration(val)
	return nil
}

type config struct {
	Web struct {
		Address string
	}
	Scraping struct {
		Interval  duration
		Nominatim struct {
			RateLimiting struct {
				Rate     int
				Interval duration
			}
		}
	}
	Database struct {
		DataSourceName string
	}
}

func readConfig(path string) (*config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(file)
	config := &config{}
	return config, dec.Decode(config)
}
