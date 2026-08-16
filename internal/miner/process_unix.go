//go:build !windows

package miner

import (
	"os/exec"
	"syscall"
	"time"
)

// setProcAttr zet het kindproces in zijn eigen procesgroep, zodat killProcess
// straks de hele groep (miner + eventuele kinderen) in één keer kan raken.
func setProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// killProcess beëindigt het proces (en zijn procesgroep) netjes met SIGTERM en
// escaleert na een korte periode naar SIGKILL. Een nil-proces wordt genegeerd.
//
// Bewust wordt hier NIET op het proces gewacht: de supervisor (runOnce) roept
// als enige cmd.Wait() aan om het kind te reapen. Dubbel wachten zou races
// geven.
func killProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	pid := cmd.Process.Pid
	// Negatieve pid = hele procesgroep (we starten met Setpgid).
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		return // proces bestaat niet (meer)
	}
	// Reageert het proces niet op SIGTERM, dan na 8s hard afsluiten. Is het
	// proces al weg, dan levert SIGKILL simpelweg ESRCH op (onschadelijk).
	time.AfterFunc(8*time.Second, func() {
		_ = syscall.Kill(-pid, syscall.SIGKILL)
	})
}
