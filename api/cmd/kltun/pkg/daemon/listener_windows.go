//go:build windows

package daemon

import (
	"net"

	"github.com/Microsoft/go-winio"
)

// CreateListener creates a platform-specific listener
// On Windows, this creates a named pipe listener
func CreateListener(socketPath string) (net.Listener, error) {
	// Use go-winio for Windows named pipes
	return winio.ListenPipe(socketPath, nil)
}
