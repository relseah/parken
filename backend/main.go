package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Address struct {
	Street      string `json:"street"`
	HouseNumber int    `json:"houseNumber"`
	Town        string `json:"town"`
	PostalCode  int    `json:"postalCode"`
}

type ParkingLot struct {
	Name     string  `json:"name"`
	Address  Address `json:"address"`
	Capacity int     `json:"capacity"`
}

var mock = []ParkingLot{
	{
		Name: "P0 Am Bismarckplatz",
		Address: Address{
			Street:      "Schneidenmühlstraße",
			HouseNumber: 5,
			Town:        "Heidelberg",
			PostalCode:  69115,
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
		Capacity: 528,
	},
}

func parkingLots(w http.ResponseWriter, r *http.Request) {
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
	http.HandleFunc("/api", parkingLots)
	http.Handle("/", http.FileServer(http.Dir("frontend")))
	log.Fatalln(http.ListenAndServe(":8000", nil))
}
