//go:build !windows
// +build !windows

package daemon

import (
	"os/exec"
	"syscall"
)

// setSysProcAttr sets platform-specific process attributes for daemonization
func setSysProcAttr(cmd *exec.Cmd) {
	// On Unix-like systems, create a new session to detach from parent
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // Create new session (detach from controlling terminal)
	}
}
