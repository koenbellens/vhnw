package config

import (
	"path/filepath"
	"testing"
)

func TestMinerUsername(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		want string
	}{
		{"alleen wallet", Config{Wallet: "BTC:abc"}, "BTC:abc"},
		{"wallet+worker", Config{Wallet: "BTC:abc", Worker: "laptop1"}, "BTC:abc.laptop1"},
		{"met referral", Config{Wallet: "BTC:abc", Worker: "pi", Referral: "ref99"}, "BTC:abc.pi#ref99"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.cfg.MinerUsername(); got != c.want {
				t.Errorf("MinerUsername() = %q, wil %q", got, c.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	if err := Default().Validate(); err == nil {
		t.Error("lege default-config moet ongeldig zijn (geen wallet)")
	}
	ok := Default()
	ok.Wallet = "BTC:abc"
	if err := ok.Validate(); err != nil {
		t.Errorf("config met wallet moet geldig zijn, kreeg: %v", err)
	}
}

func TestLoadMissingFileGivesDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "bestaat-niet.json"))
	if err != nil {
		t.Fatalf("onverwachte fout: %v", err)
	}
	if cfg.WebPort != Default().WebPort {
		t.Errorf("verwachtte default web-poort, kreeg %d", cfg.WebPort)
	}
}

func TestSaveAndLoadRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	in := Default()
	in.Wallet = "DOGE:Dxyz"
	in.Worker = "pi4"
	if err := in.Save(path); err != nil {
		t.Fatalf("opslaan mislukt: %v", err)
	}
	out, err := Load(path)
	if err != nil {
		t.Fatalf("laden mislukt: %v", err)
	}
	if out.Wallet != in.Wallet || out.Worker != in.Worker {
		t.Errorf("roundtrip mismatch: %+v", out)
	}
}
