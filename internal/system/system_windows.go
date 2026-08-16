//go:build windows

package system

import (
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	procGetTickCount64 = kernel32.NewProc("GetTickCount64")
)

// systemUptime gebruikt Windows' eigen GetTickCount64 (milliseconden sinds
// opstarten) — geen extra afhankelijkheden nodig.
func systemUptime() int64 {
	ret, _, _ := procGetTickCount64.Call()
	return int64(ret) / 1000
}

// cpuTemperature probeert de CPU-temperatuur via WMI's ACPI-thermalzone op te
// vragen. Veel (vooral zakelijke) desktopmoederborden geven dit niet door via
// ACPI, dan komt hier gewoon 0 uit (net als bij een ontbrekende sensor op
// Linux) — geen harde fout, alleen "onbekend".
func cpuTemperature() float64 {
	out, err := exec.Command("wmic", "/namespace:\\\\root\\wmi", "PATH",
		"MSAcpi_ThermalZoneTemperature", "get", "CurrentTemperature").Output()
	if err != nil {
		return 0
	}
	for _, field := range strings.Fields(string(out)) {
		if field == "CurrentTemperature" {
			continue
		}
		deciKelvin, err := strconv.ParseFloat(field, 64)
		if err != nil {
			continue
		}
		return deciKelvin/10.0 - 273.15
	}
	return 0
}

// loadAverage heeft geen directe Windows-tegenhanger (dat concept bestaat
// alleen op Unix); geeft altijd 0 terug.
func loadAverage() float64 {
	return 0
}
