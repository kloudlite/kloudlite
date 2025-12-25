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

// DaemonStatus represents the daemon's running state
type DaemonStatus struct {
	Running       bool      `json:"running"`
	StartedAt     time.Time `json:"started_at"`
	UptimeSeconds int64     `json:"uptime_seconds"`
}

// VPNStatus represents the VPN connection state
type VPNStatus struct {
	Status                  string `json:"status"`                     // idle, connecting, connected, reconnecting, disconnected
	StatusMessage           string `json:"status_message"`             // Human-readable status message
	VPNIP                   string `json:"vpn_ip,omitempty"`           // VPN IP address (when connected)
	SessionID               string `json:"session_id,omitempty"`       // Connection session ID
	ConnectionUptimeSeconds int64  `json:"connection_uptime_seconds"`  // Connection uptime in seconds
	TunnelEndpoint          string `json:"tunnel_endpoint,omitempty"`  // Tunnel server endpoint
	DashboardServer         string `json:"dashboard_server,omitempty"` // Dashboard server URL
}

// StatusResponse represents the complete status response
type StatusResponse struct {
	Daemon    DaemonStatus `json:"daemon"`
	VPN       VPNStatus    `json:"vpn"`
	Timestamp time.Time    `json:"timestamp"`
}

// getStatusMessage returns a human-readable message for the VPN status
func getStatusMessage(status string) string {
	switch status {
	case "idle":
		return "No VPN connection configured"
	case "connecting":
		return "Establishing VPN connection..."
	case "connected":
		return "VPN tunnel active"
	case "reconnecting":
		return "Connection lost, attempting to reconnect..."
	case "disconnected":
		return "VPN connection terminated"
	default:
		return "Unknown status"
	}
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

	// Create HTTP mux with CORS middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/status", h.corsMiddleware(h.handleStatus))
	mux.HandleFunc("/health", h.corsMiddleware(h.handleHealth))

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

	now := time.Now().UTC()

	// Build daemon status
	daemonStatus := DaemonStatus{
		Running: true, // If we're responding, daemon is running
	}

	if h.daemonRef != nil {
		daemonStatus.StartedAt = h.daemonRef.startedAt
		daemonStatus.UptimeSeconds = int64(now.Sub(h.daemonRef.startedAt).Seconds())
	}

	// Build VPN status
	vpnStatus := VPNStatus{
		Status: "idle",
	}

	if h.daemonRef != nil {
		h.daemonRef.connMutex.RLock()
		if len(h.daemonRef.connections) == 0 {
			vpnStatus.Status = "idle"
		} else {
			for _, conn := range h.daemonRef.connections {
				state := conn.GetState()
				vpnStatus.SessionID = conn.SessionID
				vpnStatus.TunnelEndpoint = conn.TunnelEndpoint
				vpnStatus.DashboardServer = conn.DashboardServer
				vpnStatus.VPNIP = conn.VPNIP

				if conn.StartTime.IsZero() {
					vpnStatus.ConnectionUptimeSeconds = 0
				} else {
					vpnStatus.ConnectionUptimeSeconds = int64(now.Sub(conn.StartTime).Seconds())
				}

				switch state {
				case StateConnected:
					vpnStatus.Status = "connected"
				case StateReconnecting:
					vpnStatus.Status = "reconnecting"
				case StateDisconnected:
					vpnStatus.Status = "disconnected"
				default:
					vpnStatus.Status = "idle"
				}
				break // Only one connection
			}
		}
		h.daemonRef.connMutex.RUnlock()
	}

	vpnStatus.StatusMessage = getStatusMessage(vpnStatus.Status)

	// Build complete response
	response := StatusResponse{
		Daemon:    daemonStatus,
		VPN:       vpnStatus,
		Timestamp: now,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
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

// corsMiddleware adds CORS headers to allow cross-origin requests from dashboard
func (h *HTTPSServer) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allow requests from any origin (the dashboard could be on various subdomains)
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")
		// Required for Chrome's Private Network Access (PNA) to allow public sites
		// to make requests to local/private network addresses
		w.Header().Set("Access-Control-Allow-Private-Network", "true")

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}
