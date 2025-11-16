package netconfig

import (
	"fmt"
	"os/exec"
	"strings"
)

// configureWindows configures network interface on Windows using netsh
func configureWindows(config *InterfaceConfig) error {
	// Parse IP address and netmask from CIDR
	parts := strings.Split(config.IPAddress, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid IP address format, expected CIDR notation (e.g., 10.17.0.2/24)")
	}
	ipAddr := parts[0]
	// Convert CIDR to subnet mask (e.g., /24 -> 255.255.255.0)
	subnetMask := cidrToSubnetMask(parts[1])

	// Configure IP address on interface
	// netsh interface ip set address name="<interface>" static <ip> <mask>
	cmd := exec.Command("netsh", "interface", "ip", "set", "address",
		fmt.Sprintf("name=%s", config.InterfaceName),
		"static", ipAddr, subnetMask)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to configure IP address: %w\nOutput: %s", err, string(output))
	}

	// Add routes for each specified network
	for _, routeNet := range config.Routes {
		// Parse route network and prefix
		routeParts := strings.Split(routeNet, "/")
		if len(routeParts) != 2 {
			continue
		}
		routeIP := routeParts[0]
		routeMask := cidrToSubnetMask(routeParts[1])

		// route add <network> mask <mask> <gateway> if <interface_index>
		// For simplicity, use the IP address as next hop
		cmd := exec.Command("route", "add", routeIP, "mask", routeMask, ipAddr)
		if output, err := cmd.CombinedOutput(); err != nil {
			// Log warning but don't fail - route might already exist
			fmt.Printf("Warning: failed to add route %s: %v\nOutput: %s\n", routeNet, err, string(output))
		}
	}

	return nil
}

// removeWindows removes network configuration from interface on Windows
func removeWindows(config *InterfaceConfig) error {
	// Remove routes
	for _, routeNet := range config.Routes {
		routeParts := strings.Split(routeNet, "/")
		if len(routeParts) != 2 {
			continue
		}
		routeIP := routeParts[0]

		cmd := exec.Command("route", "delete", routeIP)
		_ = cmd.Run() // Ignore errors - route might not exist
	}

	// Reset interface to DHCP (effectively removes static IP)
	cmd := exec.Command("netsh", "interface", "ip", "set", "address",
		fmt.Sprintf("name=%s", config.InterfaceName), "dhcp")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reset interface: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// cidrToSubnetMask converts CIDR prefix length to subnet mask
func cidrToSubnetMask(cidr string) string {
	masks := map[string]string{
		"8":  "255.0.0.0",
		"16": "255.255.0.0",
		"24": "255.255.255.0",
		"32": "255.255.255.255",
	}

	if mask, ok := masks[cidr]; ok {
		return mask
	}
	// Default to /24
	return "255.255.255.0"
}
