//go:build !windows

package system

import (
	"os"
	"strconv"
	"strings"
)

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
