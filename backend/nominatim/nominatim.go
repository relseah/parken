package nominatim

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/relseah/parken"
)

var defaultBaseURL = &url.URL{Scheme: "https", Host: "nominatim.openstreetmap.org"}

type place struct {
	Latitude  string `json:"lat"`
	Longitude string `json:"lon"`
}

type Client struct {
	BaseURL    *url.URL
	HTTPClient *http.Client

	rate      int
	remaining int
	ticker    *time.Ticker
	mutex     sync.Mutex
}

func (c *Client) httpClient() *http.Client {
	httpClient := c.HTTPClient
	if httpClient != nil {
		return httpClient
	}
	return http.DefaultClient
}

func (c *Client) limit() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.rate == 0 {
		return
	}
	reset := func() {
		c.remaining = c.rate - 1
	}
	select {
	case <-c.ticker.C:
		reset()
		return
	default:
	}
	if c.remaining > 0 {
		c.remaining--
	} else {
		<-c.ticker.C
		reset()
	}
}

func (c *Client) SetRate(rate int, interval time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.rate, c.remaining = rate, rate
	if rate == 0 {
		if c.ticker != nil {
			c.ticker.Stop()
			c.ticker = nil
		}
		return
	}
	if c.ticker != nil {
		c.ticker.Reset(interval)
	} else {
		c.ticker = time.NewTicker(interval)
	}
}

func (c *Client) Search(parking *parken.Parking) ([]parken.Coordinates, error) {
	u := c.BaseURL
	if u == nil {
		u = defaultBaseURL
	}
	q := u.Query()
	if u == defaultBaseURL {
		q.Set("format", "jsonv2")
	}
	q.Set("q", fmt.Sprintf("P%d %s, Heidelberg", parking.ID, parking.Name))
	u.RawQuery = q.Encode()
	u.Path = "/search"

	c.limit()
	resp, err := c.httpClient().Get(u.String())
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(resp.Body)
	var results []place
	err = dec.Decode(&results)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	coordinates := make([]parken.Coordinates, len(results))
	for i, res := range results {
		latitude, err := strconv.ParseFloat(res.Latitude, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing latitude: %w", err)
		}
		longitude, err := strconv.ParseFloat(res.Longitude, 64)
		if err != nil {
			return nil, fmt.Errorf("parsing longitude: %w", err)
		}
		coordinates[i] = parken.Coordinates{Latitude: latitude, Longitude: longitude}
	}
	return coordinates, nil
}

func NewClient(rate int, interval time.Duration) *Client {
	c := new(Client)
	c.SetRate(rate, interval)
	return c
}
