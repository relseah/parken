package nominatim

import (
	"encoding/json"
	"errors"
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

var ErrNoResults = errors.New("no search results")

func (c *Client) FetchCoordinates(parkingID int) (parken.Coordinates, error) {
	u := c.BaseURL
	if u == nil {
		u = defaultBaseURL
	}
	q := u.Query()
	if u == defaultBaseURL {
		q.Set("format", "jsonv2")
	}
	q.Set("q", "Heidelberg, P"+strconv.Itoa(parkingID))
	u.RawQuery = q.Encode()
	u.Path = "/search"

	c.limit()
	resp, err := c.httpClient().Get(u.String())
	if err != nil {
		return parken.Coordinates{}, err
	}
	dec := json.NewDecoder(resp.Body)
	var results []place
	err = dec.Decode(&results)
	resp.Body.Close()
	if err != nil {
		return parken.Coordinates{}, err
	}
	if len(results) == 0 {
		return parken.Coordinates{}, ErrNoResults
	}
	var coordinates parken.Coordinates
	coordinates.Latitude, err = strconv.ParseFloat(results[0].Latitude, 64)
	if err != nil {
		return parken.Coordinates{}, fmt.Errorf("parsing latitude: %w", err)
	}
	coordinates.Longitude, err = strconv.ParseFloat(results[0].Longitude, 64)
	if err != nil {
		return parken.Coordinates{}, fmt.Errorf("parsing longitude: %w", err)
	}
	return coordinates, nil
}
