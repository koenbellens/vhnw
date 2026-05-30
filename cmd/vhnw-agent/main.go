// Command vhnw-agent is de mining-agent van "Van Hash Naar Winst".
//
// Het beheert XMRig (start/stop + watchdog) en serveert een web-UI met een
// grote start/stop-knop en live statistieken. Bedoeld om op een oude laptop of
// Raspberry Pi te draaien.
package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/koenbellens/vhnw/internal/config"
	"github.com/koenbellens/vhnw/internal/miner"
	"github.com/koenbellens/vhnw/internal/web"
)

func main() {
	configPath := flag.String("config", "config.json", "pad naar het configuratiebestand")
	addr := flag.String("addr", "", "overschrijf het luisteradres van de web-UI (host:poort)")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config laden mislukt: %v", err)
	}

	mgr := miner.New(cfg)
	srv := web.New(mgr, *configPath)

	listenAddr := net.JoinHostPort(cfg.WebHost, strconv.Itoa(cfg.WebPort))
	if *addr != "" {
		listenAddr = *addr
	}

	httpServer := &http.Server{
		Addr:              listenAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	if cfg.AutoStart {
		if err := cfg.Validate(); err != nil {
			log.Printf("auto-start overgeslagen: %v", err)
		} else {
			log.Print("auto-start: miner wordt gestart")
			mgr.Start()
		}
	}

	go func() {
		log.Printf("VHNW miner-agent luistert op http://%s", listenAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("web-server fout: %v", err)
		}
	}()

	// Netjes afsluiten bij Ctrl-C of SIGTERM.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Print("afsluiten: miner stoppen…")
	mgr.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
}
