//go:build !windows

package execenv

import "os/exec"

func HideWindow(*exec.Cmd) {}
