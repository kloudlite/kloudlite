package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/api"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/deviceid"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/hosts"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/netconfig"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/proxyguard"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/truststore"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/wireguard"
)

// Server represents the daemon RPC server
type Server struct {
	listener     net.Listener
	hostsManager hosts.Manager
	connections  map[string]*VPNConnection
	connMutex    sync.RWMutex
	shutdownCh   chan struct{}
	wg           sync.WaitGroup
}

// VPNConnection represents an active VPN connection
type VPNConnection struct {
	SessionID  string
	Server     string
	StartTime  time.Time
	CancelFunc context.CancelFunc
}

// NewServer creates a new daemon server
func NewServer() (*Server, error) {
	hostsManager := hosts.NewManager()

	return &Server{
		hostsManager: hostsManager,
		connections:  make(map[string]*VPNConnection),
		shutdownCh:   make(chan struct{}),
	}, nil
}

// Start starts the RPC server
func (s *Server) Start(socketPath string) error {
	// Remove existing socket if it exists
	if err := os.RemoveAll(socketPath); err != nil {
		return fmt.Errorf("failed to remove existing socket: %w", err)
	}

	// Create listener
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}

	// Set socket permissions (allow all users to connect)
	// 0666 allows all users to read/write to the socket
	if err := os.Chmod(socketPath, 0666); err != nil {
		listener.Close()
		return fmt.Errorf("failed to set socket permissions: %w", err)
	}

	s.listener = listener

	// Handle shutdown signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		fmt.Println("\nReceived shutdown signal, stopping daemon...")
		s.Stop()
	}()

	fmt.Printf("Daemon server listening on %s\n", socketPath)

	// Accept connections
	for {
		select {
		case <-s.shutdownCh:
			return nil
		default:
		}

		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.shutdownCh:
				return nil
			default:
				fmt.Printf("Error accepting connection: %v\n", err)
				continue
			}
		}

		s.wg.Add(1)
		go s.handleConnection(conn)
	}
}

// Stop stops the server
func (s *Server) Stop() {
	close(s.shutdownCh)

	// Cancel all active VPN connections
	s.connMutex.Lock()
	for _, conn := range s.connections {
		if conn.CancelFunc != nil {
			conn.CancelFunc()
		}
	}
	s.connMutex.Unlock()

	if s.listener != nil {
		s.listener.Close()
	}

	s.wg.Wait()
}

// handleConnection handles a single client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer s.wg.Done()
	defer conn.Close()

	decoder := json.NewDecoder(conn)
	encoder := json.NewEncoder(conn)

	for {
		var req Request
		if err := decoder.Decode(&req); err != nil {
			if err == io.EOF {
				return
			}
			fmt.Printf("Error decoding request: %v\n", err)
			return
		}

		resp := s.handleRequest(&req)
		if err := encoder.Encode(resp); err != nil {
			fmt.Printf("Error encoding response: %v\n", err)
			return
		}
	}
}

// handleRequest handles a single RPC request
func (s *Server) handleRequest(req *Request) *Response {
	if req.JSONRPC != "2.0" {
		return NewErrorResponse(req.ID, ErrCodeInvalidRequest, "Invalid JSON-RPC version", "")
	}

	switch req.Method {
	case MethodPing:
		return s.handlePing(req)
	case MethodInstallCA:
		return s.handleInstallCA(req)
	case MethodUninstallCA:
		return s.handleUninstallCA(req)
	case MethodHostsAdd:
		return s.handleHostsAdd(req)
	case MethodHostsRemove:
		return s.handleHostsRemove(req)
	case MethodHostsList:
		return s.handleHostsList(req)
	case MethodHostsSync:
		return s.handleHostsSync(req)
	case MethodHostsClean:
		return s.handleHostsClean(req)
	case MethodHostsFlush:
		return s.handleHostsFlush(req)
	case MethodVPNConnect:
		return s.handleVPNConnect(req)
	case MethodVPNQuit:
		return s.handleVPNQuit(req)
	case MethodStatus:
		return s.handleStatus(req)
	default:
		return NewErrorResponse(req.ID, ErrCodeMethodNotFound, "Method not found", req.Method)
	}
}

// Handler methods

func (s *Server) handlePing(req *Request) *Response {
	result := PingResult{Message: "pong"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

func (s *Server) handleInstallCA(req *Request) *Response {
	var params InstallCAParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid parameters", err.Error())
	}

	// Install to all stores
	stores := []string{"system", "nss", "java"}
	if err := truststore.InstallAll(params.CertPath, stores); err != nil {
		result := InstallCAResult{Success: false, Message: err.Error()}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	result := InstallCAResult{Success: true, Message: "CA certificate installed successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

func (s *Server) handleUninstallCA(req *Request) *Response {
	var params UninstallCAParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid parameters", err.Error())
	}

	// Uninstall from all stores
	stores := []string{"system", "nss", "java"}
	if err := truststore.UninstallAll(params.CertPath, stores); err != nil {
		result := UninstallCAResult{Success: false, Message: err.Error()}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	result := UninstallCAResult{Success: true, Message: "CA certificate uninstalled successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

func (s *Server) handleHostsAdd(req *Request) *Response {
	var params HostsAddParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid parameters", err.Error())
	}

	if err := s.hostsManager.Add(params.Hostname, params.IP, params.Comment); err != nil {
		result := HostsAddResult{Success: false, Message: err.Error()}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	result := HostsAddResult{Success: true, Message: "Host entry added successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

func (s *Server) handleHostsRemove(req *Request) *Response {
	var params HostsRemoveParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid parameters", err.Error())
	}

	if err := s.hostsManager.Remove(params.Hostname); err != nil {
		result := HostsRemoveResult{Success: false, Message: err.Error()}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	result := HostsRemoveResult{Success: true, Message: "Host entry removed successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

func (s *Server) handleHostsList(req *Request) *Response {
	entries, err := s.hostsManager.List()
	if err != nil {
		return NewErrorResponse(req.ID, ErrCodeInternal, "Failed to list hosts", err.Error())
	}

	// Convert to protocol entries
	var protoEntries []HostsEntry
	for _, entry := range entries {
		protoEntries = append(protoEntries, HostsEntry{
			IP:       entry.IP,
			Hostname: entry.Hostname,
			Comment:  entry.Comment,
		})
	}

	result := HostsListResult{Entries: protoEntries}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

func (s *Server) handleHostsSync(req *Request) *Response {
	if err := s.hostsManager.Sync(); err != nil {
		result := HostsSyncResult{Success: false, Message: err.Error()}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	result := HostsSyncResult{Success: true, Message: "Hosts synchronized successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

func (s *Server) handleHostsClean(req *Request) *Response {
	if err := s.hostsManager.Clean(); err != nil {
		result := HostsCleanResult{Success: false, Message: err.Error()}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	result := HostsCleanResult{Success: true, Message: "Hosts cleaned successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

func (s *Server) handleHostsFlush(req *Request) *Response {
	if err := s.hostsManager.Flush(); err != nil {
		result := HostsFlushResult{Success: false, Message: err.Error()}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	result := HostsFlushResult{Success: true, Message: "DNS cache flushed successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

func (s *Server) handleVPNConnect(req *Request) *Response {
	var params VPNConnectParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return NewErrorResponse(req.ID, ErrCodeInvalidParams, "Invalid parameters", err.Error())
	}

	// Token and server must be provided - no persistence
	token := params.Token
	server := params.Server

	// Validate we have token and server
	if token == "" || server == "" {
		result := VPNConnectResult{Success: false, Message: "Token and server are required. Credentials are not saved - you must provide them on each connection."}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	// Generate session ID
	sessionID := fmt.Sprintf("conn-%d", time.Now().Unix())

	// Create context for this connection
	ctx, cancel := context.WithCancel(context.Background())

	conn := &VPNConnection{
		SessionID:  sessionID,
		Server:     server,
		StartTime:  time.Now(),
		CancelFunc: cancel,
	}

	s.connMutex.Lock()
	// Disconnect all existing connections before starting new one
	for existingSessionID, existingConn := range s.connections {
		fmt.Printf("Disconnecting existing connection: %s\n", existingSessionID)
		existingConn.CancelFunc()
		delete(s.connections, existingSessionID)
	}
	// Add the new connection
	s.connections[sessionID] = conn
	s.connMutex.Unlock()

	// Start VPN connection in background with server and token
	go s.runVPNConnection(ctx, sessionID, server, token)

	result := VPNConnectResult{
		Success:   true,
		Message:   "VPN connection started successfully",
		SessionID: sessionID,
	}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

func (s *Server) handleVPNQuit(req *Request) *Response {
	// Find the active connection (we only support one connection at a time for now)
	s.connMutex.Lock()
	var sessionID string
	var conn *VPNConnection
	for id, c := range s.connections {
		sessionID = id
		conn = c
		break
	}
	s.connMutex.Unlock()

	if conn == nil {
		result := VPNQuitResult{Success: false, Message: "No active VPN connection"}
		resp, _ := NewSuccessResponse(req.ID, result)
		return resp
	}

	// Cancel the connection
	if conn.CancelFunc != nil {
		conn.CancelFunc()
	}

	// Remove from connections map
	s.connMutex.Lock()
	delete(s.connections, sessionID)
	s.connMutex.Unlock()

	result := VPNQuitResult{Success: true, Message: "VPN connection stopped successfully"}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

func (s *Server) handleStatus(req *Request) *Response {
	s.connMutex.RLock()
	var connStatuses []ConnectionStatus
	for _, conn := range s.connections {
		connStatuses = append(connStatuses, ConnectionStatus{
			SessionID: conn.SessionID,
			Server:    conn.Server,
			Connected: true,
			Uptime:    int64(time.Since(conn.StartTime).Seconds()),
		})
	}
	s.connMutex.RUnlock()

	result := StatusResult{
		Running:     true,
		Connections: connStatuses,
	}
	resp, _ := NewSuccessResponse(req.ID, result)
	return resp
}

// runVPNConnection runs a VPN connection with separate API calls
func (s *Server) runVPNConnection(ctx context.Context, sessionID, server, token string) {
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
		if err := os.WriteFile(certFile, []byte(caCert), 0600); err != nil {
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

		pgClient = proxyguard.NewClient("127.0.0.1:51821", wgConfig.ServerEndpoint)
		if err := pgClient.Start(ctx); err != nil {
			fmt.Printf("[Session %s] Failed to start proxyguard: %v\n", sessionID, err)
			return
		}
		defer pgClient.Stop()
		fmt.Printf("[Session %s] ✓ Proxyguard client started\n", sessionID)
	}

	// 4. Start WireGuard device
	fmt.Printf("[Session %s] Starting WireGuard device...\n", sessionID)
	wgDeviceConfig := &wireguard.Config{
		InterfaceName: "utun", // macOS
		ListenPort:    51820,
		MTU:           1420,
	}

	wgDevice, err := wireguard.NewDevice(ctx, wgDeviceConfig)
	if err != nil {
		fmt.Printf("[Session %s] Failed to create WireGuard device: %v\n", sessionID, err)
		return
	}

	if err := wgDevice.Start(ctx); err != nil {
		fmt.Printf("[Session %s] Failed to start WireGuard device: %v\n", sessionID, err)
		return
	}
	defer wgDevice.Close()

	if err := wgDevice.SetConfig(wgConfig.Config); err != nil {
		fmt.Printf("[Session %s] Failed to set WireGuard config: %v\n", sessionID, err)
		return
	}

	// Configure IP address and routes on the WireGuard interface
	fmt.Printf("[Session %s] Configuring network interface...\n", sessionID)
	netCfg := &netconfig.InterfaceConfig{
		InterfaceName: wgDevice.InterfaceName(),
		IPAddress:     wgConfig.AssignedIP + "/24",
		Routes:        []string{"10.17.0.0/24", "10.43.0.0/16"},
		Gateway:       "10.17.0.1",
	}

	if err := netconfig.ConfigureInterface(netCfg); err != nil {
		fmt.Printf("[Session %s] Failed to configure network interface: %v\n", sessionID, err)
		return
	}

	fmt.Printf("[Session %s] ✓ WireGuard device started (IP: %s)\n", sessionID, wgConfig.AssignedIP)

	// 4. Start hosts polling goroutine (polls every 10 seconds)
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

// pollHosts polls the hosts API every 10 seconds and updates /etc/hosts
func (s *Server) pollHosts(ctx context.Context, sessionID string, apiClient *api.Client, done chan<- struct{}) {
	defer close(done)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	// Track current hosts to detect changes
	currentHosts := make(map[string]string) // hostname -> IP

	// Initial fetch
	s.fetchAndUpdateHosts(ctx, sessionID, apiClient, currentHosts)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.fetchAndUpdateHosts(ctx, sessionID, apiClient, currentHosts)
		}
	}
}

// fetchAndUpdateHosts fetches hosts from API and updates /etc/hosts
func (s *Server) fetchAndUpdateHosts(ctx context.Context, sessionID string, apiClient *api.Client, currentHosts map[string]string) {
	hosts, err := apiClient.GetHosts()
	if err != nil {
		fmt.Printf("[Session %s] Warning: Failed to fetch hosts: %v\n", sessionID, err)
		return
	}

	// Build map of new hosts
	newHosts := make(map[string]string)
	for _, host := range hosts {
		newHosts[host.Hostname] = host.IP
	}

	// Remove hosts that no longer exist
	for hostname := range currentHosts {
		if _, exists := newHosts[hostname]; !exists {
			fmt.Printf("[Session %s] Removing host: %s\n", sessionID, hostname)
			if err := s.hostsManager.Remove(hostname); err != nil {
				fmt.Printf("[Session %s] Warning: Failed to remove host %s: %v\n", sessionID, hostname, err)
			}
		}
	}

	// Add or update hosts
	for hostname, ip := range newHosts {
		if currentIP, exists := currentHosts[hostname]; !exists || currentIP != ip {
			fmt.Printf("[Session %s] Adding/updating host: %s -> %s\n", sessionID, hostname, ip)
			if err := s.hostsManager.Add(hostname, ip, fmt.Sprintf("# kltun session %s", sessionID)); err != nil {
				fmt.Printf("[Session %s] Warning: Failed to add host %s: %v\n", sessionID, hostname, err)
			}
		}
	}

	// Update current hosts map
	for k := range currentHosts {
		delete(currentHosts, k)
	}
	for k, v := range newHosts {
		currentHosts[k] = v
	}

	fmt.Printf("[Session %s] ✓ Hosts updated (%d entries)\n", sessionID, len(currentHosts))
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
