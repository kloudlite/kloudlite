package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// WireGuardHandler handles WireGuard peer management requests
type WireGuardHandler struct {
	logger        *zap.Logger
	device        string
	cidr          string
	serverAddress string // Server's WireGuard address (e.g., 10.17.0.1)
	endpoint      string // Server's public endpoint (e.g., tunnel.example.com:443)

	// IP allocation tracking
	mu           sync.Mutex
	allocatedIPs map[string]string // publicKey -> allocated IP
}

// WireGuardHandlerConfig holds configuration for the WireGuard handler
type WireGuardHandlerConfig struct {
	Device        string
	CIDR          string // e.g., "10.17.0.0/24"
	ServerAddress string // e.g., "10.17.0.1"
	Endpoint      string // e.g., "tunnel.example.com:443"
}

// NewWireGuardHandler creates a new WireGuardHandler
func NewWireGuardHandler(logger *zap.Logger, cfg WireGuardHandlerConfig) *WireGuardHandler {
	if cfg.Device == "" {
		cfg.Device = "wg0"
	}
	if cfg.CIDR == "" {
		cfg.CIDR = "10.17.0.0/24"
	}
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = "10.17.0.1"
	}

	return &WireGuardHandler{
		logger:        logger,
		device:        cfg.Device,
		cidr:          cfg.CIDR,
		serverAddress: cfg.ServerAddress,
		endpoint:      cfg.Endpoint,
		allocatedIPs:  make(map[string]string),
	}
}

// PublicKeyResponse represents the response for the public key endpoint
type PublicKeyResponse struct {
	PublicKey string `json:"publicKey"`
	Device    string `json:"device"`
}

// CreatePeerRequest represents the request to create a peer
type CreatePeerRequest struct {
	DeviceName string `json:"deviceName"`
}

// CreatePeerResponse represents the response with full WireGuard config
type CreatePeerResponse struct {
	Success   bool   `json:"success"`
	PublicKey string `json:"publicKey"`
	IP        string `json:"ip"`
	Config    string `json:"config"` // Full WireGuard config for the client
}

// DeletePeerRequest represents the request to delete a peer
type DeletePeerRequest struct {
	PublicKey string `json:"publicKey"`
}

// PeerResponse represents a generic peer operation response
type PeerResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetPublicKeyHandler returns an http.HandlerFunc that returns the server's public key
func (h *WireGuardHandler) GetPublicKeyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		device := r.URL.Query().Get("device")
		if device == "" {
			device = h.device
		}

		publicKey, err := h.getPublicKey(device)
		if err != nil {
			h.logger.Error("failed to get public key", zap.String("device", device), zap.Error(err))
			http.Error(w, fmt.Sprintf("failed to get public key: %v", err), http.StatusInternalServerError)
			return
		}

		response := PublicKeyResponse{
			PublicKey: publicKey,
			Device:    device,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			h.logger.Error("failed to encode response", zap.Error(err))
		}
	}
}

// PeerHandler returns an http.HandlerFunc that handles peer operations (POST for create, DELETE for delete)
func (h *WireGuardHandler) PeerHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		device := r.URL.Query().Get("device")
		if device == "" {
			device = h.device
		}

		switch r.Method {
		case http.MethodPost:
			h.handleCreatePeer(w, r, device)
		case http.MethodDelete:
			h.handleDeletePeer(w, r, device)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func (h *WireGuardHandler) handleCreatePeer(w http.ResponseWriter, r *http.Request, device string) {
	var req CreatePeerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.DeviceName == "" {
		http.Error(w, "deviceName is required", http.StatusBadRequest)
		return
	}

	// Generate key pair for the peer
	privateKey, publicKey, err := h.generateKeyPair()
	if err != nil {
		h.logger.Error("failed to generate key pair", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to generate key pair: %v", err), http.StatusInternalServerError)
		return
	}

	// Allocate an IP for the peer
	peerIP, err := h.allocateIP(publicKey)
	if err != nil {
		h.logger.Error("failed to allocate IP", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to allocate IP: %v", err), http.StatusInternalServerError)
		return
	}

	// Add peer to WireGuard
	if err := h.addPeer(device, publicKey, peerIP); err != nil {
		h.logger.Error("failed to add peer",
			zap.String("device", device),
			zap.String("publicKey", publicKey),
			zap.Error(err))
		// Release the allocated IP on failure
		h.releaseIP(publicKey)
		http.Error(w, fmt.Sprintf("failed to add peer: %v", err), http.StatusInternalServerError)
		return
	}

	// Get server's public key for the config
	serverPublicKey, err := h.getPublicKey(device)
	if err != nil {
		h.logger.Error("failed to get server public key", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to get server public key: %v", err), http.StatusInternalServerError)
		return
	}

	// Generate client WireGuard config
	clientConfig := h.generateClientConfig(privateKey, peerIP, serverPublicKey)

	h.logger.Info("peer created successfully",
		zap.String("device", device),
		zap.String("deviceName", req.DeviceName),
		zap.String("publicKey", publicKey),
		zap.String("ip", peerIP))

	response := CreatePeerResponse{
		Success:   true,
		PublicKey: publicKey,
		IP:        peerIP,
		Config:    clientConfig,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *WireGuardHandler) handleDeletePeer(w http.ResponseWriter, r *http.Request, device string) {
	var req DeletePeerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	if req.PublicKey == "" {
		http.Error(w, "publicKey is required", http.StatusBadRequest)
		return
	}

	if err := h.deletePeer(device, req.PublicKey); err != nil {
		h.logger.Error("failed to delete peer",
			zap.String("device", device),
			zap.String("publicKey", req.PublicKey),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to delete peer: %v", err), http.StatusInternalServerError)
		return
	}

	// Release the allocated IP
	h.releaseIP(req.PublicKey)

	h.logger.Info("peer deleted successfully",
		zap.String("device", device),
		zap.String("publicKey", req.PublicKey))

	response := PeerResponse{
		Success: true,
		Message: "peer deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// generateKeyPair generates a WireGuard key pair
func (h *WireGuardHandler) generateKeyPair() (privateKey, publicKey string, err error) {
	// Generate private key
	privCmd := exec.Command("wg", "genkey")
	privOutput, err := privCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate private key: %w", err)
	}
	privateKey = strings.TrimSpace(string(privOutput))

	// Derive public key from private key
	pubCmd := exec.Command("wg", "pubkey")
	pubCmd.Stdin = strings.NewReader(privateKey)
	pubOutput, err := pubCmd.Output()
	if err != nil {
		return "", "", fmt.Errorf("failed to derive public key: %w", err)
	}
	publicKey = strings.TrimSpace(string(pubOutput))

	return privateKey, publicKey, nil
}

// allocateIP allocates an IP address from the CIDR range
func (h *WireGuardHandler) allocateIP(publicKey string) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if this public key already has an IP
	if ip, exists := h.allocatedIPs[publicKey]; exists {
		return ip, nil
	}

	// Parse CIDR
	_, ipNet, err := net.ParseCIDR(h.cidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR: %w", err)
	}

	// Get existing peers to find used IPs
	usedIPs := make(map[string]bool)
	usedIPs[h.serverAddress] = true // Server's IP is always used

	for _, ip := range h.allocatedIPs {
		usedIPs[ip] = true
	}

	// Also check WireGuard for existing peers
	existingPeers, err := h.getExistingPeerIPs()
	if err != nil {
		h.logger.Warn("failed to get existing peer IPs, continuing with in-memory allocation", zap.Error(err))
	}
	for _, ip := range existingPeers {
		usedIPs[ip] = true
	}

	// Find first available IP (skip network address and server address)
	ip := ipNet.IP.Mask(ipNet.Mask)
	for i := 0; i < 256; i++ {
		// Increment IP
		ip = incrementIP(ip)

		// Check if IP is still in the network
		if !ipNet.Contains(ip) {
			break
		}

		ipStr := ip.String()
		if !usedIPs[ipStr] {
			h.allocatedIPs[publicKey] = ipStr
			return ipStr, nil
		}
	}

	return "", fmt.Errorf("no available IP addresses in CIDR %s", h.cidr)
}

// releaseIP releases an allocated IP address
func (h *WireGuardHandler) releaseIP(publicKey string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.allocatedIPs, publicKey)
}

// getExistingPeerIPs retrieves IPs of existing peers from WireGuard
func (h *WireGuardHandler) getExistingPeerIPs() ([]string, error) {
	cmd := exec.Command("wg", "show", h.device, "allowed-ips")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var ips []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			// Format: <public-key>\t<allowed-ips>
			allowedIPs := strings.Split(parts[1], ",")
			for _, allowedIP := range allowedIPs {
				// Remove CIDR notation if present
				ip := strings.Split(allowedIP, "/")[0]
				if ip != "" {
					ips = append(ips, ip)
				}
			}
		}
	}

	return ips, nil
}

// incrementIP increments an IP address by 1
func incrementIP(ip net.IP) net.IP {
	result := make(net.IP, len(ip))
	copy(result, ip)

	for i := len(result) - 1; i >= 0; i-- {
		result[i]++
		if result[i] != 0 {
			break
		}
	}

	return result
}

// addPeer adds a peer to the WireGuard device
func (h *WireGuardHandler) addPeer(device, publicKey, peerIP string) error {
	// wg set <device> peer <public-key> allowed-ips <ip>/32
	allowedIP := fmt.Sprintf("%s/32", peerIP)
	cmd := exec.Command("wg", "set", device, "peer", publicKey, "allowed-ips", allowedIP)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wg set failed: %s: %w", string(output), err)
	}
	return nil
}

// generateClientConfig generates a WireGuard config for the client
func (h *WireGuardHandler) generateClientConfig(privateKey, peerIP, serverPublicKey string) string {
	config := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/32

[Peer]
PublicKey = %s
AllowedIPs = %s
Endpoint = %s
PersistentKeepalive = 25
`, privateKey, peerIP, serverPublicKey, h.cidr, h.endpoint)

	return config
}

// getPublicKey retrieves the public key for the specified WireGuard device
func (h *WireGuardHandler) getPublicKey(device string) (string, error) {
	// wg show <device> public-key
	cmd := exec.Command("wg", "show", device, "public-key")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("wg show failed: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// deletePeer removes a peer from the WireGuard device
func (h *WireGuardHandler) deletePeer(device string, publicKey string) error {
	// wg set <device> peer <public-key> remove
	cmd := exec.Command("wg", "set", device, "peer", publicKey, "remove")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wg set failed: %s: %w", string(output), err)
	}
	return nil
}

// GeneratePresharedKey generates a WireGuard preshared key (optional, for extra security)
func (h *WireGuardHandler) GeneratePresharedKey() (string, error) {
	cmd := exec.Command("wg", "genpsk")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to generate preshared key: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// EncodeBase64 encodes a string to base64 (utility for key encoding)
func EncodeBase64(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}
