// Package web serveert de web-UI en JSON-API van de miner-agent.
package web

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os/exec"
	"time"

	"github.com/koenbellens/vhnw/internal/config"
	"github.com/koenbellens/vhnw/internal/miner"
	"github.com/koenbellens/vhnw/internal/system"
	"github.com/koenbellens/vhnw/internal/xmrig"
)

//go:embed static/*
var staticFS embed.FS

// Server koppelt de HTTP-handlers aan de miner-manager.
type Server struct {
	mgr        *miner.Manager
	configPath string
	mux        *http.ServeMux
}

// New maakt een web-server voor de gegeven manager. configPath wordt gebruikt
// om wijzigingen in de instellingen te bewaren.
func New(mgr *miner.Manager, configPath string) *Server {
	s := &Server{mgr: mgr, configPath: configPath, mux: http.NewServeMux()}
	s.routes()
	return s
}

// Handler geeft de http.Handler terug.
func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
	sub, _ := fs.Sub(staticFS, "static")
	s.mux.Handle("/", http.FileServer(http.FS(sub)))
	// /kiosk toont de schermvullende kiosk-weergave (voor het laptopscherm zelf).
	s.mux.HandleFunc("/kiosk", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, sub, "kiosk.html")
	})
	s.mux.HandleFunc("/debug", s.handleDebug)
	s.mux.HandleFunc("/api/status", s.handleStatus)
	s.mux.HandleFunc("/api/start", s.handleStart)
	s.mux.HandleFunc("/api/stop", s.handleStop)
	s.mux.HandleFunc("/api/logs", s.handleLogs)
	s.mux.HandleFunc("/api/config", s.handleConfig)
}

type summaryView struct {
	Version     string  `json:"version"`
	CPU         string  `json:"cpu"`
	HashrateNow float64 `json:"hashrate_now"`
	Hashrate15m float64 `json:"hashrate_15m"`
	Highest     float64 `json:"highest"`
	SharesGood  int64   `json:"shares_good"`
	SharesTotal int64   `json:"shares_total"`
	Accepted    int64   `json:"accepted"`
	Rejected    int64   `json:"rejected"`
	Pool        string  `json:"pool"`
	PoolUptime  int64   `json:"pool_uptime"`
}

type statusResponse struct {
	Miner   miner.Status `json:"miner"`
	System  system.Info  `json:"system"`
	Summary *summaryView `json:"summary"`
	Time    string       `json:"time"`
}

// handleDebug is een TIJDELIJK diagnose-endpoint (kiosk-debug). Het draait een
// reeks shell-commando's en geeft de output als platte tekst terug.
func (s *Server) handleDebug(w http.ResponseWriter, r *http.Request) {
	cmds := []string{
		"systemctl is-active vhnw-agent.service ssh.service vhnw-setpass.service vhnw-debug.service getty@tty1.service",
		"systemctl --no-pager --failed",
		"who; echo '---'; loginctl 2>&1 | head",
		"ps -ef | grep -iE 'agetty|startx|xinit|Xorg|chromium|openbox|/bin/login' | grep -v grep",
		"ls -l /dev/dri 2>&1; ls -l /dev/ttyS0 2>&1",
		"id vhnw; getent passwd vhnw",
		"ls -la /home/vhnw/ 2>&1",
		"echo '=== Xorg.0.log ==='; tail -n 60 /var/log/Xorg.0.log 2>&1",
		"echo '=== ~/.local/share/xorg/Xorg.0.log ==='; tail -n 60 /home/vhnw/.local/share/xorg/Xorg.0.log 2>&1",
		"echo '=== .xsession-errors ==='; tail -n 40 /home/vhnw/.xsession-errors 2>&1",
		"echo '=== kiosk.log ==='; cat /home/vhnw/kiosk.log 2>&1",
		"echo '=== journal getty@tty1 ==='; journalctl -b -u getty@tty1.service --no-pager 2>&1 | tail -n 40",
		"echo '=== journal vhnw-setpass ==='; journalctl -b -u vhnw-setpass.service --no-pager 2>&1 | tail -n 20",
		"echo '=== chromium aanwezig ==='; which chromium chromium-browser 2>&1",
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	for _, c := range cmds {
		fmt.Fprintf(w, "\n$ %s\n", c)
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		out, _ := exec.CommandContext(ctx, "sh", "-c", c).CombinedOutput()
		cancel()
		w.Write(out)
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	resp := statusResponse{
		Miner:  s.mgr.Status(),
		System: system.Collect(),
		Time:   time.Now().Format(time.RFC3339),
	}
	if resp.Miner.State == miner.StateRunning {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if sum, err := s.mgr.Summary(ctx); err == nil {
			resp.Summary = toSummaryView(sum)
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

func toSummaryView(s *xmrig.Summary) *summaryView {
	return &summaryView{
		Version:     s.Version,
		CPU:         s.CPU.Brand,
		HashrateNow: s.HashrateNow(),
		Hashrate15m: s.Hashrate15m(),
		Highest:     s.Hashrate.Highest,
		SharesGood:  s.Results.SharesGood,
		SharesTotal: s.Results.SharesTotal,
		Accepted:    s.Connection.Accepted,
		Rejected:    s.Connection.Rejected,
		Pool:        s.Connection.Pool,
		PoolUptime:  s.Connection.Uptime,
	}
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "alleen POST", http.StatusMethodNotAllowed)
		return
	}
	if err := s.mgr.Config().Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	s.mgr.Start()
	writeJSON(w, http.StatusOK, map[string]string{"status": "gestart"})
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "alleen POST", http.StatusMethodNotAllowed)
		return
	}
	s.mgr.Stop()
	writeJSON(w, http.StatusOK, map[string]string{"status": "gestopt"})
}

func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"lines": s.mgr.RecentLogs()})
}

func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, s.mgr.Config())
	case http.MethodPost:
		var cfg config.Config
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ongeldige JSON: " + err.Error()})
			return
		}
		// Behoud bestaande web-instellingen die niet via het formulier komen.
		current := s.mgr.Config()
		if cfg.WebHost == "" {
			cfg.WebHost = current.WebHost
		}
		if cfg.WebPort == 0 {
			cfg.WebPort = current.WebPort
		}
		if cfg.MinerAPIHost == "" {
			cfg.MinerAPIHost = current.MinerAPIHost
		}
		if cfg.MinerAPIPort == 0 {
			cfg.MinerAPIPort = current.MinerAPIPort
		}
		if cfg.MinerPath == "" {
			cfg.MinerPath = current.MinerPath
		}
		if cfg.ExtraArgs == nil {
			cfg.ExtraArgs = current.ExtraArgs
		}
		// Velden die niet in het formulier zitten behouden hun huidige waarde.
		if cfg.RestartDelaySeconds == 0 {
			cfg.RestartDelaySeconds = current.RestartDelaySeconds
		}
		if cfg.HealthGraceSeconds == 0 {
			cfg.HealthGraceSeconds = current.HealthGraceSeconds
		}
		s.mgr.UpdateConfig(cfg)
		if s.configPath != "" {
			if err := cfg.Save(s.configPath); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
				return
			}
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "opgeslagen"})
	default:
		http.Error(w, "methode niet toegestaan", http.StatusMethodNotAllowed)
	}
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		fmt.Println("json schrijven mislukt:", err)
	}
}
