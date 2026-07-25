//go:build windows

package execenv

import (
	"os/exec"
	"syscall"
)

// HideWindow prevents GUI-launched child processes from opening a console window.
func HideWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
}
