package parken

import "net/url"

type Coordinates struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Address struct {
	Street      string `json:"street"`
	HouseNumber string `json:"houseNumber,omitempty"`
	Town        string `json:"town"`
	PostalCode  int    `json:"postalCode"`
}

type Parking struct {
	ID             int         `json:"id"`
	Name           string      `json:"name"`
	Operator       string      `json:"operator"`
	Address        Address     `json:"address"`
	Coordinates    Coordinates `json:"coordinates"`
	PhoneNumber    string      `json:"phoneNumber"`
	Website        *url.URL    `json:"website"`
	Email          string      `json:"email"`
	Prices         string      `json:"prices"`
	LongTermPrices string      `json:"longTermPrices"`
	OpeningHours   string      `json:"openingHours"`
	OpenAllDay     bool        `json:"openAllDay"`
	Spots          int         `json:"available"`
	Capacity       int         `json:"capacity"`
}
