package wireguard

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"golang.zx2c4.com/wireguard/conn"
	"golang.zx2c4.com/wireguard/device"
	"golang.zx2c4.com/wireguard/tun"
)

// Config represents WireGuard configuration
type Config struct {
	// InterfaceName is the name of the TUN interface
	InterfaceName string

	// ListenPort is the UDP port WireGuard listens on
	ListenPort int

	// ConfigFile is the path to WireGuard configuration file (optional)
	ConfigFile string

	// MTU is the maximum transmission unit
	MTU int

	// Logger is the device logger
	Logger *device.Logger
}

// Device represents a running WireGuard device
type Device struct {
	tunDevice tun.Device
	wgDevice  *device.Device
	logger    *device.Logger
	config    *Config
}

// NewDevice creates a new WireGuard device
func NewDevice(ctx context.Context, config *Config) (*Device, error) {
	if config.InterfaceName == "" {
		config.InterfaceName = getDefaultInterfaceName()
	}

	if config.MTU == 0 {
		config.MTU = device.DefaultMTU
	}

	if config.Logger == nil {
		config.Logger = device.NewLogger(
			device.LogLevelVerbose,
			fmt.Sprintf("[WireGuard:%s] ", config.InterfaceName),
		)
	}

	// Create TUN device
	tunDevice, err := tun.CreateTUN(config.InterfaceName, config.MTU)
	if err != nil {
		return nil, fmt.Errorf("failed to create TUN device: %w", err)
	}

	// Get the actual interface name (may differ from requested on some platforms)
	actualName, err := tunDevice.Name()
	if err != nil {
		tunDevice.Close()
		return nil, fmt.Errorf("failed to get TUN device name: %w", err)
	}

	if actualName != config.InterfaceName {
		config.Logger.Verbosef("TUN interface created with name: %s (requested: %s)", actualName, config.InterfaceName)
		config.InterfaceName = actualName
	}

	d := &Device{
		tunDevice: tunDevice,
		config:    config,
		logger:    config.Logger,
	}

	return d, nil
}

// Start starts the WireGuard device
func (d *Device) Start(ctx context.Context) error {
	d.logger.Verbosef("Creating TUN interface: %s", d.config.InterfaceName)

	// Create WireGuard device
	d.logger.Verbosef("Starting WireGuard device")
	wgDevice := device.NewDevice(d.tunDevice, conn.NewDefaultBind(), d.logger)
	d.wgDevice = wgDevice

	// Load configuration if provided
	if d.config.ConfigFile != "" {
		d.logger.Verbosef("Loading configuration from: %s", d.config.ConfigFile)
		if err := d.LoadConfig(d.config.ConfigFile); err != nil {
			d.Close()
			return fmt.Errorf("failed to load config: %w", err)
		}
	}

	// Start device
	d.logger.Verbosef("WireGuard device is up")

	// Wait for context cancellation
	go func() {
		<-ctx.Done()
		d.logger.Verbosef("Shutting down WireGuard device")
		d.Close()
	}()

	return nil
}

// LoadConfig loads WireGuard configuration from a file
func (d *Device) LoadConfig(configFile string) error {
	if d.wgDevice == nil {
		return fmt.Errorf("device not started")
	}

	configData, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Set configuration using UAPI/IPC format
	// Config should already be in IPC format from the server
	if err := d.wgDevice.IpcSet(string(configData)); err != nil {
		return fmt.Errorf("failed to set configuration: %w", err)
	}

	d.logger.Verbosef("Configuration loaded successfully")
	return nil
}

// SetConfig sets WireGuard configuration from string (INI or IPC format)
func (d *Device) SetConfig(config string) error {
	if d.wgDevice == nil {
		return fmt.Errorf("device not started")
	}

	// Convert INI format to IPC format if needed
	ipcConfig := convertINIToIPC(config)

	if err := d.wgDevice.IpcSet(ipcConfig); err != nil {
		return fmt.Errorf("failed to set configuration: %w", err)
	}

	d.logger.Verbosef("Configuration applied successfully")
	return nil
}

// convertINIToIPC converts WireGuard INI config format to IPC format
// INI format: [Interface]/[Peer] sections with Key = Value
// IPC format: key=value\n pairs (peer sections start with public_key=)
func convertINIToIPC(config string) string {
	// If it doesn't look like INI format, return as-is
	if !strings.Contains(config, "[Interface]") && !strings.Contains(config, "[Peer]") {
		return config
	}

	var result strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(config))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip section headers
		if line == "[Interface]" || line == "[Peer]" {
			continue
		}

		// Parse key = value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Map INI keys to IPC keys
		ipcKey := mapINIKeyToIPC(key)
		if ipcKey == "" {
			continue // Skip unsupported keys like Address
		}

		result.WriteString(ipcKey)
		result.WriteString("=")
		result.WriteString(value)
		result.WriteString("\n")
	}

	return result.String()
}

// mapINIKeyToIPC maps INI config keys to IPC format keys
func mapINIKeyToIPC(key string) string {
	switch key {
	case "PrivateKey":
		return "private_key"
	case "PublicKey":
		return "public_key"
	case "ListenPort":
		return "listen_port"
	case "AllowedIPs":
		return "allowed_ip"
	case "Endpoint":
		return "endpoint"
	case "PersistentKeepalive":
		return "persistent_keepalive_interval"
	case "PresharedKey":
		return "preshared_key"
	// These are not part of IPC format (handled by network stack)
	case "Address", "DNS", "MTU", "Table", "PreUp", "PostUp", "PreDown", "PostDown":
		return ""
	default:
		return ""
	}
}

// GetConfig retrieves the current WireGuard configuration
func (d *Device) GetConfig() (string, error) {
	if d.wgDevice == nil {
		return "", fmt.Errorf("device not started")
	}

	config, err := d.wgDevice.IpcGet()
	if err != nil {
		return "", fmt.Errorf("failed to get configuration: %w", err)
	}

	return config, nil
}

// InterfaceName returns the actual TUN interface name
func (d *Device) InterfaceName() string {
	return d.config.InterfaceName
}

// Close shuts down the WireGuard device
func (d *Device) Close() error {
	if d.wgDevice != nil {
		d.wgDevice.Close()
		d.wgDevice = nil
	}

	if d.tunDevice != nil {
		if err := d.tunDevice.Close(); err != nil {
			log.Printf("Error closing TUN device: %v", err)
			return err
		}
		d.tunDevice = nil
	}

	return nil
}

// Wait waits for the device to be closed
func (d *Device) Wait() {
	if d.wgDevice != nil {
		d.wgDevice.Wait()
	}
}

// getDefaultInterfaceName returns a platform-specific default interface name
func getDefaultInterfaceName() string {
	switch runtime.GOOS {
	case "darwin":
		return "utun"
	case "linux":
		return "wg0"
	case "windows":
		return "Kloudlite"
	default:
		return "wg0"
	}
}
