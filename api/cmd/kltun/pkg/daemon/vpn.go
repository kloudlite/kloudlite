package daemon

import (
	"context"
	"fmt"
	"os"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/api"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/deviceid"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/netconfig"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/truststore"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/wireguard"

	"codeberg.org/eduVPN/proxyguard"
)

// runVPNConnectionWithResult runs a VPN connection and sends result on channel after initial setup
func (s *Server) runVPNConnectionWithResult(ctx context.Context, sessionID, server, token string, done chan struct{}, resultChan chan<- error) {
	defer close(done) // Signal completion when function returns
	fmt.Printf("[Session %s] Starting VPN connection\n", sessionID)

	// Create API client
	apiClient := api.NewClient(server, token)

	// Get or create device ID
	deviceID, err := deviceid.GetOrCreateDeviceID()
	if err != nil {
		fmt.Printf("[Session %s] Failed to get device ID: %v\n", sessionID, err)
		resultChan <- fmt.Errorf("failed to get device ID: %w", err)
		return
	}
	fmt.Printf("[Session %s] Using device ID: %s\n", sessionID, deviceID)

	// 1. Get WireGuard configuration (one-time call)
	fmt.Printf("[Session %s] Fetching WireGuard configuration...\n", sessionID)
	wgConfig, err := apiClient.GetWireGuardConfig(deviceID)
	if err != nil {
		fmt.Printf("[Session %s] Failed to get WireGuard config: %v\n", sessionID, err)
		resultChan <- fmt.Errorf("failed to get WireGuard config: %w", err)
		return
	}
	fmt.Printf("[Session %s] WireGuard config received - AssignedIP: %s\n", sessionID, wgConfig.AssignedIP)
	fmt.Printf("[Session %s] WireGuard IPC config:\n%s\n", sessionID, wgConfig.Config)

	// 2. Get CA certificate (one-time call)
	fmt.Printf("[Session %s] Fetching CA certificate...\n", sessionID)
	caCert, err := apiClient.GetCACert()
	if err != nil {
		fmt.Printf("[Session %s] Failed to get CA cert: %v\n", sessionID, err)
		resultChan <- fmt.Errorf("failed to get CA cert: %w", err)
		return
	}

	// Install CA certificate
	if caCert != "" {
		fmt.Printf("[Session %s] Installing CA certificate...\n", sessionID)
		certFile := fmt.Sprintf("/tmp/kltun-ca-%s.crt", sessionID)
		if err := os.WriteFile(certFile, []byte(caCert), 0o600); err != nil {
			fmt.Printf("[Session %s] Failed to write CA cert: %v\n", sessionID, err)
			resultChan <- fmt.Errorf("failed to write CA cert: %w", err)
			return
		}
		defer os.Remove(certFile)

		stores := []string{"system", "nss", "java"}
		if err := truststore.InstallAll(certFile, stores); err != nil {
			fmt.Printf("[Session %s] Warning: Failed to install CA cert: %v\n", sessionID, err)
		} else {
			fmt.Printf("[Session %s] ✓ CA certificate installed\n", sessionID)
		}
	}

	// 3. Start proxyguard client if ServerEndpoint is provided
	var pgClient *proxyguard.Client
	if wgConfig.ServerEndpoint != "" {
		fmt.Printf("[Session %s] Starting proxyguard client...\n", sessionID)
		fmt.Printf("[Session %s] Local: 127.0.0.1:51821 -> Remote: %s\n", sessionID, wgConfig.ServerEndpoint)

		// Initialize the official proxyguard client
		// Peer field expects just IP:PORT (not a URL)
		pgClient = &proxyguard.Client{
			ListenPort:    51821,                               // Local port for WireGuard to connect to
			TCPSourcePort: 0,                                   // Let kernel choose source port
			Fwmark:        -1,                                  // Firewall mark disabled
			PeerIPS:       []string{},                          // Empty peer IPs
			Peer:          wgConfig.ServerEndpoint, // Server endpoint (e.g., "203.0.113.1:443")
		}

		// Setup the client (creates UDP listener)
		if _, err := pgClient.Setup(ctx); err != nil {
			fmt.Printf("[Session %s] Failed to setup proxyguard: %v\n", sessionID, err)
			resultChan <- fmt.Errorf("failed to setup proxyguard: %w", err)
			return
		}
		defer pgClient.Close()

		// Start the tunnel in background (connects to server and starts forwarding)
		go func() {
			if err := pgClient.Tunnel(ctx, 51820); err != nil && ctx.Err() == nil {
				fmt.Printf("[Session %s] Proxyguard tunnel error: %v\n", sessionID, err)
			}
		}()

		fmt.Printf("[Session %s] ✓ Proxyguard client started\n", sessionID)
	}

	// 4. Start WireGuard device
	fmt.Printf("[Session %s] Starting WireGuard device...\n", sessionID)
	wgDeviceConfig := &wireguard.Config{
		ListenPort: 51820,
		MTU:        1420,
	}

	wgDevice, err := wireguard.NewDevice(ctx, wgDeviceConfig)
	if err != nil {
		fmt.Printf("[Session %s] Failed to create WireGuard device: %v\n", sessionID, err)
		resultChan <- fmt.Errorf("failed to create WireGuard device: %w", err)
		return
	}

	// Configure IP address and routes on the WireGuard interface
	fmt.Printf("[Session %s] Configuring network interface...\n", sessionID)
	netCfg := &netconfig.InterfaceConfig{
		InterfaceName: wgDevice.InterfaceName(),
		IPAddress:     fmt.Sprintf("%s/32", wgConfig.AssignedIP),
		Routes:        []string{"10.17.0.0/24", "10.43.0.0/16"},
		Gateway:       "10.17.0.1",
	}

	if err := netconfig.ConfigureInterface(netCfg); err != nil {
		fmt.Printf("[Session %s] Failed to configure network interface: %v\n", sessionID, err)
		resultChan <- fmt.Errorf("failed to configure network interface: %w", err)
		return
	}

	if err := wgDevice.Start(ctx); err != nil {
		fmt.Printf("[Session %s] Failed to start WireGuard device: %v\n", sessionID, err)
		resultChan <- fmt.Errorf("failed to start WireGuard device: %w", err)
		return
	}
	defer wgDevice.Close()
	fmt.Printf("[WGConfig] %s", wgConfig.Config)
	if err := wgDevice.SetConfig(wgConfig.Config); err != nil {
		fmt.Printf("[Session %s] Failed to set WireGuard config: %v\n", sessionID, err)
		resultChan <- fmt.Errorf("failed to set WireGuard config: %w", err)
		return
	}

	fmt.Printf("[Session %s] ✓ WireGuard device started (IP: %s)\n", sessionID, wgConfig.AssignedIP)

	// Signal success - connection is established
	resultChan <- nil

	// Continue with hosts polling in background
	fmt.Printf("[Session %s] Starting hosts polling (every 10 seconds)...\n", sessionID)
	hostsDone := make(chan struct{})
	s.pollHosts(ctx, sessionID, apiClient, hostsDone)

	fmt.Printf("[Session %s] VPN connection established successfully\n", sessionID)
}

// runVPNConnection runs a VPN connection with separate API calls
func (s *Server) runVPNConnection(ctx context.Context, sessionID, server, token string, done chan struct{}) {
	defer close(done) // Signal completion when function returns
	fmt.Printf("[Session %s] Starting VPN connection\n", sessionID)

	// Create API client
	apiClient := api.NewClient(server, token)

	// Get or create device ID
	deviceID, err := deviceid.GetOrCreateDeviceID()
	if err != nil {
		fmt.Printf("[Session %s] Failed to get device ID: %v\n", sessionID, err)
		return
	}
	fmt.Printf("[Session %s] Using device ID: %s\n", sessionID, deviceID)

	// 1. Get WireGuard configuration (one-time call)
	fmt.Printf("[Session %s] Fetching WireGuard configuration...\n", sessionID)
	wgConfig, err := apiClient.GetWireGuardConfig(deviceID)
	if err != nil {
		fmt.Printf("[Session %s] Failed to get WireGuard config: %v\n", sessionID, err)
		return
	}
	fmt.Printf("[Session %s] WireGuard config received - AssignedIP: %s\n", sessionID, wgConfig.AssignedIP)
	fmt.Printf("[Session %s] WireGuard IPC config:\n%s\n", sessionID, wgConfig.Config)

	// 2. Get CA certificate (one-time call)
	fmt.Printf("[Session %s] Fetching CA certificate...\n", sessionID)
	caCert, err := apiClient.GetCACert()
	if err != nil {
		fmt.Printf("[Session %s] Failed to get CA cert: %v\n", sessionID, err)
		return
	}

	// Install CA certificate
	if caCert != "" {
		fmt.Printf("[Session %s] Installing CA certificate...\n", sessionID)
		certFile := fmt.Sprintf("/tmp/kltun-ca-%s.crt", sessionID)
		if err := os.WriteFile(certFile, []byte(caCert), 0o600); err != nil {
			fmt.Printf("[Session %s] Failed to write CA cert: %v\n", sessionID, err)
			return
		}
		defer os.Remove(certFile)

		stores := []string{"system", "nss", "java"}
		if err := truststore.InstallAll(certFile, stores); err != nil {
			fmt.Printf("[Session %s] Warning: Failed to install CA cert: %v\n", sessionID, err)
		} else {
			fmt.Printf("[Session %s] ✓ CA certificate installed\n", sessionID)
		}
	}

	// 3. Start proxyguard client if ServerEndpoint is provided
	var pgClient *proxyguard.Client
	if wgConfig.ServerEndpoint != "" {
		fmt.Printf("[Session %s] Starting proxyguard client...\n", sessionID)
		fmt.Printf("[Session %s] Local: 127.0.0.1:51821 -> Remote: %s\n", sessionID, wgConfig.ServerEndpoint)

		// Initialize the official proxyguard client
		// Peer field expects just IP:PORT (not a URL)
		pgClient = &proxyguard.Client{
			ListenPort:    51821,                               // Local port for WireGuard to connect to
			TCPSourcePort: 0,                                   // Let kernel choose source port
			Fwmark:        -1,                                  // Firewall mark disabled
			PeerIPS:       []string{},                          // Empty peer IPs
			Peer:          wgConfig.ServerEndpoint, // Server endpoint (e.g., "203.0.113.1:443")
		}

		// Setup the client (creates UDP listener)
		if _, err := pgClient.Setup(ctx); err != nil {
			fmt.Printf("[Session %s] Failed to setup proxyguard: %v\n", sessionID, err)
			return
		}
		defer pgClient.Close()

		// Start the tunnel in background (connects to server and starts forwarding)
		go func() {
			if err := pgClient.Tunnel(ctx, 51820); err != nil && ctx.Err() == nil {
				fmt.Printf("[Session %s] Proxyguard tunnel error: %v\n", sessionID, err)
			}
		}()

		fmt.Printf("[Session %s] ✓ Proxyguard client started\n", sessionID)
	}

	// 4. Start WireGuard device
	fmt.Printf("[Session %s] Starting WireGuard device...\n", sessionID)
	wgDeviceConfig := &wireguard.Config{
		ListenPort: 51820,
		MTU:        1420,
	}

	wgDevice, err := wireguard.NewDevice(ctx, wgDeviceConfig)
	if err != nil {
		fmt.Printf("[Session %s] Failed to create WireGuard device: %v\n", sessionID, err)
		return
	}

	// Configure IP address and routes on the WireGuard interface
	fmt.Printf("[Session %s] Configuring network interface...\n", sessionID)
	netCfg := &netconfig.InterfaceConfig{
		InterfaceName: wgDevice.InterfaceName(),
		IPAddress:     fmt.Sprintf("%s/32", wgConfig.AssignedIP),
		Routes:        []string{"10.17.0.0/24", "10.43.0.0/16"},
		Gateway:       "10.17.0.1",
	}

	if err := netconfig.ConfigureInterface(netCfg); err != nil {
		fmt.Printf("[Session %s] Failed to configure network interface: %v\n", sessionID, err)
		return
	}

	if err := wgDevice.Start(ctx); err != nil {
		fmt.Printf("[Session %s] Failed to start WireGuard device: %v\n", sessionID, err)
		return
	}
	defer wgDevice.Close()
	fmt.Printf("[WGConfig] %s", wgConfig.Config)
	if err := wgDevice.SetConfig(wgConfig.Config); err != nil {
		fmt.Printf("[Session %s] Failed to set WireGuard config: %v\n", sessionID, err)
		return
	}

	fmt.Printf("[Session %s] ✓ WireGuard device started (IP: %s)\n", sessionID, wgConfig.AssignedIP)

	// 5. Start hosts polling goroutine (polls every 10 seconds)
	fmt.Printf("[Session %s] Starting hosts polling (every 10 seconds)...\n", sessionID)
	hostsCtx, hostsCancel := context.WithCancel(ctx)
	defer hostsCancel()

	hostsDone := make(chan struct{})
	go s.pollHosts(hostsCtx, sessionID, apiClient, hostsDone)

	fmt.Printf("[Session %s] VPN connection established successfully\n", sessionID)

	// Wait for context cancellation
	<-ctx.Done()

	fmt.Printf("[Session %s] Stopping VPN connection...\n", sessionID)

	// Wait for hosts polling to stop
	hostsCancel()
	<-hostsDone

	// Clean up network configuration
	fmt.Printf("[Session %s] Cleaning up network configuration...\n", sessionID)
	if err := netconfig.RemoveInterface(netCfg); err != nil {
		fmt.Printf("[Session %s] Warning: Failed to remove network configuration: %v\n", sessionID, err)
	}

	// Cleanup all hosts for this session
	s.cleanupSessionHosts(sessionID)

	fmt.Printf("[Session %s] VPN connection stopped\n", sessionID)
}

// cleanupSessionHosts removes all hosts for a session
func (s *Server) cleanupSessionHosts(sessionID string) {
	fmt.Printf("[Session %s] Cleaning up hosts...\n", sessionID)

	// List all hosts managed by hostsManager
	entries, err := s.hostsManager.List()
	if err != nil {
		fmt.Printf("[Session %s] Warning: Failed to list hosts: %v\n", sessionID, err)
		return
	}

	// Remove hosts that belong to this session
	sessionComment := fmt.Sprintf("# kltun session %s", sessionID)
	for _, entry := range entries {
		if entry.Comment == sessionComment {
			if err := s.hostsManager.Remove(entry.Hostname); err != nil {
				fmt.Printf("[Session %s] Warning: Failed to remove host %s: %v\n", sessionID, entry.Hostname, err)
			}
		}
	}
}
