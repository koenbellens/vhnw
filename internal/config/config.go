// Package config laadt en bewaart de instellingen van de VHNW miner-agent.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Config bevat alle instellingen van de agent. De velden worden uit een
// JSON-bestand geladen (zie config.example.json) en kunnen vanuit de web-UI
// worden aangepast.
type Config struct {
	// Web-UI van de agent zelf.
	WebHost string `json:"web_host"`
	WebPort int    `json:"web_port"`

	// XMRig binary + mining-instellingen.
	MinerPath string `json:"miner_path"`
	Algo      string `json:"algo"`
	Pool      string `json:"pool"`
	TLS       bool   `json:"tls"`

	// unMineable / pool-account. Username wordt opgebouwd als
	// "<wallet>[.<worker>][#<referral>]".
	Wallet   string `json:"wallet"`
	Worker   string `json:"worker"`
	Referral string `json:"referral"`
	Password string `json:"password"`

	// XMRig's eigen HTTP API, gebruikt de agent om live statistieken op te halen.
	MinerAPIHost string `json:"miner_api_host"`
	MinerAPIPort int    `json:"miner_api_port"`

	// Extra command line argumenten die ongewijzigd aan XMRig worden meegegeven.
	ExtraArgs []string `json:"extra_args"`

	// Gedrag van de agent.
	AutoStart           bool `json:"auto_start"`
	RestartDelaySeconds int  `json:"restart_delay_seconds"`
	HealthGraceSeconds  int  `json:"health_grace_seconds"`
	// Aantal CPU threads (0 = XMRig bepaalt het zelf, automatisch).
	Threads int `json:"threads"`
}

// Default geeft een bruikbare standaardconfiguratie terug. De waarden zijn
// afgestemd op unMineable RandomX-mining (zoals BTC/andere coins minen met CPU).
func Default() Config {
	return Config{
		WebHost:             "0.0.0.0",
		WebPort:             8420,
		MinerPath:           "xmrig",
		Algo:                "rx",
		Pool:                "rx.unmineable.com:3333",
		TLS:                 false,
		Wallet:              "",
		Worker:              hostnameWorker(),
		Referral:            "",
		Password:            "x",
		MinerAPIHost:        "127.0.0.1",
		MinerAPIPort:        18000,
		ExtraArgs:           []string{},
		AutoStart:           false,
		RestartDelaySeconds: 5,
		HealthGraceSeconds:  90,
		Threads:             0,
	}
}

// Load leest de configuratie uit het opgegeven pad. Ontbrekende velden krijgen
// hun standaardwaarde. Bestaat het bestand niet, dan wordt de default-config
// teruggegeven (zonder fout) zodat de agent altijd kan starten.
func Load(path string) (Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("config lezen: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("config parsen (%s): %w", path, err)
	}
	cfg.applyDefaults()
	return cfg, nil
}

// Save schrijft de configuratie als nette JSON naar het opgegeven pad.
func (c Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("config opslaan: %w", err)
	}
	return nil
}

// MinerUsername bouwt de pool-username op uit wallet, worker en referral.
func (c Config) MinerUsername() string {
	user := c.Wallet
	if c.Worker != "" {
		user += "." + c.Worker
	}
	if c.Referral != "" {
		user += "#" + c.Referral
	}
	return user
}

// Validate controleert of de minimaal vereiste velden zijn ingevuld om te
// kunnen minen.
func (c Config) Validate() error {
	var missing []string
	if strings.TrimSpace(c.Wallet) == "" {
		missing = append(missing, "wallet")
	}
	if strings.TrimSpace(c.Pool) == "" {
		missing = append(missing, "pool")
	}
	if strings.TrimSpace(c.MinerPath) == "" {
		missing = append(missing, "miner_path")
	}
	if len(missing) > 0 {
		return fmt.Errorf("config onvolledig, vul in: %s", strings.Join(missing, ", "))
	}
	return nil
}

func (c *Config) applyDefaults() {
	d := Default()
	if c.WebHost == "" {
		c.WebHost = d.WebHost
	}
	if c.WebPort == 0 {
		c.WebPort = d.WebPort
	}
	if c.MinerPath == "" {
		c.MinerPath = d.MinerPath
	}
	if c.Algo == "" {
		c.Algo = d.Algo
	}
	if c.Pool == "" {
		c.Pool = d.Pool
	}
	if c.Password == "" {
		c.Password = d.Password
	}
	if c.MinerAPIHost == "" {
		c.MinerAPIHost = d.MinerAPIHost
	}
	if c.MinerAPIPort == 0 {
		c.MinerAPIPort = d.MinerAPIPort
	}
	if c.Worker == "" {
		c.Worker = d.Worker
	}
	if c.RestartDelaySeconds == 0 {
		c.RestartDelaySeconds = d.RestartDelaySeconds
	}
	if c.HealthGraceSeconds == 0 {
		c.HealthGraceSeconds = d.HealthGraceSeconds
	}
	if c.ExtraArgs == nil {
		c.ExtraArgs = []string{}
	}
}

func hostnameWorker() string {
	h, err := os.Hostname()
	if err != nil || h == "" {
		return "worker1"
	}
	return h
}
