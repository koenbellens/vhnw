//go:build windows

package miner

import "os/exec"

// setProcAttr doet op Windows niets bijzonders: er bestaat geen procesgroep-
// concept zoals op Unix. Eén xmrig-kindproces volstaat.
func setProcAttr(cmd *exec.Cmd) {}

// killProcess beëindigt het proces. Windows kent geen SIGTERM/SIGKILL-
// onderscheid zoals Unix; Kill() sluit het proces direct af (vergelijkbaar
// met SIGKILL). xmrig heeft geen nette shutdown-fase nodig om data te
// bewaren, dus een directe kill is hier prima.
func killProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Kill()
}
