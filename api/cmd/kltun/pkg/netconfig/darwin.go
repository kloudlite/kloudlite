package netconfig

import (
	"fmt"
	"os/exec"
	"strings"
)

// configureDarwin configures network interface on macOS using ifconfig and route
func configureDarwin(config *InterfaceConfig) error {
	// Parse IP address and netmask from CIDR
	parts := strings.Split(config.IPAddress, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid IP address format, expected CIDR notation (e.g., 10.17.0.2/24)")
	}
	ipAddr := parts[0]

	// Determine gateway (if not provided, use first IP in subnet)
	gateway := config.Gateway
	if gateway == "" {
		// For 10.17.0.2/24, gateway is typically 10.17.0.1
		ipParts := strings.Split(ipAddr, ".")
		if len(ipParts) == 4 {
			gateway = fmt.Sprintf("%s.%s.%s.1", ipParts[0], ipParts[1], ipParts[2])
		}
	}

	// Configure IP address on interface
	// ifconfig <interface> inet <ip> <gateway> up
	cmd := exec.Command("ifconfig", config.InterfaceName, "inet", ipAddr, gateway, "up")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to configure IP address: %w\nOutput: %s", err, string(output))
	}

	// Add routes for each specified network
	for _, routeNet := range config.Routes {
		// route -n add -net <network> -interface <interface>
		cmd := exec.Command("route", "-n", "add", "-net", routeNet, "-interface", config.InterfaceName)
		if output, err := cmd.CombinedOutput(); err != nil {
			// Log warning but don't fail - route might already exist
			fmt.Printf("Warning: failed to add route %s: %v\nOutput: %s\n", routeNet, err, string(output))
		}
	}

	return nil
}

// removeDarwin removes network configuration from interface on macOS
func removeDarwin(config *InterfaceConfig) error {
	// Remove routes
	for _, routeNet := range config.Routes {
		cmd := exec.Command("route", "-n", "delete", "-net", routeNet)
		_ = cmd.Run() // Ignore errors - route might not exist
	}

	// Down the interface
	cmd := exec.Command("ifconfig", config.InterfaceName, "down")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to bring down interface: %w\nOutput: %s", err, string(output))
	}

	return nil
}
