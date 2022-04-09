package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Address struct {
	Street      string `json:"street"`
	HouseNumber int    `json:"houseNumber"`
	Town        string `json:"town"`
	PostalCode  int    `json:"postalCode"`
}

type Parking struct {
	Name        string      `json:"name"`
	Address     Address     `json:"address"`
	Coordinates Coordinates `json:"coordinates"`
	Capacity    int         `json:"capacity"`
}

var mock = []Parking{
	{
		Name: "P0 Am Bismarckplatz",
		Address: Address{
			Street:      "Schneidmühlstraße",
			HouseNumber: 5,
			Town:        "Heidelberg",
			PostalCode:  69115,
		},
		Coordinates: Coordinates{
			Latitude:  49.4096239,
			Longitude: 8.691726098614684,
		},
		Capacity: 38,
	},
	{
		Name: "P1 Poststraße",
		Address: Address{
			Street:      "Postraße",
			HouseNumber: 0,
			Town:        "Heidelberg",
			PostalCode:  69115,
		},
		Coordinates: Coordinates{
			Latitude:  49.40774055,
			Longitude: 8.69046050266848,
		},
		Capacity: 528,
	},
}

func parkings(w http.ResponseWriter, r *http.Request) {
	payload, err := json.Marshal(mock)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	_, err = w.Write(payload)
	if err != nil {
		log.Println(err)
	}
}

func main() {
	http.HandleFunc("/api", parkings)
	http.Handle("/", http.FileServer(http.Dir("frontend")))
	log.Fatalln(http.ListenAndServe(":8000", nil))
}
