package handlers

import (
	"encoding/json"
	"net/http"
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

// ServeHTTP implements http.Handler for VPN status endpoint
func (h *VPNStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
