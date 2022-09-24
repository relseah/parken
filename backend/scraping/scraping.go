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

type rawParking struct {
	ID               string `json:"uid"`
	Name             string
	Closed           bool   `json:"is_closed"`
	Operator         string `json:"management"`
	Address          string
	PhoneNumber      string `json:"phone"`
	Website          string
	Email            string
	Prices           string `json:"shortterm_parker"`
	LongTermPrices   string `json:"longterm_parker"`
	OpeningHours     string `json:"opening_hours"`
	OpenAllDay       bool   `json:"all_day"`
	ChargingStations string `json:"e_charge_station"`
	Zone             struct {
		ID, Name string
	} `json:"parkingzone"`
	Status struct {
		// General  string `json:"status"`
		Spots    int `json:"current"`
		Capacity int `json:"total"`
	} `json:"parkingupdate"`
}

type Result struct {
	Updated  time.Time        `json:"updated"`
	Zones    map[int]string   `json:"zones"`
	Parkings []parken.Parking `json:"parkings"`
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
	client := s.Client
	if client != nil {
		return client
	}
	return http.DefaultClient
}

var ErrAPI = errors.New("returned status does not indicate success")
var ErrNoUpdate = errors.New("no more recent data available")

func (s *Scraper) Scrape(updated time.Time) (Result, error) {
	type body struct {
		Status string
		Data   struct {
			Updated  string
			Parkings json.RawMessage `json:"parkinglocations"`
		}
	}
	file, err := os.Open("dummy.json")
	if err != nil {
		return Result{}, err
	}
	defer file.Close()
	dec := json.NewDecoder(file)
	b := &body{}
	err = dec.Decode(b)
	if err != nil {
		return Result{}, err
	}
	err = file.Close()
	if err != nil {
		return Result{}, err
	}
	if b.Status != "success" {
		return Result{}, ErrAPI
	}
	t, err := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", b.Data.Updated)
	if err != nil {
		return Result{}, err
	}
	t = t.UTC()
	res := Result{Updated: t}
	if t.Equal(updated) || t.Before(updated) {
		return res, ErrNoUpdate
	}

	var rawParkings []rawParking
	err = json.Unmarshal(b.Data.Parkings, &rawParkings)
	if err != nil {
		return res, err
	}
	res.Zones = make(map[int]string)
	res.Parkings = make([]parken.Parking, 0, len(b.Data.Parkings))
	for i := 0; i < len(rawParkings); i++ {
		raw := &rawParkings[i]
		if raw.Closed {
			continue
		}
		id, err := strconv.Atoi(raw.ID)
		if err != nil {
			return res, fmt.Errorf("parsing ID: %w", err)
		}
		zoneID, err := strconv.Atoi(raw.Zone.ID)
		if err != nil {
			return res, fmt.Errorf("parsing ID of zone: %w", err)
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
		// to-do: Verify consistency between ID and name.
		if _, ok := res.Zones[zoneID]; !ok {
			res.Zones[zoneID] = raw.Zone.Name
		}
		p := parken.Parking{
			ID:               id,
			Name:             raw.Name,
			Zone:             zoneID,
			Operator:         raw.Operator,
			Address:          address,
			PhoneNumber:      raw.PhoneNumber,
			Website:          website,
			Email:            raw.Email,
			Prices:           raw.Prices,
			LongTermPrices:   raw.LongTermPrices,
			OpeningHours:     raw.OpeningHours,
			OpenAllDay:       raw.OpenAllDay,
			ChargingStations: raw.ChargingStations,
			Spots:            raw.Status.Capacity - raw.Status.Spots,
			Capacity:         raw.Status.Capacity,
		}
		res.Parkings = append(res.Parkings, p)
	}
	return res, nil
}
