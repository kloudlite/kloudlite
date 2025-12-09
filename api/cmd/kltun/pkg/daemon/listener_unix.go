//go:build !windows

package daemon

import (
	"net"
	"os"
)

// CreateListener creates a platform-specific listener
// On Unix-like systems, this creates a Unix socket listener
func CreateListener(socketPath string) (net.Listener, error) {
	// Remove existing socket if it exists
	if err := os.RemoveAll(socketPath); err != nil {
		return nil, err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, err
	}

	// Set socket permissions (allow all users to connect)
	if err := os.Chmod(socketPath, 0o666); err != nil {
		listener.Close()
		return nil, err
	}

	return listener, nil
}
