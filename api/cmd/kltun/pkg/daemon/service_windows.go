//go:build windows

package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	// ServiceName is the name of the Windows service
	ServiceName = "kltund"

	// ServiceDisplayName is the display name of the Windows service
	ServiceDisplayName = "Kloudlite Tunnel Daemon"

	// ServiceDescription is the description of the Windows service
	ServiceDescription = "Daemon service for Kloudlite tunnel management"

	// SocketPath is the path to the named pipe
	SocketPath = `\\.\pipe\kltund`
)

// ServiceManager manages the daemon service on Windows
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
	// Open service manager
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	// Check if service already exists
	s, err := m.OpenService(ServiceName)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", ServiceName)
	}

	// Create service
	s, err = m.CreateService(ServiceName, sm.executablePath, mgr.Config{
		DisplayName: ServiceDisplayName,
		Description: ServiceDescription,
		StartType:   mgr.StartAutomatic,
	}, "daemon", "run")
	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}
	defer s.Close()

	// Start the service
	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	fmt.Println("Daemon service installed successfully")
	return nil
}

// Uninstall uninstalls the daemon service
func (sm *ServiceManager) Uninstall() error {
	// Open service manager
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	// Open service
	s, err := m.OpenService(ServiceName)
	if err != nil {
		return fmt.Errorf("service %s is not installed", ServiceName)
	}
	defer s.Close()

	// Stop the service
	status, err := s.Control(svc.Stop)
	if err != nil {
		// Continue even if stop fails
		fmt.Printf("Warning: failed to stop service: %v\n", err)
	} else {
		// Wait for service to stop
		timeout := time.Now().Add(10 * time.Second)
		for status.State != svc.Stopped {
			if time.Now().After(timeout) {
				return fmt.Errorf("timeout waiting for service to stop")
			}
			time.Sleep(300 * time.Millisecond)
			status, err = s.Query()
			if err != nil {
				return fmt.Errorf("failed to query service status: %w", err)
			}
		}
	}

	// Delete the service
	if err := s.Delete(); err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	fmt.Println("Daemon service uninstalled successfully")
	return nil
}

// Start starts the daemon service
func (sm *ServiceManager) Start() error {
	// Open service manager
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	// Open service
	s, err := m.OpenService(ServiceName)
	if err != nil {
		return fmt.Errorf("service %s is not installed", ServiceName)
	}
	defer s.Close()

	// Start the service
	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start service: %w", err)
	}

	fmt.Println("Daemon service started successfully")
	return nil
}

// Stop stops the daemon service
func (sm *ServiceManager) Stop() error {
	// Open service manager
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	// Open service
	s, err := m.OpenService(ServiceName)
	if err != nil {
		return fmt.Errorf("service %s is not installed", ServiceName)
	}
	defer s.Close()

	// Stop the service
	status, err := s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("failed to stop service: %w", err)
	}

	// Wait for service to stop
	timeout := time.Now().Add(10 * time.Second)
	for status.State != svc.Stopped {
		if time.Now().After(timeout) {
			return fmt.Errorf("timeout waiting for service to stop")
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("failed to query service status: %w", err)
		}
	}

	fmt.Println("Daemon service stopped successfully")
	return nil
}

// Status returns the status of the daemon service
func (sm *ServiceManager) Status() (bool, error) {
	// Open service manager
	m, err := mgr.Connect()
	if err != nil {
		return false, fmt.Errorf("failed to connect to service manager: %w", err)
	}
	defer m.Disconnect()

	// Open service
	s, err := m.OpenService(ServiceName)
	if err != nil {
		// Service not installed
		return false, nil
	}
	defer s.Close()

	// Query service status
	status, err := s.Query()
	if err != nil {
		return false, fmt.Errorf("failed to query service status: %w", err)
	}

	return status.State == svc.Running, nil
}

// IsInstalled checks if the daemon service is installed
func (sm *ServiceManager) IsInstalled() bool {
	m, err := mgr.Connect()
	if err != nil {
		return false
	}
	defer m.Disconnect()

	s, err := m.OpenService(ServiceName)
	if err != nil {
		return false
	}
	defer s.Close()

	return true
}

// GetSocketPath returns the socket path for the daemon
func (sm *ServiceManager) GetSocketPath() string {
	return SocketPath
}

// EnsureRunning ensures the daemon is running, installing and starting it if necessary
func (sm *ServiceManager) EnsureRunning() error {
	// Check if we need admin
	needsAdmin := false

	// Check if installed
	if !sm.IsInstalled() {
		needsAdmin = true
	}

	// Check if running
	running, err := sm.Status()
	if err != nil {
		return err
	}

	if !running {
		needsAdmin = true
	}

	// If we need admin and we're not admin, re-exec with elevation
	if needsAdmin && !isAdmin() {
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
	// Need admin to restart
	if !isAdmin() {
		return sm.escalateAndRestart()
	}

	// Stop first (ignore errors if not running)
	sm.Stop()

	// Start fresh
	return sm.Start()
}

// escalateAndRestart escalates privileges and restarts the daemon
func (sm *ServiceManager) escalateAndRestart() error {
	fmt.Println("Restarting kltun daemon...")

	// Use PowerShell to request elevation for restart
	cmd := exec.Command("powershell", "-Command", "Start-Process", "-Verb", "RunAs", "-FilePath", sm.executablePath, "-ArgumentList", "daemon stop; Start-Sleep -Seconds 1; daemon start", "-Wait")
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

	// Use PowerShell to request elevation
	cmd := exec.Command("powershell", "-Command", "Start-Process", "-Verb", "RunAs", "-FilePath", sm.executablePath, "-ArgumentList", "daemon install")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install daemon with elevated privileges: %w", err)
	}

	return nil
}

// isAdmin checks if the current process is running as administrator
func isAdmin() bool {
	// This is a simplified check - in production you'd use Windows API
	cmd := exec.Command("net", "session")
	err := cmd.Run()
	return err == nil
}

// windowsService implements svc.Handler
type windowsService struct {
	server *Server
}

// Execute implements svc.Handler
func (ws *windowsService) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	s <- svc.Status{State: svc.StartPending}

	// Start the server
	go func() {
		if err := ws.server.Start(SocketPath); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	s <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	// Wait for stop/shutdown signal
	for {
		c := <-r
		switch c.Cmd {
		case svc.Interrogate:
			s <- c.CurrentStatus
		case svc.Stop, svc.Shutdown:
			s <- svc.Status{State: svc.StopPending}
			ws.server.Stop()
			return false, 0
		}
	}
}

// RunAsService runs the daemon as a Windows service
func RunAsService() error {
	server, err := NewServer()
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	ws := &windowsService{server: server}

	if err := svc.Run(ServiceName, ws); err != nil {
		return fmt.Errorf("failed to run service: %w", err)
	}

	return nil
}
