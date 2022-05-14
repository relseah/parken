package parken

import (
	"encoding/json"
	"net/url"
)

type URL struct {
	*url.URL
}

func (u URL) MarshalJSON() ([]byte, error) {
	if u.URL == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(u.String())
}

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
	Website        URL         `json:"website"`
	Email          string      `json:"email"`
	Prices         string      `json:"prices"`
	LongTermPrices string      `json:"longTermPrices"`
	OpeningHours   string      `json:"openingHours"`
	OpenAllDay     bool        `json:"openAllDay"`
	Spots          int         `json:"available"`
	Capacity       int         `json:"capacity"`
}
