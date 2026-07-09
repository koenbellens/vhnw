// Package xmrig praat met de ingebouwde HTTP API van XMRig en zet de
// /2/summary respons om in een handzame struct.
package xmrig

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Summary is een uitgeklede weergave van XMRig's /2/summary endpoint, met
// alleen de velden die de agent en web-UI nodig hebben.
type Summary struct {
	Version  string `json:"version"`
	Uptime   int64  `json:"uptime"`
	Hashrate struct {
		// Total bevat [10s, 60s, 15m] gemiddelden. Waarden kunnen null zijn
		// vlak na het starten, vandaar *float64.
		Total   []*float64 `json:"total"`
		Highest float64    `json:"highest"`
	} `json:"hashrate"`
	Results struct {
		HashesTotal int64   `json:"hashes_total"`
		SharesGood  int64   `json:"shares_good"`
		SharesTotal int64   `json:"shares_total"`
		Best        []int64 `json:"best"`
	} `json:"results"`
	Connection struct {
		Pool     string  `json:"pool"`
		Uptime   int64   `json:"uptime"`
		Ping     float64 `json:"ping"`
		Accepted int64   `json:"accepted"`
		Rejected int64   `json:"rejected"`
	} `json:"connection"`
	CPU struct {
		Brand string `json:"brand"`
	} `json:"cpu"`
}

// HashrateNow geeft het meest recente (10s) hashrate-gemiddelde in H/s.
func (s Summary) HashrateNow() float64 {
	if len(s.Hashrate.Total) > 0 && s.Hashrate.Total[0] != nil {
		return *s.Hashrate.Total[0]
	}
	return 0
}

// Hashrate15m geeft het 15-minuten gemiddelde in H/s (0 indien onbekend).
func (s Summary) Hashrate15m() float64 {
	if len(s.Hashrate.Total) > 2 && s.Hashrate.Total[2] != nil {
		return *s.Hashrate.Total[2]
	}
	return 0
}

// Client haalt statistieken op van een draaiende XMRig instance.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient maakt een client voor de XMRig HTTP API op host:port.
func NewClient(host string, port int) *Client {
	return &Client{
		baseURL: fmt.Sprintf("http://%s:%d", host, port),
		http:    &http.Client{Timeout: 3 * time.Second},
	}
}

// FetchSummary haalt /2/summary op en parset het resultaat.
func (c *Client) FetchSummary(ctx context.Context) (*Summary, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/2/summary", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("xmrig api status %d", resp.StatusCode)
	}
	var s Summary
	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, fmt.Errorf("xmrig summary decoden: %w", err)
	}
	return &s, nil
}
