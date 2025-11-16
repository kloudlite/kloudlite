//go:build darwin

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

const (
	// LaunchdLabel is the label for the launchd service
	LaunchdLabel = "io.kloudlite.kltund"

	// LaunchdPlistPath is the path to the launchd plist file
	LaunchdPlistPath = "/Library/LaunchDaemons/io.kloudlite.kltund.plist"

	// SocketPath is the path to the Unix socket
	SocketPath = "/var/run/kltund.sock"
)

// launchdPlistTemplate is the template for the launchd plist file
const launchdPlistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.ExecutablePath}}</string>
        <string>daemon</string>
        <string>run</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <dict>
        <key>SuccessfulExit</key>
        <false/>
    </dict>
    <key>StandardOutPath</key>
    <string>/var/log/kltund.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/kltund.error.log</string>
    <key>UserName</key>
    <string>root</string>
    <key>GroupName</key>
    <string>wheel</string>
</dict>
</plist>
`

// ServiceManager manages the daemon service on macOS
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

	// Create plist file
	tmpl, err := template.New("plist").Parse(launchdPlistTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse plist template: %w", err)
	}

	file, err := os.Create(LaunchdPlistPath)
	if err != nil {
		return fmt.Errorf("failed to create plist file: %w", err)
	}
	defer file.Close()

	data := struct {
		Label          string
		ExecutablePath string
	}{
		Label:          LaunchdLabel,
		ExecutablePath: sm.executablePath,
	}

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to write plist file: %w", err)
	}

	// Set permissions
	if err := os.Chmod(LaunchdPlistPath, 0644); err != nil {
		return fmt.Errorf("failed to set plist permissions: %w", err)
	}

	// Load the service
	cmd := exec.Command("launchctl", "load", LaunchdPlistPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to load service: %w\nOutput: %s", err, output)
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

	// Unload the service
	cmd := exec.Command("launchctl", "unload", LaunchdPlistPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Continue even if unload fails (service might not be loaded)
		fmt.Printf("Warning: failed to unload service: %v\nOutput: %s\n", err, output)
	}

	// Remove plist file
	if err := os.Remove(LaunchdPlistPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove plist file: %w", err)
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

	cmd := exec.Command("launchctl", "start", LaunchdLabel)
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

	cmd := exec.Command("launchctl", "stop", LaunchdLabel)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop service: %w\nOutput: %s", err, output)
	}

	fmt.Println("Daemon service stopped successfully")
	return nil
}

// Status returns the status of the daemon service
func (sm *ServiceManager) Status() (bool, error) {
	// The daemon runs as root in the system domain, so we need to check with sudo
	// First try without sudo (for non-root users)
	cmd := exec.Command("launchctl", "list", LaunchdLabel)
	output, err := cmd.CombinedOutput()

	if err == nil && len(output) > 0 {
		// Service found in user domain
		return true, nil
	}

	// Try with sudo to check system domain
	cmd = exec.Command("sudo", "-n", "launchctl", "list", LaunchdLabel)
	output, err = cmd.CombinedOutput()

	if err != nil {
		// If sudo -n fails due to password requirement, try checking process list
		// This is a fallback that doesn't require sudo
		cmd = exec.Command("pgrep", "-f", "kltun daemon run")
		output, err = cmd.CombinedOutput()
		if err == nil && len(output) > 0 {
			return true, nil
		}
		return false, nil
	}

	// If we got output, the service is loaded
	return len(output) > 0, nil
}

// IsInstalled checks if the daemon service is installed
func (sm *ServiceManager) IsInstalled() bool {
	_, err := os.Stat(LaunchdPlistPath)
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
