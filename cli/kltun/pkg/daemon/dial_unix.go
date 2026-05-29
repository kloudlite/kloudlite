//go:build !windows

package daemon

import (
	"net"
	"time"
)

// DialDaemon connects to the daemon using platform-specific method
// On Unix-like systems, this connects to a Unix socket
func DialDaemon(socketPath string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("unix", socketPath, timeout)
}
