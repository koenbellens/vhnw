// Command vhnw-hub is de centrale server van het VHNW community-dashboard
// (Laag 3). Mining-apparaten sturen hier periodiek hun status naartoe; de hub
// toont ze samen op één live pagina.
//
// Instellingen via vlaggen of omgevingsvariabelen:
//
//	--addr / PORT            luisteradres (PORT zet ":$PORT")
//	--data / VHNW_HUB_DATA   pad naar het opslagbestand (JSON)
//	--token / VHNW_HUB_TOKEN gedeeld geheim; leeg = elke melding wordt geaccepteerd
//	--offline / VHNW_HUB_OFFLINE  seconden stilte voordat een apparaat offline heet
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/koenbellens/vhnw/internal/hub"
)

func main() {
	addr := flag.String("addr", envOr("ADDR", ":8088"), "luisteradres van de hub (host:poort)")
	dataPath := flag.String("data", envOr("VHNW_HUB_DATA", "hub-data.json"), "pad naar het opslagbestand")
	token := flag.String("token", os.Getenv("VHNW_HUB_TOKEN"), "gedeeld geheim voor rapportage (leeg = open)")
	offlineSecs := flag.Int("offline", envInt("VHNW_HUB_OFFLINE", 60), "seconden stilte voordat een apparaat offline is")
	flag.Parse()

	// Op hosting-platforms (Fly, Heroku, ...) komt de poort vaak via $PORT.
	listenAddr := *addr
	if p := os.Getenv("PORT"); p != "" {
		listenAddr = ":" + p
	}

	store := hub.NewStore(*dataPath, time.Duration(*offlineSecs)*time.Second)
	srv := hub.NewServer(store, *token)

	if *token == "" {
		log.Print("LET OP: hub draait zonder token (open). Zet VHNW_HUB_TOKEN voor productie.")
	}

	httpServer := &http.Server{
		Addr:              listenAddr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("VHNW community-hub luistert op http://%s", listenAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("hub web-server fout: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
