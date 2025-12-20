package handlers

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
)

const (
	// Label used to identify kloudlite environment namespaces
	kloudliteEnvironmentLabel = "kloudlite.io/environment"
	// Annotation for environment owner
	kloudliteCreatedByAnnotation = "kloudlite.io/created-by"
)

// HostsHandler handles hosts configuration requests
type HostsHandler struct {
	logger *zap.Logger
	cache  *HostsCache
}

// HostEntry represents a single hosts file entry
type HostEntry struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Type     string `json:"type,omitempty"` // "ingress" or "service"
}

// HostsResponse represents the response from the hosts endpoint
type HostsResponse struct {
	Hosts []HostEntry `json:"hosts"`
}

// NewHostsHandler creates a new HostsHandler
func NewHostsHandler(logger *zap.Logger, cache *HostsCache) *HostsHandler {
	return &HostsHandler{
		logger: logger,
		cache:  cache,
	}
}

// ServeHTTP implements http.Handler for hosts endpoint
func (h *HostsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get hosts from cache - no K8s API calls!
	hosts := h.cache.GetHosts()

	h.logger.Debug("returning hosts from cache", zap.Int("count", len(hosts)))

	response := HostsResponse{
		Hosts: hosts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode hosts response", zap.Error(err))
	}
}
