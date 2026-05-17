//go:build windows

package daemon

import (
	"net"
	"time"

	"github.com/Microsoft/go-winio"
)

// DialDaemon connects to the daemon using platform-specific method
// On Windows, this connects to a named pipe
func DialDaemon(socketPath string, timeout time.Duration) (net.Conn, error) {
	return winio.DialPipe(socketPath, &timeout)
}
