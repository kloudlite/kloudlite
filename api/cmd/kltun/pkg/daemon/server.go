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
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/hosts"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/netconfig"
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/wgkeys"
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

	// HTTPS server for status/health endpoints (daemon-level, not per-connection)
	httpsServer     *HTTPSServer
	httpsServerOnce sync.Once
	tlsCertPEM      []byte
	tlsKeyPEM       []byte
}

// ConnectionState represents the state of a VPN connection
type ConnectionState string

const (
	StateConnected    ConnectionState = "connected"
	StateDisconnected ConnectionState = "disconnected"
	StateReconnecting ConnectionState = "reconnecting"
)

// VPNConnection represents an active VPN connection
type VPNConnection struct {
	SessionID  string
	Server     string
	StartTime  time.Time
	CancelFunc context.CancelFunc
	DoneChan   chan struct{} // Signals when cleanup is complete

	// Connection state for auto-reconnect
	State     ConnectionState
	StateLock sync.RWMutex

	// Credentials for reconnection (stored after initial connect)
	DashboardServer string // Dashboard server URL (e.g., https://beanbag.khost.dev)
	PermanentToken  string
	TunnelEndpoint  string
	TunnelInfo      *api.TunnelEndpointResponse
	DeviceID        string
	KeyPair         *wgkeys.KeyPair

	// WireGuard device and network config (for cleanup during reconnection)
	WireGuardDevice *wireguard.Device
	NetConfig       *netconfig.InterfaceConfig
	WGMutex         sync.Mutex // Protects WireGuardDevice and NetConfig

	// Reconnection control
	ReconnectChan chan struct{} // Signal to trigger reconnection attempt

	// VPN IP address (e.g., "10.17.0.2")
	VPNIP string
}

// GetState returns the current connection state
func (c *VPNConnection) GetState() ConnectionState {
	c.StateLock.RLock()
	defer c.StateLock.RUnlock()
	return c.State
}

// SetState sets the connection state
func (c *VPNConnection) SetState(state ConnectionState) {
	c.StateLock.Lock()
	defer c.StateLock.Unlock()
	c.State = state
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
	// Create platform-specific listener (Unix socket on Unix, named pipe on Windows)
	listener, err := CreateListener(socketPath)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
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

	// Stop HTTPS server if running
	if s.httpsServer != nil {
		fmt.Println("Stopping HTTPS status server...")
		if err := s.httpsServer.Stop(); err != nil {
			fmt.Printf("Warning: Failed to stop HTTPS server: %v\n", err)
		}
	}

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

// StartHTTPSServer starts the HTTPS status server on 127.0.0.1:443
// This is called once when the first VPN connection is established and TLS certs are available
// The server runs for the lifetime of the daemon, regardless of VPN connection state
func (s *Server) StartHTTPSServer(certPEM, keyPEM []byte) {
	s.httpsServerOnce.Do(func() {
		s.tlsCertPEM = certPEM
		s.tlsKeyPEM = keyPEM
		s.httpsServer = NewHTTPSServer("127.0.0.1", certPEM, keyPEM, s)

		go func() {
			// Create a long-lived context for the HTTPS server
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Store cancel func so Stop() can use it if needed
			go func() {
				<-s.shutdownCh
				cancel()
			}()

			if err := s.httpsServer.Start(ctx); err != nil {
				fmt.Printf("[HTTPS] Server error: %v\n", err)
			}
		}()

		fmt.Println("[HTTPS] Status server started on 127.0.0.1:443")
	})
}
