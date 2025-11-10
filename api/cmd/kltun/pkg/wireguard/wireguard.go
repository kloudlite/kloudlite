package wireguard

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"

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

	d := &Device{
		config: config,
		logger: config.Logger,
	}

	return d, nil
}

// Start starts the WireGuard device
func (d *Device) Start(ctx context.Context) error {
	d.logger.Verbosef("Creating TUN interface: %s", d.config.InterfaceName)

	// Create TUN device
	tunDevice, err := tun.CreateTUN(d.config.InterfaceName, d.config.MTU)
	if err != nil {
		return fmt.Errorf("failed to create TUN device: %w", err)
	}
	d.tunDevice = tunDevice

	// Get the actual interface name (may differ from requested on some platforms)
	actualName, err := tunDevice.Name()
	if err != nil {
		tunDevice.Close()
		return fmt.Errorf("failed to get TUN device name: %w", err)
	}

	if actualName != d.config.InterfaceName {
		d.logger.Verbosef("TUN interface created with name: %s (requested: %s)", actualName, d.config.InterfaceName)
		d.config.InterfaceName = actualName
	}

	// Create WireGuard device
	d.logger.Verbosef("Starting WireGuard device")
	wgDevice := device.NewDevice(tunDevice, conn.NewDefaultBind(), d.logger)
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

	// Set configuration using UAPI
	// IpcSetOperation expects an io.Reader
	if err := d.wgDevice.IpcSet(string(configData)); err != nil {
		return fmt.Errorf("failed to set configuration: %w", err)
	}

	d.logger.Verbosef("Configuration loaded successfully")
	return nil
}

// SetConfig sets WireGuard configuration from string
func (d *Device) SetConfig(config string) error {
	if d.wgDevice == nil {
		return fmt.Errorf("device not started")
	}

	if err := d.wgDevice.IpcSet(config); err != nil {
		return fmt.Errorf("failed to set configuration: %w", err)
	}

	d.logger.Verbosef("Configuration applied successfully")
	return nil
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
