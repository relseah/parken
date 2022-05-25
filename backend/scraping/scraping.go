package scraping

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/relseah/parken"
)

type RawParking struct {
	ID             string `json:"uid"`
	Name           string
	Operator       string `json:"management"`
	Address        string
	PhoneNumber    string `json:"phone"`
	Website        string
	Email          string
	Prices         string `json:"shortterm_parker"`
	LongTermPrices string `json:"longterm_parker"`
	OpeningHours   string `json:"opening_hours"`
	OpenAllDay     bool   `json:"all_day"`
	Status         struct {
		// General  string `json:"status"`
		Spots    int `json:"current"`
		Capacity int `json:"total"`
	} `json:"parkingupdate"`
}

type Result struct {
	Updated  time.Time
	Parkings []parken.Parking
}

var ErrAddressFormat = errors.New("invalid address format")

func ParseAddress(rawAddress string) (parken.Address, error) {
	lines := strings.Split(rawAddress, ",")
	if len(lines) != 2 {
		return parken.Address{}, ErrAddressFormat
	}
	for i := 0; i < len(lines); i++ {
		lines[i] = strings.TrimSpace(lines[i])
	}

	i := strings.LastIndex(lines[0], " ")
	var street string
	if i == -1 {
		street = lines[0]
	} else {
		street = lines[0][:i]
	}
	address := parken.Address{Street: street}
	if i != -1 {
		address.HouseNumber = lines[0][i+1:]
	}

	i = strings.Index(lines[1], " ")
	if i == -1 {
		return parken.Address{}, ErrAddressFormat
	}
	address.Town = lines[1][i+1:]
	postalCode, err := strconv.Atoi(lines[1][:i])
	if err != nil {
		return address, fmt.Errorf("converting postal code: %w", err)
	}
	address.PostalCode = postalCode
	return address, nil
}

type Scraper struct {
	Client *http.Client
}

func (s *Scraper) client() *http.Client {
	if s.Client != nil {
		return s.Client
	}
	return http.DefaultClient
}

var coordinates = map[int]parken.Coordinates{
	0: {
		Latitude:  49.4096239,
		Longitude: 8.691726098614684,
	},
	1: {
		Latitude:  49.40774055,
		Longitude: 8.69046050266848,
	},
}

func (s *Scraper) ScrapeSpots() (map[int]int, error) {
	return nil, nil
}

var ErrAPI = errors.New("returned status does not indicate success")
var ErrNoUpdate = errors.New("no more recent data available")

func (s *Scraper) Scrape(updated time.Time) (Result, error) {
	type payload struct {
		Status string
		Data   struct {
			Updated  string
			Parkings []RawParking `json:"parkinglocations"`
		}
	}
	file, err := os.Open("payload.json")
	if err != nil {
		return Result{}, err
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	pl := &payload{}
	err = dec.Decode(pl)
	if err != nil {
		return Result{}, err
	}
	err = file.Close()
	if err != nil {
		return Result{}, err
	}
	if pl.Status != "success" {
		return Result{}, ErrAPI
	}
	t, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", pl.Data.Updated)
	if err != nil {
		return Result{}, err
	}
	t = t.UTC()
	if t.Equal(updated) || t.Before(updated) {
		return Result{}, ErrNoUpdate
	}

	res := Result{Updated: t, Parkings: make([]parken.Parking, 0, len(pl.Data.Parkings))}
	for _, raw := range pl.Data.Parkings {
		id, err := strconv.Atoi(raw.ID)
		if err != nil {
			return res, fmt.Errorf("parsing ID: %w", err)
		}
		address, err := ParseAddress(raw.Address)
		if err != nil {
			return res, fmt.Errorf("parsing address %s: %w", raw.Address, err)
		}
		var website parken.URL
		if raw.Website != "" {
			u, err := url.Parse(raw.Website)
			website = parken.URL{URL: u}
			if err != nil {
				return res, fmt.Errorf("parsing website’s URL: %w", err)
			}
		}
		p := parken.Parking{
			ID:             id,
			Name:           raw.Name,
			Operator:       raw.Operator,
			Address:        address,
			Coordinates:    coordinates[id],
			PhoneNumber:    raw.PhoneNumber,
			Website:        website,
			Email:          raw.Email,
			Prices:         raw.Prices,
			LongTermPrices: raw.LongTermPrices,
			OpeningHours:   raw.OpeningHours,
			OpenAllDay:     raw.OpenAllDay,
			Spots:          raw.Status.Spots,
			Capacity:       raw.Status.Capacity,
		}
		res.Parkings = append(res.Parkings, p)
	}
	return res, nil
}
