//go:build windows

package hosts

// NewManager creates a new platform-specific manager
func NewManager() Manager {
	return NewWindowsManager()
}
