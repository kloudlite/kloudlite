package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// PeerInfo stores information about a registered peer
type PeerInfo struct {
	PublicKey  string `json:"publicKey"`
	IP         string `json:"ip"`
	DeviceName string `json:"deviceName"`
}

// WireGuardHandler handles WireGuard peer management requests
type WireGuardHandler struct {
	logger        *zap.Logger
	device        string
	cidr          string
	serverAddress string // Server's WireGuard address (e.g., 10.17.0.1)
	endpoint      string // Server's public endpoint (e.g., tunnel.example.com:443)
	storagePath   string // Path to persist peers

	// IP allocation tracking
	mu    sync.Mutex
	peers map[string]*PeerInfo // publicKey -> PeerInfo
}

// WireGuardHandlerConfig holds configuration for the WireGuard handler
type WireGuardHandlerConfig struct {
	Device        string
	CIDR          string // e.g., "10.17.0.0/24"
	ServerAddress string // e.g., "10.17.0.1"
	Endpoint      string // e.g., "tunnel.example.com:443"
	StoragePath   string // e.g., "/var/lib/tunnel-server/peers.json"
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
	if cfg.StoragePath == "" {
		cfg.StoragePath = "/var/lib/tunnel-server/peers.json"
	}

	h := &WireGuardHandler{
		logger:        logger,
		device:        cfg.Device,
		cidr:          cfg.CIDR,
		serverAddress: cfg.ServerAddress,
		endpoint:      cfg.Endpoint,
		storagePath:   cfg.StoragePath,
		peers:         make(map[string]*PeerInfo),
	}

	// Load persisted peers on startup
	if err := h.loadPeers(); err != nil {
		logger.Warn("failed to load persisted peers", zap.Error(err))
	}

	return h
}

// PublicKeyResponse represents the response for the public key endpoint
type PublicKeyResponse struct {
	PublicKey string `json:"publicKey"`
	Device    string `json:"device"`
}

// CreatePeerRequest represents the request to create a peer
type CreatePeerRequest struct {
	DeviceName string `json:"deviceName"`
	PublicKey  string `json:"publicKey"` // Client's WireGuard public key
}

// CreatePeerResponse represents the response for peer creation
type CreatePeerResponse struct {
	Success         bool   `json:"success"`
	IP              string `json:"ip"`
	ServerPublicKey string `json:"serverPublicKey"` // Server's WireGuard public key
	CIDR            string `json:"cidr"`            // VPN CIDR (e.g., "10.17.0.0/24")
	AlreadyExists   bool   `json:"alreadyExists"`   // True if peer already existed
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

	if req.PublicKey == "" {
		http.Error(w, "publicKey is required", http.StatusBadRequest)
		return
	}

	// Check if peer already exists
	existingPeer, alreadyExists := h.getPeer(req.PublicKey)

	var peerIP string
	if alreadyExists {
		// Peer already exists, return existing IP
		peerIP = existingPeer.IP
		h.logger.Info("peer already exists, returning existing configuration",
			zap.String("device", device),
			zap.String("deviceName", req.DeviceName),
			zap.String("publicKey", req.PublicKey),
			zap.String("ip", peerIP))
	} else {
		// Allocate an IP for the new peer
		var err error
		peerIP, err = h.allocateIP(req.PublicKey, req.DeviceName)
		if err != nil {
			h.logger.Error("failed to allocate IP", zap.Error(err))
			http.Error(w, fmt.Sprintf("failed to allocate IP: %v", err), http.StatusInternalServerError)
			return
		}

		// Add peer to WireGuard
		if err := h.addPeer(device, req.PublicKey, peerIP); err != nil {
			h.logger.Error("failed to add peer",
				zap.String("device", device),
				zap.String("publicKey", req.PublicKey),
				zap.Error(err))
			// Release the allocated IP on failure
			h.removePeer(req.PublicKey)
			http.Error(w, fmt.Sprintf("failed to add peer: %v", err), http.StatusInternalServerError)
			return
		}

		// Persist peers to disk
		if err := h.savePeers(); err != nil {
			h.logger.Error("failed to persist peers", zap.Error(err))
		}

		h.logger.Info("peer created successfully",
			zap.String("device", device),
			zap.String("deviceName", req.DeviceName),
			zap.String("publicKey", req.PublicKey),
			zap.String("ip", peerIP))
	}

	// Get server's public key
	serverPublicKey, err := h.getPublicKey(device)
	if err != nil {
		h.logger.Error("failed to get server public key", zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to get server public key: %v", err), http.StatusInternalServerError)
		return
	}

	response := CreatePeerResponse{
		Success:         true,
		IP:              peerIP,
		ServerPublicKey: serverPublicKey,
		CIDR:            h.cidr,
		AlreadyExists:   alreadyExists,
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

	// Remove peer from map and persist
	h.removePeer(req.PublicKey)
	if err := h.savePeers(); err != nil {
		h.logger.Error("failed to persist peers after delete", zap.Error(err))
	}

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

// allocateIP allocates an IP address from the CIDR range and stores peer info
func (h *WireGuardHandler) allocateIP(publicKey, deviceName string) (string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if this public key already has an IP
	if peer, exists := h.peers[publicKey]; exists {
		return peer.IP, nil
	}

	// Parse CIDR
	_, ipNet, err := net.ParseCIDR(h.cidr)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR: %w", err)
	}

	// Get used IPs
	usedIPs := make(map[string]bool)
	usedIPs[h.serverAddress] = true // Server's IP is always used

	for _, peer := range h.peers {
		usedIPs[peer.IP] = true
	}

	// Also check WireGuard for existing peers not in our map
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
			h.peers[publicKey] = &PeerInfo{
				PublicKey:  publicKey,
				IP:         ipStr,
				DeviceName: deviceName,
			}
			return ipStr, nil
		}
	}

	return "", fmt.Errorf("no available IP addresses in CIDR %s", h.cidr)
}

// removePeer removes a peer from the in-memory map
func (h *WireGuardHandler) removePeer(publicKey string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.peers, publicKey)
}

// getPeer checks if a peer with the given public key exists and returns its info
func (h *WireGuardHandler) getPeer(publicKey string) (*PeerInfo, bool) {
	h.mu.Lock()
	peer, exists := h.peers[publicKey]
	h.mu.Unlock()

	if exists {
		return peer, true
	}

	// Also check WireGuard directly in case the map is out of sync
	cmd := exec.Command("wg", "show", h.device, "allowed-ips")
	output, err := cmd.Output()
	if err != nil {
		return nil, false
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[0] == publicKey {
			// Found the peer, extract IP
			allowedIPs := strings.Split(parts[1], ",")
			if len(allowedIPs) > 0 {
				ip := strings.Split(allowedIPs[0], "/")[0]
				if ip != "" {
					// Update our map for future lookups
					peer := &PeerInfo{
						PublicKey:  publicKey,
						IP:         ip,
						DeviceName: "unknown",
					}
					h.mu.Lock()
					h.peers[publicKey] = peer
					h.mu.Unlock()
					return peer, true
				}
			}
		}
	}

	return nil, false
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

// loadPeers loads peers from the persistent storage file and re-adds them to WireGuard
func (h *WireGuardHandler) loadPeers() error {
	data, err := os.ReadFile(h.storagePath)
	if err != nil {
		if os.IsNotExist(err) {
			h.logger.Info("no existing peers file found, starting fresh")
			return nil
		}
		return fmt.Errorf("failed to read peers file: %w", err)
	}

	var peers []*PeerInfo
	if err := json.Unmarshal(data, &peers); err != nil {
		return fmt.Errorf("failed to parse peers file: %w", err)
	}

	h.mu.Lock()
	for _, peer := range peers {
		h.peers[peer.PublicKey] = peer
	}
	h.mu.Unlock()

	// Re-add peers to WireGuard
	for _, peer := range peers {
		if err := h.addPeer(h.device, peer.PublicKey, peer.IP); err != nil {
			h.logger.Warn("failed to re-add peer to WireGuard",
				zap.String("publicKey", peer.PublicKey),
				zap.String("ip", peer.IP),
				zap.Error(err))
		} else {
			h.logger.Info("restored peer from storage",
				zap.String("deviceName", peer.DeviceName),
				zap.String("ip", peer.IP))
		}
	}

	h.logger.Info("loaded peers from storage", zap.Int("count", len(peers)))
	return nil
}

// savePeers persists peers to the storage file
func (h *WireGuardHandler) savePeers() error {
	h.mu.Lock()
	peers := make([]*PeerInfo, 0, len(h.peers))
	for _, peer := range h.peers {
		peers = append(peers, peer)
	}
	h.mu.Unlock()

	data, err := json.MarshalIndent(peers, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal peers: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(h.storagePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Write atomically by writing to temp file first
	tmpPath := h.storagePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write peers file: %w", err)
	}

	if err := os.Rename(tmpPath, h.storagePath); err != nil {
		return fmt.Errorf("failed to rename peers file: %w", err)
	}

	h.logger.Debug("persisted peers to storage", zap.Int("count", len(peers)))
	return nil
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
