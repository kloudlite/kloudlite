package netconfig

import (
	"fmt"
	"runtime"
)

// InterfaceConfig represents network interface configuration
type InterfaceConfig struct {
	// InterfaceName is the name of the network interface
	InterfaceName string

	// IPAddress is the IP address with CIDR (e.g., "10.17.0.2/24")
	IPAddress string

	// Routes is a list of routes to add (e.g., ["10.17.0.0/24", "10.43.0.0/16"])
	Routes []string

	// Gateway is the default gateway IP (optional)
	Gateway string
}

// ConfigureInterface configures IP address and routes on a network interface
func ConfigureInterface(config *InterfaceConfig) error {
	if config.InterfaceName == "" {
		return fmt.Errorf("interface name is required")
	}

	if config.IPAddress == "" {
		return fmt.Errorf("IP address is required")
	}

	switch runtime.GOOS {
	case "darwin":
		return configureDarwin(config)
	case "linux":
		return configureLinux(config)
	case "windows":
		return configureWindows(config)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// RemoveInterface removes IP address and routes from a network interface
func RemoveInterface(config *InterfaceConfig) error {
	if config.InterfaceName == "" {
		return fmt.Errorf("interface name is required")
	}

	switch runtime.GOOS {
	case "darwin":
		return removeDarwin(config)
	case "linux":
		return removeLinux(config)
	case "windows":
		return removeWindows(config)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}
