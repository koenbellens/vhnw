// Command mockminer bootst XMRig na voor het testen van de agent ZONDER echt te
// minen. Het serveert een /2/summary endpoint met oplopende (nep)statistieken.
//
// Ondersteunde vlaggen (de rest wordt genegeerd, net als bij XMRig):
//
//	--http-host <host>     adres voor de nep-API
//	--http-port <poort>    poort voor de nep-API
//	--crash-after <dur>    sluit na deze tijd af (om de watchdog te testen)
//	--hang-after <dur>     bevries de API na deze tijd (om de watchdog te testen)
package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"
)

func main() {
	host := "127.0.0.1"
	port := "18000"
	var crashAfter, hangAfter time.Duration

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--http-host":
			if i+1 < len(args) {
				host = args[i+1]
				i++
			}
		case "--http-port":
			if i+1 < len(args) {
				port = args[i+1]
				i++
			}
		case "--crash-after":
			if i+1 < len(args) {
				crashAfter, _ = time.ParseDuration(args[i+1])
				i++
			}
		case "--hang-after":
			if i+1 < len(args) {
				hangAfter, _ = time.ParseDuration(args[i+1])
				i++
			}
		}
	}

	start := time.Now()
	var shares int64
	var hung atomic.Bool

	go func() {
		t := time.NewTicker(2 * time.Second)
		for range t.C {
			atomic.AddInt64(&shares, 1)
			fmt.Printf("[mockminer] accepted share #%d  ~%.0f H/s\n", atomic.LoadInt64(&shares), 1234.5)
		}
	}()

	if crashAfter > 0 {
		go func() {
			time.Sleep(crashAfter)
			fmt.Println("[mockminer] simuleer crash, afsluiten")
			os.Exit(1)
		}()
	}
	if hangAfter > 0 {
		go func() {
			time.Sleep(hangAfter)
			fmt.Println("[mockminer] simuleer hang, API reageert niet meer")
			hung.Store(true)
		}()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/2/summary", func(w http.ResponseWriter, r *http.Request) {
		if hung.Load() {
			// Simuleer een vastgelopen miner: nooit antwoorden.
			select {}
		}
		up := int64(time.Since(start).Seconds())
		hr := 1234.5
		resp := map[string]any{
			"version": "mock-6.21.0",
			"uptime":  up,
			"hashrate": map[string]any{
				"total":   []float64{hr, hr * 0.98, hr * 0.95},
				"highest": hr * 1.05,
			},
			"results": map[string]any{
				"hashes_total": up * 1234,
				"shares_good":  atomic.LoadInt64(&shares),
				"shares_total": atomic.LoadInt64(&shares),
				"best":         []int64{100000},
			},
			"connection": map[string]any{
				"pool":     "mock.unmineable.local:3333",
				"uptime":   up,
				"ping":     12,
				"accepted": atomic.LoadInt64(&shares),
				"rejected": 0,
			},
			"cpu": map[string]any{"brand": "Mock CPU @ test"},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	addr := net.JoinHostPort(host, port)
	fmt.Printf("[mockminer] nep-API op http://%s/2/summary\n", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Println("[mockminer] fout:", err)
		os.Exit(1)
	}
}
