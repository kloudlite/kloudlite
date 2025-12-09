//go:build !windows

package wintun

// EnsureAvailable is a no-op on non-Windows platforms.
// WireGuard uses native TUN devices on Linux/macOS.
func EnsureAvailable() error {
	return nil
}
