package hub

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/koenbellens/vhnw/internal/community"
)

//go:embed static/*
var staticFS embed.FS

// Server serveert het community-dashboard en het rapportage-endpoint.
type Server struct {
	store *Store
	token string
	mux   *http.ServeMux
}

// NewServer maakt een hub-server. Is token leeg, dan accepteert de hub elke
// melding (handig voor lokaal testen); anders moet elke melding het token
// meesturen via "Authorization: Bearer <token>".
func NewServer(store *Store, token string) *Server {
	s := &Server{store: store, token: token, mux: http.NewServeMux()}
	sub, _ := fs.Sub(staticFS, "static")
	s.mux.Handle("/", http.FileServer(http.FS(sub)))
	s.mux.HandleFunc("/api/report", s.handleReport)
	s.mux.HandleFunc("/api/devices", s.handleDevices)
	s.mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	return s
}

// Handler geeft de http.Handler terug.
func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) handleReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "alleen POST", http.StatusMethodNotAllowed)
		return
	}
	if !s.authorized(r) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "ongeldig token"})
		return
	}
	var hb community.Heartbeat
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&hb); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "ongeldige JSON"})
		return
	}
	if !s.store.Report(hb) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "device_id ontbreekt"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.store.Snapshot())
}

func (s *Server) authorized(r *http.Request) bool {
	if s.token == "" {
		return true
	}
	h := r.Header.Get("Authorization")
	return strings.TrimSpace(strings.TrimPrefix(h, "Bearer ")) == s.token
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
