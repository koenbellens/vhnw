// Package system leest basale systeeminformatie uit (hostname, temperatuur,
// uptime) zodat de web-UI kan tonen hoe het apparaat erbij staat.
package system

import (
	"os"
	"runtime"
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
	LoadAverage float64 `json:"load_average"` // 1-minuut load average (alleen Unix; 0 op Windows)
}

// Collect verzamelt de huidige systeeminformatie. De daadwerkelijke metingen
// (temperatuur, uptime, load average) zitten OS-specifiek in
// system_unix.go/system_windows.go.
func Collect() Info {
	host, _ := os.Hostname()
	return Info{
		Hostname:    host,
		OS:          runtime.GOOS,
		TempC:       cpuTemperature(),
		UptimeSecs:  systemUptime(),
		LoadAverage: loadAverage(),
	}
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
