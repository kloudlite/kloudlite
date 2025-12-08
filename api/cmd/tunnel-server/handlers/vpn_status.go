package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// VPNStatusHandler handles VPN connectivity check requests
type VPNStatusHandler struct {
	logger    *zap.Logger
	startTime time.Time
}

// VPNStatusResponse represents the response from the VPN status endpoint
type VPNStatusResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	Uptime    string `json:"uptime"`
}

// NewVPNStatusHandler creates a new VPNStatusHandler
func NewVPNStatusHandler(logger *zap.Logger) *VPNStatusHandler {
	return &VPNStatusHandler{
		logger:    logger,
		startTime: time.Now(),
	}
}

// setCORSHeaders adds CORS headers to allow cross-origin requests from dashboard
func (h *VPNStatusHandler) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	// Allow requests from *.khost.dev domains
	if origin != "" && strings.HasSuffix(origin, ".khost.dev") {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400")
	}
}

// ServeHTTP implements http.Handler for VPN status endpoint
func (h *VPNStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers for all responses
	h.setCORSHeaders(w, r)

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uptime := time.Since(h.startTime).Round(time.Second)

	response := VPNStatusResponse{
		Status:    "connected",
		Message:   "VPN tunnel is active",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    uptime.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode VPN status response", zap.Error(err))
	}
}

// WithCORS wraps a handler with CORS support for OPTIONS preflight
// This handles OPTIONS before JWT middleware and adds CORS headers
func WithCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Add CORS headers for allowed origins
		if origin != "" && strings.HasSuffix(origin, ".khost.dev") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Access-Control-Allow-Credentials", "true") // Allow cookies
			w.Header().Set("Access-Control-Max-Age", "86400")
		}

		// Handle OPTIONS preflight request (no auth needed)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Pass to next handler (which may include JWT middleware)
		next.ServeHTTP(w, r)
	})
}
