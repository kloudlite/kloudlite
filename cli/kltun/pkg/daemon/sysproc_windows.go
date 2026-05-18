//go:build windows
// +build windows

package daemon

import "os/exec"

// setSysProcAttr sets platform-specific process attributes for daemonization
func setSysProcAttr(cmd *exec.Cmd) {
	// Windows doesn't need special process attributes for daemonization
	// The process will run independently when started
}
