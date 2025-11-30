package daemon

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/api"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/deviceid"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/netconfig"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/truststore"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/wgkeys"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/wireguard"
	"github.com/kloudlite/kloudlite/api/pkg/udptunnel/transport"
	"github.com/kloudlite/kloudlite/api/pkg/udptunnel/tunnel"
	"go.uber.org/zap"
)

// runVPNConnectionWithResult runs a VPN connection and sends result on channel after initial setup
// New architecture: Dashboard only provides tunnel endpoint, all other calls go directly to tunnel server
func (s *Server) runVPNConnectionWithResult(ctx context.Context, sessionID, server, token string, done chan struct{}, resultChan chan<- VPNConnectionSetupResult) {
	defer close(done) // Signal completion when function returns
	fmt.Printf("[Session %s] Starting VPN connection\n", sessionID)

	// Create Dashboard API client (only for getting tunnel endpoint and token exchange)
	dashboardClient := api.NewClient(server, token)

	// Exchange short-lived token for long-lived permanent token (1 year)
	// This is critical for hosts polling which continues for the duration of the VPN connection
	fmt.Printf("[Session %s] Exchanging temporary token for permanent token...\n", sessionID)
	exchangeResp, err := dashboardClient.ExchangeToken(token)
	if err != nil {
		fmt.Printf("[Session %s] Failed to exchange token: %v\n", sessionID, err)
		resultChan <- VPNConnectionSetupResult{Error: fmt.Errorf("failed to exchange token: %w", err)}
		return
	}
	permanentToken := exchangeResp.ConnectionToken
	fmt.Printf("[Session %s] ✓ Token exchanged successfully (valid for 1 year)\n", sessionID)

	// Get or create device ID
	deviceID, err := deviceid.GetOrCreateDeviceID()
	if err != nil {
		fmt.Printf("[Session %s] Failed to get device ID: %v\n", sessionID, err)
		resultChan <- VPNConnectionSetupResult{Error: fmt.Errorf("failed to get device ID: %w", err)}
		return
	}
	fmt.Printf("[Session %s] Using device ID: %s\n", sessionID, deviceID)

	// Get or create WireGuard key pair
	keyPair, err := wgkeys.GetOrCreateKeyPair()
	if err != nil {
		fmt.Printf("[Session %s] Failed to get WireGuard keys: %v\n", sessionID, err)
		resultChan <- VPNConnectionSetupResult{Error: fmt.Errorf("failed to get WireGuard keys: %w", err)}
		return
	}
	fmt.Printf("[Session %s] Using WireGuard public key: %s...\n", sessionID, keyPair.PublicKey[:20])

	// 1. Get tunnel endpoint from Dashboard (the only Dashboard call needed)
	fmt.Printf("[Session %s] Getting tunnel endpoint from Dashboard...\n", sessionID)
	tunnelInfo, err := dashboardClient.GetTunnelEndpoint()
	if err != nil {
		fmt.Printf("[Session %s] Failed to get tunnel endpoint: %v\n", sessionID, err)
		resultChan <- VPNConnectionSetupResult{Error: fmt.Errorf("failed to get tunnel endpoint: %w", err)}
		return
	}
	fmt.Printf("[Session %s] Tunnel endpoint: %s (IP: %s)\n", sessionID, tunnelInfo.Hostname, tunnelInfo.IP)

	// Add vpn-connect hostname to /etc/hosts BEFORE connecting
	fmt.Printf("[Session %s] Adding vpn-connect host entry: %s -> %s\n", sessionID, tunnelInfo.Hostname, tunnelInfo.IP)
	if err := s.hostsManager.Add(tunnelInfo.Hostname, tunnelInfo.IP, fmt.Sprintf("# kltun session %s vpn-connect", sessionID)); err != nil {
		fmt.Printf("[Session %s] Warning: Failed to add vpn-connect host: %v\n", sessionID, err)
		// Don't fail - try to connect anyway (might work via DNS)
	}

	tunnelEndpoint := tunnelInfo.TunnelEndpoint

	// 2. Create tunnel server client for direct communication (with permanent token)
	tunnelClient := api.NewTunnelClient(tunnelEndpoint, permanentToken)

	// 3. Create WireGuard peer on tunnel server (send our public key)
	fmt.Printf("[Session %s] Registering WireGuard peer on tunnel server...\n", sessionID)
	peerResp, err := tunnelClient.CreatePeer(deviceID, keyPair.PublicKey)
	if err != nil {
		fmt.Printf("[Session %s] Failed to register WireGuard peer: %v\n", sessionID, err)
		resultChan <- VPNConnectionSetupResult{Error: fmt.Errorf("failed to register WireGuard peer: %w", err)}
		return
	}
	if peerResp.AlreadyExists {
		fmt.Printf("[Session %s] WireGuard peer already exists - IP: %s\n", sessionID, peerResp.IP)
	} else {
		fmt.Printf("[Session %s] WireGuard peer created - IP: %s\n", sessionID, peerResp.IP)
	}

	// 4. Get CA certificate from tunnel server with retry
	fmt.Printf("[Session %s] Fetching CA certificate from tunnel server...\n", sessionID)
	var caCert string
	var caCertErr error
	for attempt := 1; attempt <= 3; attempt++ {
		caCert, caCertErr = tunnelClient.GetCACert()
		if caCertErr == nil && caCert != "" {
			break
		}
		if attempt < 3 {
			fmt.Printf("[Session %s] CA cert fetch attempt %d failed, retrying in %ds...\n", sessionID, attempt, attempt)
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	// Track CA cert installation status
	caCertInstalled := false
	var caCertInstallErr string

	if caCertErr != nil {
		caCertInstallErr = fmt.Sprintf("failed to fetch CA cert: %v", caCertErr)
		fmt.Printf("[Session %s] ✗ %s\n", sessionID, caCertInstallErr)
	} else if caCert == "" {
		caCertInstallErr = "CA certificate is empty"
		fmt.Printf("[Session %s] ✗ %s\n", sessionID, caCertInstallErr)
	} else {
		// Install CA certificate
		fmt.Printf("[Session %s] Installing CA certificate to trust stores...\n", sessionID)
		certFile := filepath.Join(os.TempDir(), fmt.Sprintf("kltun-ca-%s.crt", sessionID))
		if err := os.WriteFile(certFile, []byte(caCert), 0o600); err != nil {
			caCertInstallErr = fmt.Sprintf("failed to write CA cert: %v", err)
			fmt.Printf("[Session %s] ✗ %s\n", sessionID, caCertInstallErr)
		} else {
			defer os.Remove(certFile)
			stores := []string{"system", "nss", "java"}
			if err := truststore.InstallAll(certFile, stores); err != nil {
				caCertInstallErr = fmt.Sprintf("failed to install CA cert: %v", err)
				fmt.Printf("[Session %s] ✗ %s\n", sessionID, caCertInstallErr)
			} else {
				caCertInstalled = true
				fmt.Printf("[Session %s] ✓ CA certificate installed\n", sessionID)
			}
		}
	}

	// 5. Start UDP-over-WebSocket client
	fmt.Printf("[Session %s] Starting UDP-over-WebSocket client...\n", sessionID)
	fmt.Printf("[Session %s] Local: 127.0.0.1:51821 -> Remote: %s\n", sessionID, tunnelEndpoint)

	// Create logger for UDP tunnel
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("[Session %s] Failed to create logger: %v\n", sessionID, err)
		resultChan <- VPNConnectionSetupResult{Error: fmt.Errorf("failed to create logger: %w", err)}
		return
	}
	defer logger.Sync()

	// Create WebSocket dialer with TLS 1.3 (required by tunnel-server)
	transportConfig := transport.DefaultConfig()
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true, // Tunnel server uses self-signed cert
	}
	// Add Authorization header for WebSocket connection (using permanent token)
	wsHeaders := http.Header{
		"Authorization": []string{"Bearer " + permanentToken},
	}
	dialer := transport.NewWebSocketDialer(
		transportConfig,
		tlsConfig,
		wsHeaders,
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
		resultChan <- VPNConnectionSetupResult{Error: fmt.Errorf("failed to create WireGuard device: %w", err)}
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
		resultChan <- VPNConnectionSetupResult{Error: fmt.Errorf("failed to configure network interface: %w", err)}
		return
	}

	if err := wgDevice.Start(ctx); err != nil {
		fmt.Printf("[Session %s] Failed to start WireGuard device: %v\n", sessionID, err)
		resultChan <- VPNConnectionSetupResult{Error: fmt.Errorf("failed to start WireGuard device: %w", err)}
		return
	}
	defer wgDevice.Close()

	// Build WireGuard config locally using our private key and server's response
	wgConfig := buildWireGuardConfig(keyPair.PrivateKey, peerResp.IP, peerResp.ServerPublicKey, peerResp.CIDR)
	fmt.Printf("[WGConfig] %s", wgConfig)
	if err := wgDevice.SetConfig(wgConfig); err != nil {
		fmt.Printf("[Session %s] Failed to set WireGuard config: %v\n", sessionID, err)
		resultChan <- VPNConnectionSetupResult{Error: fmt.Errorf("failed to set WireGuard config: %w", err)}
		return
	}

	fmt.Printf("[Session %s] ✓ WireGuard device started (IP: %s)\n", sessionID, peerResp.IP)

	// Signal success - connection is established with CA cert status
	resultChan <- VPNConnectionSetupResult{
		Error:           nil,
		CACertInstalled: caCertInstalled,
		CACertError:     caCertInstallErr,
	}

	// Continue with hosts polling in background (now using tunnel client)
	fmt.Printf("[Session %s] Starting hosts polling (every 10 seconds)...\n", sessionID)
	hostsCtx, hostsCancel := context.WithCancel(ctx)
	hostsDone := make(chan struct{})
	go s.pollHostsFromTunnel(hostsCtx, sessionID, tunnelClient, hostsDone)

	fmt.Printf("[Session %s] VPN connection established successfully\n", sessionID)

	// Wait for context cancellation (VPN quit)
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

// runVPNConnection runs a VPN connection with the new architecture
// Dashboard only provides tunnel endpoint, all other calls go directly to tunnel server
func (s *Server) runVPNConnection(ctx context.Context, sessionID, server, token string, done chan struct{}) {
	defer close(done) // Signal completion when function returns
	fmt.Printf("[Session %s] Starting VPN connection\n", sessionID)

	// Create Dashboard API client (only for getting tunnel endpoint and token exchange)
	dashboardClient := api.NewClient(server, token)

	// Exchange short-lived token for long-lived permanent token (1 year)
	// This is critical for hosts polling which continues for the duration of the VPN connection
	fmt.Printf("[Session %s] Exchanging temporary token for permanent token...\n", sessionID)
	exchangeResp, err := dashboardClient.ExchangeToken(token)
	if err != nil {
		fmt.Printf("[Session %s] Failed to exchange token: %v\n", sessionID, err)
		return
	}
	permanentToken := exchangeResp.ConnectionToken
	fmt.Printf("[Session %s] ✓ Token exchanged successfully (valid for 1 year)\n", sessionID)

	// Get or create device ID
	deviceID, err := deviceid.GetOrCreateDeviceID()
	if err != nil {
		fmt.Printf("[Session %s] Failed to get device ID: %v\n", sessionID, err)
		return
	}
	fmt.Printf("[Session %s] Using device ID: %s\n", sessionID, deviceID)

	// Get or create WireGuard key pair
	keyPair, err := wgkeys.GetOrCreateKeyPair()
	if err != nil {
		fmt.Printf("[Session %s] Failed to get WireGuard keys: %v\n", sessionID, err)
		return
	}
	fmt.Printf("[Session %s] Using WireGuard public key: %s...\n", sessionID, keyPair.PublicKey[:20])

	// 1. Get tunnel endpoint from Dashboard (the only Dashboard call needed)
	fmt.Printf("[Session %s] Getting tunnel endpoint from Dashboard...\n", sessionID)
	tunnelInfo, err := dashboardClient.GetTunnelEndpoint()
	if err != nil {
		fmt.Printf("[Session %s] Failed to get tunnel endpoint: %v\n", sessionID, err)
		return
	}
	fmt.Printf("[Session %s] Tunnel endpoint: %s (IP: %s)\n", sessionID, tunnelInfo.Hostname, tunnelInfo.IP)

	// Add vpn-connect hostname to /etc/hosts BEFORE connecting
	fmt.Printf("[Session %s] Adding vpn-connect host entry: %s -> %s\n", sessionID, tunnelInfo.Hostname, tunnelInfo.IP)
	if err := s.hostsManager.Add(tunnelInfo.Hostname, tunnelInfo.IP, fmt.Sprintf("# kltun session %s vpn-connect", sessionID)); err != nil {
		fmt.Printf("[Session %s] Warning: Failed to add vpn-connect host: %v\n", sessionID, err)
		// Don't fail - try to connect anyway (might work via DNS)
	}

	tunnelEndpoint := tunnelInfo.TunnelEndpoint

	// 2. Create tunnel server client for direct communication (with permanent token)
	tunnelClient := api.NewTunnelClient(tunnelEndpoint, permanentToken)

	// 3. Create WireGuard peer on tunnel server (send our public key)
	fmt.Printf("[Session %s] Registering WireGuard peer on tunnel server...\n", sessionID)
	peerResp, err := tunnelClient.CreatePeer(deviceID, keyPair.PublicKey)
	if err != nil {
		fmt.Printf("[Session %s] Failed to register WireGuard peer: %v\n", sessionID, err)
		return
	}
	if peerResp.AlreadyExists {
		fmt.Printf("[Session %s] WireGuard peer already exists - IP: %s\n", sessionID, peerResp.IP)
	} else {
		fmt.Printf("[Session %s] WireGuard peer created - IP: %s\n", sessionID, peerResp.IP)
	}

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
	// Add Authorization header for WebSocket connection (using permanent token)
	wsHeaders := http.Header{
		"Authorization": []string{"Bearer " + permanentToken},
	}
	dialer := transport.NewWebSocketDialer(
		transportConfig,
		tlsConfig,
		wsHeaders,
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

	// Build WireGuard config locally using our private key and server's response
	wgConfig := buildWireGuardConfig(keyPair.PrivateKey, peerResp.IP, peerResp.ServerPublicKey, peerResp.CIDR)
	fmt.Printf("[WGConfig] %s", wgConfig)
	if err := wgDevice.SetConfig(wgConfig); err != nil {
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
	// Match both regular hosts and vpn-connect host (different comment formats)
	sessionPrefix := fmt.Sprintf("# kltun session %s", sessionID)
	for _, entry := range entries {
		// Match comments that start with the session prefix
		// This catches both "# kltun session X" and "# kltun session X vpn-connect"
		if len(entry.Comment) >= len(sessionPrefix) && entry.Comment[:len(sessionPrefix)] == sessionPrefix {
			fmt.Printf("[Session %s] Removing host: %s\n", sessionID, entry.Hostname)
			if err := s.hostsManager.Remove(entry.Hostname); err != nil {
				fmt.Printf("[Session %s] Warning: Failed to remove host %s: %v\n", sessionID, entry.Hostname, err)
			}
		}
	}
}

// buildWireGuardConfig generates a WireGuard configuration string
// The endpoint is 127.0.0.1:51821 because the client runs a local UDP proxy
// that tunnels traffic over WebSocket to the server
func buildWireGuardConfig(privateKey, peerIP, serverPublicKey, cidr string) string {
	// AllowedIPs includes:
	// - VPN CIDR (e.g., 10.17.0.0/24) for VPN gateway communication
	// - Service CIDR (10.43.0.0/16) for ClusterIP service access
	return fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/32

[Peer]
PublicKey = %s
AllowedIPs = %s, 10.43.0.0/16
Endpoint = 127.0.0.1:51821
PersistentKeepalive = 25
`, privateKey, peerIP, serverPublicKey, cidr)
}
