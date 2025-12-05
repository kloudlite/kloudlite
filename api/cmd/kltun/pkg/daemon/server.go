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
	"github.com/kloudlite/kloudlite/api/cmd/kltun/pkg/wgkeys"
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
	PermanentToken string
	TunnelEndpoint string
	TunnelInfo     *api.TunnelEndpointResponse
	DeviceID       string
	KeyPair        *wgkeys.KeyPair

	// Reconnection control
	ReconnectChan chan struct{} // Signal to trigger reconnection attempt
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
	if err := os.Chmod(socketPath, 0o666); err != nil {
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
