package netconfig

import (
	"fmt"
	"os/exec"
)

// configureLinux configures network interface on Linux using ip command
func configureLinux(config *InterfaceConfig) error {
	// Configure IP address on interface
	// ip addr add <ip/cidr> dev <interface>
	cmd := exec.Command("ip", "addr", "add", config.IPAddress, "dev", config.InterfaceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to configure IP address: %w\nOutput: %s", err, string(output))
	}

	// Bring interface up
	cmd = exec.Command("ip", "link", "set", "dev", config.InterfaceName, "up")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to bring up interface: %w\nOutput: %s", err, string(output))
	}

	// Add routes for each specified network
	for _, routeNet := range config.Routes {
		// ip route add <network> dev <interface>
		cmd := exec.Command("ip", "route", "add", routeNet, "dev", config.InterfaceName)
		if output, err := cmd.CombinedOutput(); err != nil {
			// Log warning but don't fail - route might already exist
			fmt.Printf("Warning: failed to add route %s: %v\nOutput: %s\n", routeNet, err, string(output))
		}
	}

	return nil
}

// removeLinux removes network configuration from interface on Linux
func removeLinux(config *InterfaceConfig) error {
	// Remove routes
	for _, routeNet := range config.Routes {
		cmd := exec.Command("ip", "route", "del", routeNet, "dev", config.InterfaceName)
		_ = cmd.Run() // Ignore errors - route might not exist
	}

	// Remove IP address
	cmd := exec.Command("ip", "addr", "del", config.IPAddress, "dev", config.InterfaceName)
	_ = cmd.Run() // Ignore errors - might not be configured

	// Bring interface down
	cmd = exec.Command("ip", "link", "set", "dev", config.InterfaceName, "down")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to bring down interface: %w\nOutput: %s", err, string(output))
	}

	return nil
}
