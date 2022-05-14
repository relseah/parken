package main

import (
	"encoding/json"
	"os"
)

type config struct {
	Server struct {
		Address string
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
