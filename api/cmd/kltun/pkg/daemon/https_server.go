package daemon

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// HTTPSServer serves status/health endpoints over HTTPS on the VPN IP
type HTTPSServer struct {
	server    *http.Server
	vpnIP     string
	certPEM   []byte
	keyPEM    []byte
	daemonRef *Server
	running   bool
	mu        sync.Mutex
}

// HTTPSServerStatus represents the status response from the HTTPS server
type HTTPSServerStatus struct {
	Status         string    `json:"status"`          // connected, disconnected, reconnecting
	VPNIP          string    `json:"vpn_ip"`          // VPN IP address
	SessionID      string    `json:"session_id"`      // Connection session ID
	UptimeSeconds  int64     `json:"uptime_seconds"`  // Connection uptime in seconds
	TunnelEndpoint string    `json:"tunnel_endpoint"` // Tunnel server endpoint
	Timestamp      time.Time `json:"timestamp"`       // Current timestamp
}

// NewHTTPSServer creates a new HTTPS server for kltun
func NewHTTPSServer(vpnIP string, certPEM, keyPEM []byte, daemon *Server) *HTTPSServer {
	return &HTTPSServer{
		vpnIP:     vpnIP,
		certPEM:   certPEM,
		keyPEM:    keyPEM,
		daemonRef: daemon,
	}
}

// Start starts the HTTPS server
func (h *HTTPSServer) Start(ctx context.Context) error {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return fmt.Errorf("HTTPS server already running")
	}
	h.running = true
	h.mu.Unlock()

	// Load TLS certificate
	cert, err := tls.X509KeyPair(h.certPEM, h.keyPEM)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	// Create HTTP mux
	mux := http.NewServeMux()
	mux.HandleFunc("/status", h.handleStatus)
	mux.HandleFunc("/health", h.handleHealth)

	// Create server bound to VPN IP
	addr := fmt.Sprintf("%s:443", h.vpnIP)
	h.server = &http.Server{
		Addr:      addr,
		Handler:   mux,
		TLSConfig: tlsConfig,
	}

	fmt.Printf("[HTTPS] Starting server on %s\n", addr)

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		// ListenAndServeTLS with empty strings because certs are in TLSConfig
		if err := h.server.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
		close(errChan)
	}()

	// Wait for context cancellation or error
	select {
	case <-ctx.Done():
		return h.Stop()
	case err := <-errChan:
		h.mu.Lock()
		h.running = false
		h.mu.Unlock()
		return err
	}
}

// Stop stops the HTTPS server
func (h *HTTPSServer) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.running || h.server == nil {
		return nil
	}

	fmt.Printf("[HTTPS] Stopping server on %s:443\n", h.vpnIP)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.server.Shutdown(ctx)
	h.running = false
	return err
}

// handleStatus handles GET /status requests
func (h *HTTPSServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get connection status from daemon
	status := HTTPSServerStatus{
		Status:    "unknown",
		VPNIP:     h.vpnIP,
		Timestamp: time.Now().UTC(),
	}

	if h.daemonRef != nil {
		h.daemonRef.connMutex.RLock()
		for _, conn := range h.daemonRef.connections {
			state := conn.GetState()
			status.SessionID = conn.SessionID
			status.TunnelEndpoint = conn.TunnelEndpoint
			status.UptimeSeconds = int64(time.Since(conn.StartTime).Seconds())

			switch state {
			case StateConnected:
				status.Status = "connected"
			case StateReconnecting:
				status.Status = "reconnecting"
			case StateDisconnected:
				status.Status = "disconnected"
			default:
				status.Status = "unknown"
			}
			break // Only one connection
		}
		h.daemonRef.connMutex.RUnlock()
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// handleHealth handles GET /health requests
func (h *HTTPSServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
