package daemon

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/api"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/deviceid"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/netconfig"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/truststore"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/wireguard"
	"github.com/kloudlite/kloudlite/api/pkg/udptunnel/transport"
	"github.com/kloudlite/kloudlite/api/pkg/udptunnel/tunnel"
	"go.uber.org/zap"
)

// runVPNConnectionWithResult runs a VPN connection and sends result on channel after initial setup
// New architecture: Dashboard only provides tunnel endpoint, all other calls go directly to tunnel server
func (s *Server) runVPNConnectionWithResult(ctx context.Context, sessionID, server, token string, done chan struct{}, resultChan chan<- error) {
	defer close(done) // Signal completion when function returns
	fmt.Printf("[Session %s] Starting VPN connection\n", sessionID)

	// Create Dashboard API client (only for getting tunnel endpoint)
	dashboardClient := api.NewClient(server, token)

	// Get or create device ID
	deviceID, err := deviceid.GetOrCreateDeviceID()
	if err != nil {
		fmt.Printf("[Session %s] Failed to get device ID: %v\n", sessionID, err)
		resultChan <- fmt.Errorf("failed to get device ID: %w", err)
		return
	}
	fmt.Printf("[Session %s] Using device ID: %s\n", sessionID, deviceID)

	// 1. Get tunnel endpoint from Dashboard (the only Dashboard call needed)
	fmt.Printf("[Session %s] Getting tunnel endpoint from Dashboard...\n", sessionID)
	tunnelEndpoint, err := dashboardClient.GetTunnelEndpoint()
	if err != nil {
		fmt.Printf("[Session %s] Failed to get tunnel endpoint: %v\n", sessionID, err)
		resultChan <- fmt.Errorf("failed to get tunnel endpoint: %w", err)
		return
	}
	fmt.Printf("[Session %s] Tunnel endpoint: %s\n", sessionID, tunnelEndpoint)

	// 2. Create tunnel server client for direct communication
	tunnelClient := api.NewTunnelClient(tunnelEndpoint)

	// 3. Create WireGuard peer on tunnel server
	fmt.Printf("[Session %s] Creating WireGuard peer on tunnel server...\n", sessionID)
	peerResp, err := tunnelClient.CreatePeer(deviceID)
	if err != nil {
		fmt.Printf("[Session %s] Failed to create WireGuard peer: %v\n", sessionID, err)
		resultChan <- fmt.Errorf("failed to create WireGuard peer: %w", err)
		return
	}
	fmt.Printf("[Session %s] WireGuard peer created - IP: %s\n", sessionID, peerResp.IP)

	// 4. Get CA certificate from tunnel server
	fmt.Printf("[Session %s] Fetching CA certificate from tunnel server...\n", sessionID)
	caCert, err := tunnelClient.GetCACert()
	if err != nil {
		fmt.Printf("[Session %s] Warning: Failed to get CA cert: %v\n", sessionID, err)
		// Don't fail - CA cert might not be available
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

	// 5. Start UDP-over-WebSocket client
	fmt.Printf("[Session %s] Starting UDP-over-WebSocket client...\n", sessionID)
	fmt.Printf("[Session %s] Local: 127.0.0.1:51821 -> Remote: %s\n", sessionID, tunnelEndpoint)

	// Create logger for UDP tunnel
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("[Session %s] Failed to create logger: %v\n", sessionID, err)
		resultChan <- fmt.Errorf("failed to create logger: %w", err)
		return
	}
	defer logger.Sync()

	// Create WebSocket dialer with TLS 1.3 (required by tunnel-server)
	transportConfig := transport.DefaultConfig()
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true, // Tunnel server uses self-signed cert
	}
	dialer := transport.NewWebSocketDialer(
		transportConfig,
		tlsConfig,
		nil, // No custom headers
		logger,
	)

	// Create UDP tunnel client
	// Local: 127.0.0.1:51821 (where WireGuard will connect)
	// Server: wss://tunnel-endpoint (WebSocket server)
	// Remote: 127.0.0.1:51820 (WireGuard on server side)
	serverURL := "wss://" + tunnelEndpoint + "/ws"
	udpClient := tunnel.NewUDPClient(
		"127.0.0.1:51821",
		serverURL,
		"127.0.0.1:51820",
		dialer,
		logger,
	)

	// Start UDP tunnel client in background
	go func() {
		if err := udpClient.Start(ctx); err != nil && ctx.Err() == nil {
			fmt.Printf("[Session %s] UDP tunnel error: %v\n", sessionID, err)
		}
	}()

	fmt.Printf("[Session %s] ✓ UDP-over-WebSocket client started\n", sessionID)

	// 6. Start WireGuard device
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
		IPAddress:     fmt.Sprintf("%s/32", peerResp.IP),
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
	fmt.Printf("[WGConfig] %s", peerResp.Config)
	if err := wgDevice.SetConfig(peerResp.Config); err != nil {
		fmt.Printf("[Session %s] Failed to set WireGuard config: %v\n", sessionID, err)
		resultChan <- fmt.Errorf("failed to set WireGuard config: %w", err)
		return
	}

	fmt.Printf("[Session %s] ✓ WireGuard device started (IP: %s)\n", sessionID, peerResp.IP)

	// Signal success - connection is established
	resultChan <- nil

	// Continue with hosts polling in background (now using tunnel client)
	fmt.Printf("[Session %s] Starting hosts polling (every 10 seconds)...\n", sessionID)
	hostsDone := make(chan struct{})
	s.pollHostsFromTunnel(ctx, sessionID, tunnelClient, hostsDone)

	fmt.Printf("[Session %s] VPN connection established successfully\n", sessionID)
}

// runVPNConnection runs a VPN connection with the new architecture
// Dashboard only provides tunnel endpoint, all other calls go directly to tunnel server
func (s *Server) runVPNConnection(ctx context.Context, sessionID, server, token string, done chan struct{}) {
	defer close(done) // Signal completion when function returns
	fmt.Printf("[Session %s] Starting VPN connection\n", sessionID)

	// Create Dashboard API client (only for getting tunnel endpoint)
	dashboardClient := api.NewClient(server, token)

	// Get or create device ID
	deviceID, err := deviceid.GetOrCreateDeviceID()
	if err != nil {
		fmt.Printf("[Session %s] Failed to get device ID: %v\n", sessionID, err)
		return
	}
	fmt.Printf("[Session %s] Using device ID: %s\n", sessionID, deviceID)

	// 1. Get tunnel endpoint from Dashboard (the only Dashboard call needed)
	fmt.Printf("[Session %s] Getting tunnel endpoint from Dashboard...\n", sessionID)
	tunnelEndpoint, err := dashboardClient.GetTunnelEndpoint()
	if err != nil {
		fmt.Printf("[Session %s] Failed to get tunnel endpoint: %v\n", sessionID, err)
		return
	}
	fmt.Printf("[Session %s] Tunnel endpoint: %s\n", sessionID, tunnelEndpoint)

	// 2. Create tunnel server client for direct communication
	tunnelClient := api.NewTunnelClient(tunnelEndpoint)

	// 3. Create WireGuard peer on tunnel server
	fmt.Printf("[Session %s] Creating WireGuard peer on tunnel server...\n", sessionID)
	peerResp, err := tunnelClient.CreatePeer(deviceID)
	if err != nil {
		fmt.Printf("[Session %s] Failed to create WireGuard peer: %v\n", sessionID, err)
		return
	}
	fmt.Printf("[Session %s] WireGuard peer created - IP: %s\n", sessionID, peerResp.IP)

	// 4. Get CA certificate from tunnel server
	fmt.Printf("[Session %s] Fetching CA certificate from tunnel server...\n", sessionID)
	caCert, err := tunnelClient.GetCACert()
	if err != nil {
		fmt.Printf("[Session %s] Warning: Failed to get CA cert: %v\n", sessionID, err)
		// Don't fail - CA cert might not be available
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

	// 5. Start UDP-over-WebSocket client
	fmt.Printf("[Session %s] Starting UDP-over-WebSocket client...\n", sessionID)
	fmt.Printf("[Session %s] Local: 127.0.0.1:51821 -> Remote: %s\n", sessionID, tunnelEndpoint)

	// Create logger for UDP tunnel
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("[Session %s] Failed to create logger: %v\n", sessionID, err)
		return
	}
	defer logger.Sync()

	// Create WebSocket dialer with TLS 1.3 (required by tunnel-server)
	transportConfig := transport.DefaultConfig()
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true, // Tunnel server uses self-signed cert
	}
	dialer := transport.NewWebSocketDialer(
		transportConfig,
		tlsConfig,
		nil, // No custom headers
		logger,
	)

	// Create UDP tunnel client
	// Local: 127.0.0.1:51821 (where WireGuard will connect)
	// Server: wss://tunnel-endpoint (WebSocket server)
	// Remote: 127.0.0.1:51820 (WireGuard on server side)
	serverURL := "wss://" + tunnelEndpoint + "/ws"
	udpClient := tunnel.NewUDPClient(
		"127.0.0.1:51821",
		serverURL,
		"127.0.0.1:51820",
		dialer,
		logger,
	)

	// Start UDP tunnel client in background
	go func() {
		if err := udpClient.Start(ctx); err != nil && ctx.Err() == nil {
			fmt.Printf("[Session %s] UDP tunnel error: %v\n", sessionID, err)
		}
	}()

	fmt.Printf("[Session %s] ✓ UDP-over-WebSocket client started\n", sessionID)

	// 6. Start WireGuard device
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
		IPAddress:     fmt.Sprintf("%s/32", peerResp.IP),
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
	fmt.Printf("[WGConfig] %s", peerResp.Config)
	if err := wgDevice.SetConfig(peerResp.Config); err != nil {
		fmt.Printf("[Session %s] Failed to set WireGuard config: %v\n", sessionID, err)
		return
	}

	fmt.Printf("[Session %s] ✓ WireGuard device started (IP: %s)\n", sessionID, peerResp.IP)

	// 7. Start hosts polling goroutine (polls from tunnel server every 10 seconds)
	fmt.Printf("[Session %s] Starting hosts polling (every 10 seconds)...\n", sessionID)
	hostsCtx, hostsCancel := context.WithCancel(ctx)
	defer hostsCancel()

	hostsDone := make(chan struct{})
	go s.pollHostsFromTunnel(hostsCtx, sessionID, tunnelClient, hostsDone)

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
