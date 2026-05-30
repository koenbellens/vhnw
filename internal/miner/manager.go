// Package miner beheert het XMRig-proces: starten, stoppen en automatisch
// herstarten (watchdog) zodat de miner niet meer stilvalt.
package miner

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/koenbellens/vhnw/internal/config"
	"github.com/koenbellens/vhnw/internal/xmrig"
)

// Toestanden van de miner, in het Nederlands voor de web-UI.
const (
	StateStopped    = "gestopt"
	StateStarting   = "bezig met starten"
	StateRunning    = "actief"
	StateRestarting = "herstarten"
	StateError      = "fout"
)

const logBufferSize = 200

// Status is een momentopname van de toestand van de miner.
type Status struct {
	State        string    `json:"state"`
	Desired      bool      `json:"desired"`
	PID          int       `json:"pid"`
	RunningSecs  int64     `json:"running_secs"`
	StartedAt    time.Time `json:"started_at"`
	RestartCount int       `json:"restart_count"`
	LastError    string    `json:"last_error"`
	LastExit     string    `json:"last_exit"`
	ConfigValid  bool      `json:"config_valid"`
	ConfigError  string    `json:"config_error"`
}

// Manager superviseert precies één XMRig-proces.
type Manager struct {
	mu sync.Mutex

	cfg    config.Config
	client *xmrig.Client

	desired      bool
	supervising  bool
	state        string
	cmd          *exec.Cmd
	startedAt    time.Time
	restartCount int
	lastError    string
	lastExit     string

	// kill signaleert de huidige run dat hij gestopt moet worden.
	logs *ringBuffer
}

// New maakt een Manager met de opgegeven configuratie.
func New(cfg config.Config) *Manager {
	return &Manager{
		cfg:    cfg,
		client: xmrig.NewClient(cfg.MinerAPIHost, cfg.MinerAPIPort),
		state:  StateStopped,
		logs:   newRingBuffer(logBufferSize),
	}
}

// Config geeft een kopie van de huidige configuratie terug.
func (m *Manager) Config() config.Config {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cfg
}

// UpdateConfig past de configuratie aan. Draait de miner al, dan wordt hij met
// de nieuwe instellingen herstart.
func (m *Manager) UpdateConfig(cfg config.Config) {
	m.mu.Lock()
	m.cfg = cfg
	m.client = xmrig.NewClient(cfg.MinerAPIHost, cfg.MinerAPIPort)
	running := m.cmd != nil && m.cmd.Process != nil
	m.mu.Unlock()

	if running {
		// Stuur een signaal naar het lopende proces; de supervisor start het
		// daarna met de nieuwe argumenten opnieuw op.
		m.signalRestart()
	}
}

// Start zet de gewenste toestand op "draaien" en start de supervisor.
func (m *Manager) Start() {
	m.mu.Lock()
	m.desired = true
	m.lastError = ""
	if !m.supervising {
		m.supervising = true
		m.mu.Unlock()
		go m.supervise()
		return
	}
	m.mu.Unlock()
}

// Stop zet de gewenste toestand op "gestopt" en beëindigt het proces.
func (m *Manager) Stop() {
	m.mu.Lock()
	m.desired = false
	cmd := m.cmd
	m.mu.Unlock()
	killProcess(cmd)
}

// signalRestart beëindigt het huidige proces terwijl desired==true blijft,
// waardoor de supervisor het met verse config opnieuw start.
func (m *Manager) signalRestart() {
	m.mu.Lock()
	cmd := m.cmd
	m.mu.Unlock()
	killProcess(cmd)
}

// Status geeft een momentopname terug.
func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()

	st := Status{
		State:        m.state,
		Desired:      m.desired,
		RestartCount: m.restartCount,
		LastError:    m.lastError,
		LastExit:     m.lastExit,
		StartedAt:    m.startedAt,
	}
	if m.cmd != nil && m.cmd.Process != nil && m.state == StateRunning {
		st.PID = m.cmd.Process.Pid
		st.RunningSecs = int64(time.Since(m.startedAt).Seconds())
	}
	if err := m.cfg.Validate(); err != nil {
		st.ConfigError = err.Error()
	} else {
		st.ConfigValid = true
	}
	return st
}

// Summary haalt de live statistieken op bij XMRig.
func (m *Manager) Summary(ctx context.Context) (*xmrig.Summary, error) {
	m.mu.Lock()
	client := m.client
	m.mu.Unlock()
	return client.FetchSummary(ctx)
}

// RecentLogs geeft de laatst opgevangen regels van de miner terug.
func (m *Manager) RecentLogs() []string {
	return m.logs.lines()
}

// supervise houdt het proces in leven zolang desired==true.
func (m *Manager) supervise() {
	for {
		m.mu.Lock()
		if !m.desired {
			m.state = StateStopped
			m.supervising = false
			m.mu.Unlock()
			return
		}
		cfg := m.cfg
		m.mu.Unlock()

		if err := cfg.Validate(); err != nil {
			m.setError(err.Error())
			if !m.sleepInterruptible(3 * time.Second) {
				m.finish()
				return
			}
			continue
		}

		m.runOnce(cfg)

		m.mu.Lock()
		desired := m.desired
		m.mu.Unlock()
		if !desired {
			m.finish()
			return
		}

		// Onverwacht gestopt: tel een herstart en wacht even.
		m.mu.Lock()
		m.restartCount++
		m.state = StateRestarting
		delay := time.Duration(m.cfg.RestartDelaySeconds) * time.Second
		m.mu.Unlock()
		m.logs.add("[agent] miner gestopt, herstart over " + delay.String())
		if !m.sleepInterruptible(delay) {
			m.finish()
			return
		}
	}
}

// runOnce start het proces één keer en blokkeert tot het stopt.
func (m *Manager) runOnce(cfg config.Config) {
	args := buildArgs(cfg)
	cmd := exec.Command(cfg.MinerPath, args...)
	// Eigen procesgroep zodt we de miner (en eventuele kinderen) netjes kunnen
	// beëindigen.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	m.mu.Lock()
	m.state = StateStarting
	m.mu.Unlock()
	m.logs.add("[agent] start: " + cfg.MinerPath + " " + redactArgs(args))

	if err := cmd.Start(); err != nil {
		m.setError("kon miner niet starten: " + err.Error())
		return
	}

	m.mu.Lock()
	m.cmd = cmd
	m.startedAt = time.Now()
	m.state = StateRunning
	m.lastError = ""
	m.mu.Unlock()

	go m.pump(stdout)
	go m.pump(stderr)

	// Watchdog op basis van de XMRig API: reageert de miner een tijd lang niet,
	// dan beëindigen we hem zodat de supervisor herstart.
	healthDone := make(chan struct{})
	go m.watchHealth(cfg, healthDone)

	err := cmd.Wait()
	close(healthDone)

	m.mu.Lock()
	m.cmd = nil
	if err != nil {
		m.lastExit = err.Error()
	} else {
		m.lastExit = "normaal afgesloten"
	}
	m.mu.Unlock()
}

// watchHealth beëindigt het proces als XMRig's API te lang onbereikbaar is.
func (m *Manager) watchHealth(cfg config.Config, done <-chan struct{}) {
	grace := time.Duration(cfg.HealthGraceSeconds) * time.Second
	start := time.Now()
	lastHealthy := time.Now()
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			_, err := m.client.FetchSummary(ctx)
			cancel()
			if err == nil {
				lastHealthy = time.Now()
				continue
			}
			// Geef de miner na het starten wat tijd om de API te openen.
			if time.Since(start) < grace {
				continue
			}
			if time.Since(lastHealthy) > grace {
				m.logs.add("[agent] watchdog: miner reageert niet, wordt herstart")
				m.mu.Lock()
				m.lastError = "watchdog: miner reageerde niet en is herstart"
				cmd := m.cmd
				m.mu.Unlock()
				killProcess(cmd)
				return
			}
		}
	}
}

func (m *Manager) pump(r io.Reader) {
	if r == nil {
		return
	}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 256*1024)
	for scanner.Scan() {
		m.logs.add(scanner.Text())
	}
}

func (m *Manager) setError(msg string) {
	m.mu.Lock()
	m.state = StateError
	m.lastError = msg
	m.mu.Unlock()
	m.logs.add("[agent] FOUT: " + msg)
}

func (m *Manager) finish() {
	m.mu.Lock()
	m.state = StateStopped
	m.supervising = false
	m.mu.Unlock()
	m.logs.add("[agent] gestopt")
}

func (m *Manager) sleepInterruptible(d time.Duration) (ok bool) {
	// Wacht d, maar stop direct als desired op false gaat.
	const step = 200 * time.Millisecond
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		m.mu.Lock()
		desired := m.desired
		m.mu.Unlock()
		if !desired {
			return false
		}
		time.Sleep(step)
	}
	m.mu.Lock()
	desired := m.desired
	m.mu.Unlock()
	return desired
}

// buildArgs zet de configuratie om in XMRig command line argumenten.
func buildArgs(cfg config.Config) []string {
	pool := cfg.Pool
	args := []string{
		"-a", cfg.Algo,
		"-o", pool,
		"-u", cfg.MinerUsername(),
		"-p", cfg.Password,
		"-k",
		"--http-host", cfg.MinerAPIHost,
		"--http-port", strconv.Itoa(cfg.MinerAPIPort),
		"--no-color",
	}
	if cfg.TLS {
		args = append(args, "--tls")
	}
	if cfg.Threads > 0 {
		args = append(args, "-t", strconv.Itoa(cfg.Threads))
	}
	args = append(args, cfg.ExtraArgs...)
	return args
}
