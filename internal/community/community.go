// Package community verzorgt de communicatie met de centrale VHNW-hub
// (Laag 3 community-dashboard). De agent stuurt periodiek een "heartbeat" met
// zijn status; de hub verzamelt die van alle apparaten en toont ze samen.
package community

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Heartbeat is het statusbericht dat een apparaat naar de hub stuurt. Het bevat
// bewust géén wallet of andere gevoelige gegevens — alleen wat nodig is om het
// apparaat in het dashboard te tonen.
type Heartbeat struct {
	DeviceID string `json:"device_id"`
	Name     string `json:"name"`
	// Type is "cpu-agent", "asic" of "gpu". Zo passen Antminers en GPU's later
	// vanzelf in hetzelfde dashboard.
	Type string `json:"type"`

	Miner string `json:"miner"` // bv. "XMRig 6.26.0"
	Algo  string `json:"algo"`
	Pool  string `json:"pool"`
	State string `json:"state"` // "actief", "gestopt", ...

	Hashrate    float64 `json:"hashrate"` // H/s
	SharesGood  int64   `json:"shares_good"`
	SharesTotal int64   `json:"shares_total"`
	TempC       float64 `json:"temp_c"`
	UptimeSecs  int64   `json:"uptime_secs"`
	Restarts    int     `json:"restarts"`

	AgentVersion string `json:"agent_version"`
}

// Known device types.
const (
	TypeCPU  = "cpu-agent"
	TypeASIC = "asic"
	TypeGPU  = "gpu"
)

// NewDeviceID maakt een stabiele, willekeurige id voor een apparaat
// ("vhnw-" + 16 hex-tekens). Valt random lezen weg, dan wordt de tijd gebruikt.
func NewDeviceID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("vhnw-%x", time.Now().UnixNano())
	}
	return "vhnw-" + hex.EncodeToString(b)
}
