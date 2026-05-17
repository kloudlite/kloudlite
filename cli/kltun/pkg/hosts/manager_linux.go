//go:build linux

package hosts

// NewManager creates a new platform-specific manager
func NewManager() Manager {
	return NewLinuxManager()
}
