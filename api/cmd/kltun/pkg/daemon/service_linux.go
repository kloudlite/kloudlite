//go:build linux

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const (
	// SystemdServiceName is the name of the systemd service
	SystemdServiceName = "kltund"

	// SystemdServicePath is the path to the systemd service file
	SystemdServicePath = "/etc/systemd/system/kltund.service"

	// SocketPath is the path to the Unix socket
	SocketPath = "/var/run/kltund.sock"
)

// systemdServiceTemplate is the template for the systemd service file
const systemdServiceTemplate = `[Unit]
Description=Kloudlite Tunnel Daemon
After=network.target

[Service]
Type=simple
ExecStart=sh -c '{{.ExecutablePath}} daemon run'
Restart=on-failure
RestartSec=5
StandardOutput=journal
StandardError=journal
User=root
Group=root

[Install]
WantedBy=multi-user.target
`

// ServiceManager manages the daemon service on Linux
type ServiceManager struct {
	executablePath string
}

// NewServiceManager creates a new service manager
func NewServiceManager() (*ServiceManager, error) {
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("failed to get executable path: %w", err)
	}

	// Resolve symlinks
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve executable path: %w", err)
	}

	return &ServiceManager{
		executablePath: execPath,
	}, nil
}

// Install installs the daemon service
func (sm *ServiceManager) Install() error {
	// Check if running as root
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root to install daemon service")
	}

	// Create service file
	tmpl, err := template.New("service").Parse(systemdServiceTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse service template: %w", err)
	}

	file, err := os.Create(SystemdServicePath)
	if err != nil {
		return fmt.Errorf("failed to create service file: %w", err)
	}
	defer file.Close()

	data := struct {
		ExecutablePath string
	}{
		ExecutablePath: sm.executablePath,
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to write service file: %w", err)
	}

	// Set permissions
	if err := os.Chmod(SystemdServicePath, 0644); err != nil {
		return fmt.Errorf("failed to set service permissions: %w", err)
	}

	// Reload systemd
	cmd := exec.Command("systemctl", "daemon-reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w\nOutput: %s", err, output)
	}

	// Enable the service
	cmd = exec.Command("systemctl", "enable", SystemdServiceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to enable service: %w\nOutput: %s", err, output)
	}

	// Start the service
	cmd = exec.Command("systemctl", "start", SystemdServiceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start service: %w\nOutput: %s", err, output)
	}

	fmt.Println("Daemon service installed successfully")
	return nil
}

// Uninstall uninstalls the daemon service
func (sm *ServiceManager) Uninstall() error {
	// Check if running as root
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root to uninstall daemon service")
	}

	// Stop the service
	cmd := exec.Command("systemctl", "stop", SystemdServiceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Continue even if stop fails (service might not be running)
		fmt.Printf("Warning: failed to stop service: %v\nOutput: %s\n", err, output)
	}

	// Disable the service
	cmd = exec.Command("systemctl", "disable", SystemdServiceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Continue even if disable fails
		fmt.Printf("Warning: failed to disable service: %v\nOutput: %s\n", err, output)
	}

	// Remove service file
	if err := os.Remove(SystemdServicePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove service file: %w", err)
	}

	// Reload systemd
	cmd = exec.Command("systemctl", "daemon-reload")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to reload systemd: %w\nOutput: %s", err, output)
	}

	fmt.Println("Daemon service uninstalled successfully")
	return nil
}

// Start starts the daemon service
func (sm *ServiceManager) Start() error {
	// Check if running as root
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root to start daemon service")
	}

	cmd := exec.Command("systemctl", "start", SystemdServiceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to start service: %w\nOutput: %s", err, output)
	}

	fmt.Println("Daemon service started successfully")
	return nil
}

// Stop stops the daemon service
func (sm *ServiceManager) Stop() error {
	// Check if running as root
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root to stop daemon service")
	}

	cmd := exec.Command("systemctl", "stop", SystemdServiceName)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop service: %w\nOutput: %s", err, output)
	}

	fmt.Println("Daemon service stopped successfully")
	return nil
}

// Status returns the status of the daemon service
func (sm *ServiceManager) Status() (bool, error) {
	cmd := exec.Command("systemctl", "is-active", SystemdServiceName)
	output, err := cmd.Output()

	if err != nil {
		// If the service is not active, is-active will return non-zero exit code
		return false, nil
	}

	// Check if output is "active"
	status := string(output)
	return status == "active\n", nil
}

// IsInstalled checks if the daemon service is installed
func (sm *ServiceManager) IsInstalled() bool {
	_, err := os.Stat(SystemdServicePath)
	return err == nil
}

// GetSocketPath returns the socket path for the daemon
func (sm *ServiceManager) GetSocketPath() string {
	return SocketPath
}

// EnsureRunning ensures the daemon is running, installing and starting it if necessary
func (sm *ServiceManager) EnsureRunning() error {
	// Check if we need root
	needsRoot := false

	// Check if installed
	if !sm.IsInstalled() {
		needsRoot = true
	}

	// Check if running
	running, err := sm.Status()
	if err != nil {
		return err
	}

	if !running {
		needsRoot = true
	}

	// If we need root and we're not root, re-exec with sudo
	if needsRoot && os.Geteuid() != 0 {
		return sm.escalateAndInstall()
	}

	// Install if not installed
	if !sm.IsInstalled() {
		if err := sm.Install(); err != nil {
			return err
		}
	}

	// Start if not running
	if !running {
		if err := sm.Start(); err != nil {
			return err
		}
	}

	return nil
}

// Restart stops and starts the daemon to reload tokens
func (sm *ServiceManager) Restart() error {
	// Need root to restart
	if os.Geteuid() != 0 {
		return sm.escalateAndRestart()
	}

	// Stop first (ignore errors if not running)
	sm.Stop()

	// Remove stale socket file
	os.Remove(sm.GetSocketPath())

	// Start fresh
	return sm.Start()
}

// escalateAndRestart escalates privileges and restarts the daemon
func (sm *ServiceManager) escalateAndRestart() error {
	fmt.Println("Restarting kltun daemon...")

	// Stop, remove socket, and start to reload tokens
	exec.Command("sudo", "systemctl", "stop", SystemdServiceName).Run()
	os.Remove(sm.GetSocketPath())

	cmd := exec.Command("sudo", "systemctl", "start", SystemdServiceName)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restart daemon: %w", err)
	}

	return nil
}

// escalateAndInstall escalates privileges and installs the daemon
func (sm *ServiceManager) escalateAndInstall() error {
	fmt.Println("Daemon is not running. Requesting administrator privileges to start it...")

	cmd := exec.Command("sudo", sm.executablePath, "daemon", "install")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install daemon with elevated privileges: %w", err)
	}

	return nil
}
