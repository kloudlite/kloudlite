package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// Client represents an RPC client
type Client struct {
	socketPath string
	timeout    time.Duration
}

// NewClient creates a new RPC client
func NewClient(socketPath string) *Client {
	return &Client{
		socketPath: socketPath,
		timeout:    30 * time.Second,
	}
}

// SetTimeout sets the client timeout
func (c *Client) SetTimeout(timeout time.Duration) {
	c.timeout = timeout
}

// startDaemon attempts to start the daemon if it's not running
func (c *Client) startDaemon() error {
	// Find kltun binary (look in same directory as current executable or in PATH)
	kltunPath, err := exec.LookPath("kltun")
	if err != nil {
		// Try to find it relative to current executable
		exe, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exe)
			kltunPath = filepath.Join(exeDir, "kltun")
			if _, err := os.Stat(kltunPath); err != nil {
				return fmt.Errorf("kltun binary not found in PATH or %s", exeDir)
			}
		} else {
			return fmt.Errorf("kltun binary not found: %w", err)
		}
	}

	fmt.Printf("Starting daemon with: sudo %s daemon run\n", kltunPath)

	// Create command with proper daemonization
	cmd := exec.Command("sudo", kltunPath, "daemon", "run")

	// Discard output to prevent blocking
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	// Platform-specific process attributes for proper daemonization
	setSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Release the process (don't wait for it)
	go cmd.Wait()

	// Wait for daemon to be ready (max 5 seconds)
	for i := 0; i < 50; i++ {
		time.Sleep(100 * time.Millisecond)
		if c.IsRunning() {
			fmt.Println("✓ Daemon started successfully")
			return nil
		}
	}

	return fmt.Errorf("daemon failed to start within 5 seconds")
}

// call makes an RPC call
func (c *Client) call(method string, params interface{}, result interface{}) error {
	// Connect to daemon
	conn, err := net.DialTimeout("unix", c.socketPath, c.timeout)
	if err != nil {
		// Check if daemon is not running and try to start it
		if _, statErr := os.Stat(c.socketPath); os.IsNotExist(statErr) {
			fmt.Println("Daemon not running, attempting to start...")
			if startErr := c.startDaemon(); startErr != nil {
				return fmt.Errorf("failed to connect to daemon: %w (auto-start failed: %v)", err, startErr)
			}
			// Retry connection after starting daemon
			conn, err = net.DialTimeout("unix", c.socketPath, c.timeout)
			if err != nil {
				return fmt.Errorf("failed to connect to daemon after auto-start: %w", err)
			}
		} else {
			return fmt.Errorf("failed to connect to daemon: %w", err)
		}
	}
	defer conn.Close()

	// Set deadline
	if err := conn.SetDeadline(time.Now().Add(c.timeout)); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}

	// Create request
	req, err := NewRequest(uuid.New().String(), method, params)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Send request
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(req); err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Receive response
	var resp Response
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&resp); err != nil {
		return fmt.Errorf("failed to receive response: %w", err)
	}

	// Check for errors
	if resp.Error != nil {
		return fmt.Errorf("RPC error %d: %s (data: %s)", resp.Error.Code, resp.Error.Message, resp.Error.Data)
	}

	// Unmarshal result
	if result != nil {
		if err := json.Unmarshal(resp.Result, result); err != nil {
			return fmt.Errorf("failed to unmarshal result: %w", err)
		}
	}

	return nil
}

// Ping pings the daemon
func (c *Client) Ping() error {
	var result PingResult
	if err := c.call(MethodPing, PingParams{}, &result); err != nil {
		return err
	}

	if result.Message != "pong" {
		return fmt.Errorf("unexpected ping response: %s", result.Message)
	}

	return nil
}

// InstallCA installs a CA certificate
func (c *Client) InstallCA(certPath string) error {
	params := InstallCAParams{CertPath: certPath}
	var result InstallCAResult

	if err := c.call(MethodInstallCA, params, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	return nil
}

// UninstallCA uninstalls a CA certificate
func (c *Client) UninstallCA(certPath string) error {
	params := UninstallCAParams{CertPath: certPath}
	var result UninstallCAResult

	if err := c.call(MethodUninstallCA, params, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	return nil
}

// HostsAdd adds a host entry
func (c *Client) HostsAdd(hostname, ip, comment string) error {
	params := HostsAddParams{
		Hostname: hostname,
		IP:       ip,
		Comment:  comment,
	}
	var result HostsAddResult

	if err := c.call(MethodHostsAdd, params, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	return nil
}

// HostsRemove removes a host entry
func (c *Client) HostsRemove(hostname string) error {
	params := HostsRemoveParams{Hostname: hostname}
	var result HostsRemoveResult

	if err := c.call(MethodHostsRemove, params, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	return nil
}

// HostsList lists all host entries
func (c *Client) HostsList() ([]HostsEntry, error) {
	var result HostsListResult

	if err := c.call(MethodHostsList, HostsListParams{}, &result); err != nil {
		return nil, err
	}

	return result.Entries, nil
}

// HostsSync synchronizes hosts
func (c *Client) HostsSync() error {
	var result HostsSyncResult

	if err := c.call(MethodHostsSync, HostsSyncParams{}, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	return nil
}

// HostsClean cleans all host entries
func (c *Client) HostsClean() error {
	var result HostsCleanResult

	if err := c.call(MethodHostsClean, HostsCleanParams{}, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	return nil
}

// HostsFlush flushes DNS cache
func (c *Client) HostsFlush() error {
	var result HostsFlushResult

	if err := c.call(MethodHostsFlush, HostsFlushParams{}, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	return nil
}

// VPNConnect starts a VPN connection
func (c *Client) VPNConnect(token, server string) (string, error) {
	params := VPNConnectParams{
		Token:  token,
		Server: server,
	}
	var result VPNConnectResult

	if err := c.call(MethodVPNConnect, params, &result); err != nil {
		return "", err
	}

	if !result.Success {
		return "", fmt.Errorf("%s", result.Message)
	}

	return result.SessionID, nil
}

// VPNQuit stops the active VPN connection
func (c *Client) VPNQuit() error {
	var result VPNQuitResult

	if err := c.call(MethodVPNQuit, VPNQuitParams{}, &result); err != nil {
		return err
	}

	if !result.Success {
		return fmt.Errorf("%s", result.Message)
	}

	return nil
}

// Status gets daemon status
func (c *Client) Status() (*StatusResult, error) {
	var result StatusResult

	if err := c.call(MethodStatus, StatusParams{}, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// IsRunning checks if the daemon is running
func (c *Client) IsRunning() bool {
	return c.Ping() == nil
}
