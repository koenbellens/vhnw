package community

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/koenbellens/vhnw/internal/config"
	"github.com/koenbellens/vhnw/internal/miner"
	"github.com/koenbellens/vhnw/internal/system"
)

// Reporter stuurt periodiek een heartbeat van dit apparaat naar de hub.
type Reporter struct {
	mgr     *miner.Manager
	getCfg  func() config.Config
	version string
	client  *http.Client

	lastOK bool // voorkomt log-spam: alleen statuswissels worden gelogd
}

// NewReporter maakt een reporter. getCfg levert telkens de actuele configuratie
// op zodat wijzigingen in de web-UI meteen meetellen.
func NewReporter(mgr *miner.Manager, getCfg func() config.Config, version string) *Reporter {
	return &Reporter{
		mgr:     mgr,
		getCfg:  getCfg,
		version: version,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

// Run blijft draaien tot ctx wordt geannuleerd en verstuurt elke interval een
// heartbeat (mits community in de config is ingeschakeld).
func (r *Reporter) Run(ctx context.Context) {
	interval := time.Duration(r.getCfg().Community.IntervalSeconds) * time.Second
	if interval <= 0 {
		interval = 15 * time.Second
	}
	t := time.NewTicker(interval)
	defer t.Stop()

	r.sendOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			r.sendOnce(ctx)
		}
	}
}

func (r *Reporter) sendOnce(ctx context.Context) {
	cfg := r.getCfg()
	if !cfg.Community.Enabled || strings.TrimSpace(cfg.Community.HubURL) == "" {
		return
	}

	hb := r.build(ctx, cfg)
	body, err := json.Marshal(hb)
	if err != nil {
		return
	}

	url := strings.TrimRight(cfg.Community.HubURL, "/") + "/api/report"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.Community.Token != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.Community.Token)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		r.note(false, "[hub] versturen mislukt: "+err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		r.note(false, fmt.Sprintf("[hub] hub gaf status %d", resp.StatusCode))
		return
	}
	r.note(true, "[hub] status gedeeld met community-dashboard")
}

// build stelt de heartbeat samen uit de actuele miner- en systeemstatus.
func (r *Reporter) build(ctx context.Context, cfg config.Config) Heartbeat {
	st := r.mgr.Status()
	sys := system.Collect()

	hb := Heartbeat{
		DeviceID:     cfg.Community.DeviceID,
		Name:         cfg.DeviceName(),
		Type:         cfg.Community.DeviceType,
		Algo:         cfg.Algo,
		Pool:         cfg.Pool,
		State:        st.State,
		TempC:        sys.TempC,
		UptimeSecs:   sys.UptimeSecs,
		Restarts:     st.RestartCount,
		AgentVersion: r.version,
	}

	if st.State == miner.StateRunning {
		sctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		if sum, err := r.mgr.Summary(sctx); err == nil {
			hb.Hashrate = sum.HashrateNow()
			hb.SharesGood = sum.Results.SharesGood
			hb.SharesTotal = sum.Results.SharesTotal
			if sum.Version != "" {
				hb.Miner = "XMRig " + sum.Version
			}
			if hb.Pool == "" {
				hb.Pool = sum.Connection.Pool
			}
		}
		cancel()
	}
	return hb
}

// note logt alleen wanneer de hub-status verandert (online <-> fout), zodat de
// miner-log niet volloopt.
func (r *Reporter) note(ok bool, msg string) {
	if ok == r.lastOK {
		return
	}
	r.lastOK = ok
	r.mgr.AddLog(msg)
}
