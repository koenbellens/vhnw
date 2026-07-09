package miner

import (
	"strings"
	"sync"
	"time"
)

// ringBuffer bewaart de laatste N logregels van de miner, thread-safe.
type ringBuffer struct {
	mu   sync.Mutex
	buf  []string
	size int
}

func newRingBuffer(size int) *ringBuffer {
	return &ringBuffer{size: size}
}

func (r *ringBuffer) add(line string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	stamped := time.Now().Format("15:04:05") + " " + line
	r.buf = append(r.buf, stamped)
	if len(r.buf) > r.size {
		r.buf = r.buf[len(r.buf)-r.size:]
	}
}

func (r *ringBuffer) lines() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.buf))
	copy(out, r.buf)
	return out
}

// redactArgs verbergt de wallet/username in loggeregels zodat een
// schermafdruk van de UI niet meteen je volledige adres prijsgeeft.
func redactArgs(args []string) string {
	out := make([]string, len(args))
	copy(out, args)
	for i := 0; i < len(out)-1; i++ {
		if out[i] == "-u" {
			out[i+1] = maskSecret(out[i+1])
		}
	}
	return strings.Join(out, " ")
}

func maskSecret(s string) string {
	if len(s) <= 8 {
		return "***"
	}
	return s[:4] + "..." + s[len(s)-4:]
}
