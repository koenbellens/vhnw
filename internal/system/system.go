// Package system leest basale systeeminformatie uit (hostname, temperatuur,
// uptime) zodat de web-UI kan tonen hoe het apparaat erbij staat.
package system

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Info beschrijft de toestand van het apparaat waarop de agent draait.
type Info struct {
	Hostname    string  `json:"hostname"`
	OS          string  `json:"os"`
	TempC       float64 `json:"temp_c"`       // CPU-temperatuur in graden Celsius (0 = onbekend)
	UptimeSecs  int64   `json:"uptime_secs"`  // uptime van het systeem
	LoadAverage float64 `json:"load_average"` // 1-minuut load average (Linux)
}

// Collect verzamelt de huidige systeeminformatie.
func Collect() Info {
	host, _ := os.Hostname()
	return Info{
		Hostname:    host,
		OS:          "linux",
		TempC:       cpuTemperature(),
		UptimeSecs:  systemUptime(),
		LoadAverage: loadAverage(),
	}
}

// cpuTemperature leest de CPU-temperatuur uit /sys/class/thermal. Dit werkt op
// de meeste Linux-laptops en op de Raspberry Pi. Geeft 0 terug als er geen
// sensor wordt gevonden.
func cpuTemperature() float64 {
	for i := 0; i < 16; i++ {
		base := "/sys/class/thermal/thermal_zone" + strconv.Itoa(i)
		raw, err := os.ReadFile(base + "/temp")
		if err != nil {
			continue
		}
		milli, err := strconv.Atoi(strings.TrimSpace(string(raw)))
		if err != nil {
			continue
		}
		// Waarden staan doorgaans in milli-graden Celsius.
		return float64(milli) / 1000.0
	}
	return 0
}

func systemUptime() int64 {
	raw, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(raw))
	if len(fields) == 0 {
		return 0
	}
	secs, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0
	}
	return int64(secs)
}

func loadAverage() float64 {
	raw, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0
	}
	fields := strings.Fields(string(raw))
	if len(fields) == 0 {
		return 0
	}
	load, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0
	}
	return load
}

// FormatUptime maakt van seconden een leesbare string zoals "2d 3u 15m".
func FormatUptime(secs int64) string {
	if secs <= 0 {
		return "0s"
	}
	d := time.Duration(secs) * time.Second
	days := int64(d.Hours()) / 24
	hours := int64(d.Hours()) % 24
	mins := int64(d.Minutes()) % 60
	s := int64(d.Seconds()) % 60
	parts := []string{}
	if days > 0 {
		parts = append(parts, strconv.FormatInt(days, 10)+"d")
	}
	if hours > 0 {
		parts = append(parts, strconv.FormatInt(hours, 10)+"u")
	}
	if mins > 0 {
		parts = append(parts, strconv.FormatInt(mins, 10)+"m")
	}
	if len(parts) == 0 {
		parts = append(parts, strconv.FormatInt(s, 10)+"s")
	}
	return strings.Join(parts, " ")
}
