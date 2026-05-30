package miner

import (
	"strings"
	"testing"

	"github.com/koenbellens/vhnw/internal/config"
)

func TestBuildArgs(t *testing.T) {
	cfg := config.Default()
	cfg.Wallet = "BTC:abc"
	cfg.Worker = "laptop1"
	cfg.MinerAPIPort = 18000

	args := buildArgs(cfg)
	joined := strings.Join(args, " ")

	for _, want := range []string{"-a rx", "-o rx.unmineable.com:3333", "-u BTC:abc.laptop1", "--http-port 18000", "--no-color"} {
		if !strings.Contains(joined, want) {
			t.Errorf("args missen %q; kreeg: %s", want, joined)
		}
	}
}

func TestBuildArgsTLSAndThreads(t *testing.T) {
	cfg := config.Default()
	cfg.Wallet = "BTC:abc"
	cfg.TLS = true
	cfg.Threads = 4

	joined := strings.Join(buildArgs(cfg), " ")
	if !strings.Contains(joined, "--tls") {
		t.Error("verwachtte --tls flag")
	}
	if !strings.Contains(joined, "-t 4") {
		t.Error("verwachtte -t 4 voor threads")
	}
}

func TestMaskSecret(t *testing.T) {
	if got := maskSecret("BTC:bc1qverylongwalletaddress1234"); strings.Contains(got, "verylong") {
		t.Errorf("wallet niet gemaskeerd: %s", got)
	}
	if got := maskSecret("short"); got != "***" {
		t.Errorf("korte string moet *** zijn, kreeg %s", got)
	}
}

func TestRedactArgsHidesUsername(t *testing.T) {
	args := []string{"-a", "rx", "-u", "BTC:bc1qverylongwalletaddress1234", "-p", "x"}
	out := redactArgs(args)
	if strings.Contains(out, "verylongwalletaddress") {
		t.Errorf("username niet geredacteerd: %s", out)
	}
}
