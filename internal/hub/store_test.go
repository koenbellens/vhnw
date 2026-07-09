package hub

import (
	"testing"
	"time"

	"github.com/koenbellens/vhnw/internal/community"
)

func TestReportRejectsEmptyID(t *testing.T) {
	s := NewStore("", time.Minute)
	if s.Report(community.Heartbeat{Name: "x"}) {
		t.Fatal("heartbeat zonder device_id had geweigerd moeten worden")
	}
	if got := s.Snapshot().TotalDevices; got != 0 {
		t.Fatalf("verwacht 0 apparaten, kreeg %d", got)
	}
}

func TestSnapshotTotalsAndOnline(t *testing.T) {
	s := NewStore("", 60*time.Second)
	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return base }

	s.Report(community.Heartbeat{DeviceID: "a", Name: "pi", Type: community.TypeCPU, Hashrate: 300, SharesGood: 10})
	s.Report(community.Heartbeat{DeviceID: "b", Name: "laptop", Type: community.TypeCPU, Hashrate: 1500, SharesGood: 5})

	// Apparaat 'a' verouderen tot voorbij de offline-grens.
	s.devices["a"].LastSeen = base.Add(-2 * time.Minute)

	snap := s.Snapshot()
	if snap.TotalDevices != 2 {
		t.Fatalf("verwacht 2 apparaten, kreeg %d", snap.TotalDevices)
	}
	if snap.OnlineDevices != 1 {
		t.Fatalf("verwacht 1 online, kreeg %d", snap.OnlineDevices)
	}
	// Alleen online apparaten tellen mee voor de totale hashrate.
	if snap.TotalHashrate != 1500 {
		t.Fatalf("verwacht totale hashrate 1500, kreeg %v", snap.TotalHashrate)
	}
	// Shares tellen voor alle apparaten.
	if snap.TotalShares != 15 {
		t.Fatalf("verwacht 15 shares, kreeg %d", snap.TotalShares)
	}
	// Online apparaat hoort bovenaan te staan.
	if !snap.Devices[0].Online {
		t.Fatalf("verwacht online apparaat bovenaan, kreeg offline")
	}
}

func TestSanitizeUnknownType(t *testing.T) {
	s := NewStore("", time.Minute)
	s.Report(community.Heartbeat{DeviceID: "x", Type: "rommel"})
	if got := s.devices["x"].Type; got != community.TypeCPU {
		t.Fatalf("onbekend type had naar cpu-agent moeten vallen, kreeg %q", got)
	}
}

func TestReportUpsertKeepsFirstSeen(t *testing.T) {
	s := NewStore("", time.Minute)
	t0 := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	s.now = func() time.Time { return t0 }
	s.Report(community.Heartbeat{DeviceID: "a", Hashrate: 100})
	first := s.devices["a"].FirstSeen

	s.now = func() time.Time { return t0.Add(time.Hour) }
	s.Report(community.Heartbeat{DeviceID: "a", Hashrate: 200})
	if !s.devices["a"].FirstSeen.Equal(first) {
		t.Fatal("FirstSeen mag niet veranderen bij een nieuwe heartbeat")
	}
	if s.devices["a"].Hashrate != 200 {
		t.Fatalf("verwacht bijgewerkte hashrate 200, kreeg %v", s.devices["a"].Hashrate)
	}
}
