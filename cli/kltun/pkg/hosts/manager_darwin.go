//go:build darwin

package hosts

// NewManager creates a new platform-specific manager
func NewManager() Manager {
	return NewDarwinManager()
}
