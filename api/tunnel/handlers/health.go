package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	state  *ServerState
	logger *zap.Logger
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(state *ServerState, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		state:  state,
		logger: logger,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status            string    `json:"status"`
	Uptime            string    `json:"uptime"`
	UptimeSeconds     float64   `json:"uptime_seconds"`
	StartTime         time.Time `json:"start_time"`
	ActiveConnections int64     `json:"active_connections"`
	TotalConnections  int64     `json:"total_connections"`
	BytesReceived     uint64    `json:"bytes_received"`
	BytesSent         uint64    `json:"bytes_sent"`
	ServerTime        time.Time `json:"server_time"`
}

// ServeHTTP implements http.Handler
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	uptime := h.state.GetUptime()
	response := HealthResponse{
		Status:            "ok",
		Uptime:            uptime.String(),
		UptimeSeconds:     uptime.Seconds(),
		StartTime:         h.state.GetStartTime(),
		ActiveConnections: h.state.GetActiveConnections(),
		TotalConnections:  h.state.GetTotalConnections(),
		BytesReceived:     h.state.GetBytesReceived(),
		BytesSent:         h.state.GetBytesSent(),
		ServerTime:        time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode health response", zap.Error(err))
	}
}
