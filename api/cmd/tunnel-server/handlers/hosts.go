package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

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
	logger    *zap.Logger
	cache     *HostsCache
	subdomain string
	domain    string
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
	// Parse HOSTED_SUBDOMAIN env var (e.g., "beanbag.khost.dev")
	var subdomain, domain string
	hostedSubdomain := os.Getenv("HOSTED_SUBDOMAIN")
	if hostedSubdomain != "" {
		parts := strings.SplitN(hostedSubdomain, ".", 2)
		if len(parts) == 2 {
			subdomain = parts[0]
			domain = parts[1]
		}
	}

	return &HostsHandler{
		logger:    logger,
		cache:     cache,
		subdomain: subdomain,
		domain:    domain,
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

	// Get client's VPN IP from request source address
	// The request comes over WireGuard, so RemoteAddr is the VPN IP
	clientIP := getClientVPNIP(r)

	// Add vpn-check entry with client's VPN IP
	// This allows dashboard to verify kltun HTTPS server is reachable
	if clientIP != "" && h.subdomain != "" && h.domain != "" {
		vpnCheckHost := fmt.Sprintf("vpn-check.%s.%s", h.subdomain, h.domain)
		hosts = append(hosts, HostEntry{
			Hostname: vpnCheckHost,
			IP:       clientIP,
			Type:     "vpn-check",
		})
		h.logger.Debug("added vpn-check entry with client VPN IP",
			zap.String("client_ip", clientIP),
			zap.String("hostname", vpnCheckHost))
	}

	h.logger.Debug("returning hosts from cache", zap.Int("count", len(hosts)))

	response := HostsResponse{
		Hosts: hosts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode hosts response", zap.Error(err))
	}
}

// getClientVPNIP extracts the client's VPN IP from the request
func getClientVPNIP(r *http.Request) string {
	// RemoteAddr is in format "ip:port"
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ""
	}

	// Only return if it's a VPN IP (10.17.x.x)
	ip := net.ParseIP(host)
	if ip == nil {
		return ""
	}

	// Check if it's in the VPN CIDR (10.17.0.0/24)
	_, vpnCIDR, _ := net.ParseCIDR("10.17.0.0/24")
	if vpnCIDR != nil && vpnCIDR.Contains(ip) {
		return host
	}

	return ""
}
