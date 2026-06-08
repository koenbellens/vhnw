// Package hub is de centrale server van het VHNW community-dashboard (Laag 3).
// Apparaten sturen hun status (heartbeat) naar de hub; de hub bewaart de laatste
// stand per apparaat en toont alles samen op één pagina.
package hub

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/koenbellens/vhnw/internal/community"
)

// Device is een apparaat zoals de hub het kent: de laatste heartbeat plus de
// momenten waarop het voor het eerst en laatst van zich liet horen.
type Device struct {
	community.Heartbeat
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
}

// DeviceView voegt afgeleide velden toe voor het dashboard.
type DeviceView struct {
	Device
	Online       bool  `json:"online"`
	LastSeenSecs int64 `json:"last_seen_secs"`
}

// Snapshot is het complete beeld dat de dashboard-API teruggeeft.
type Snapshot struct {
	Devices       []DeviceView `json:"devices"`
	TotalDevices  int          `json:"total_devices"`
	OnlineDevices int          `json:"online_devices"`
	TotalHashrate float64      `json:"total_hashrate"`
	TotalShares   int64        `json:"total_shares"`
	GeneratedAt   string       `json:"generated_at"`
}

// Store bewaart de apparaten, met optionele persistentie naar een JSON-bestand.
type Store struct {
	mu           sync.Mutex
	devices      map[string]*Device
	path         string
	offlineAfter time.Duration
	now          func() time.Time
}

// NewStore maakt een store. path mag leeg zijn (dan is er geen persistentie).
// offlineAfter bepaalt na hoeveel stilte een apparaat als offline geldt.
func NewStore(path string, offlineAfter time.Duration) *Store {
	s := &Store{
		devices:      make(map[string]*Device),
		path:         path,
		offlineAfter: offlineAfter,
		now:          time.Now,
	}
	s.load()
	return s
}

// Report verwerkt een binnenkomende heartbeat. Een lege DeviceID wordt
// geweigerd (false).
func (s *Store) Report(hb community.Heartbeat) bool {
	if hb.DeviceID == "" {
		return false
	}
	hb = sanitize(hb)

	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	dev, ok := s.devices[hb.DeviceID]
	if !ok {
		dev = &Device{FirstSeen: now}
		s.devices[hb.DeviceID] = dev
	}
	dev.Heartbeat = hb
	dev.LastSeen = now

	s.save()
	return true
}

// Snapshot geeft het huidige beeld terug, gesorteerd: online eerst, daarna op
// naam.
func (s *Store) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	snap := Snapshot{GeneratedAt: now.Format(time.RFC3339)}
	for _, dev := range s.devices {
		online := now.Sub(dev.LastSeen) <= s.offlineAfter
		v := DeviceView{
			Device:       *dev,
			Online:       online,
			LastSeenSecs: int64(now.Sub(dev.LastSeen).Seconds()),
		}
		snap.Devices = append(snap.Devices, v)
		snap.TotalDevices++
		if online {
			snap.OnlineDevices++
			snap.TotalHashrate += dev.Hashrate
		}
		snap.TotalShares += dev.SharesGood
	}
	sort.Slice(snap.Devices, func(i, j int) bool {
		a, b := snap.Devices[i], snap.Devices[j]
		if a.Online != b.Online {
			return a.Online // online apparaten bovenaan
		}
		return a.Name < b.Name
	})
	return snap
}

func (s *Store) load() {
	if s.path == "" {
		return
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	var devices map[string]*Device
	if err := json.Unmarshal(data, &devices); err != nil {
		return
	}
	if devices != nil {
		s.devices = devices
	}
}

// save schrijft de huidige stand weg. De caller houdt de lock vast.
func (s *Store) save() {
	if s.path == "" {
		return
	}
	data, err := json.MarshalIndent(s.devices, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(s.path, data, 0o644)
}

// sanitize beperkt tekstvelden en normaliseert het apparaat-type.
func sanitize(hb community.Heartbeat) community.Heartbeat {
	hb.Name = clip(hb.Name, 48)
	hb.Owner = clip(hb.Owner, 48)
	hb.Pool = clip(hb.Pool, 64)
	hb.Miner = clip(hb.Miner, 48)
	hb.Algo = clip(hb.Algo, 24)
	hb.State = clip(hb.State, 24)
	switch hb.Type {
	case community.TypeCPU, community.TypeASIC, community.TypeGPU:
		// bekend type, laat staan
	default:
		hb.Type = community.TypeCPU
	}
	if hb.Hashrate < 0 {
		hb.Hashrate = 0
	}
	return hb
}

func clip(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}
