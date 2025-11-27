package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"

	"go.uber.org/zap"
)

// WireGuardHandler handles WireGuard peer management requests
type WireGuardHandler struct {
	logger    *zap.Logger
	device    string
}

// NewWireGuardHandler creates a new WireGuardHandler
func NewWireGuardHandler(logger *zap.Logger, device string) *WireGuardHandler {
	if device == "" {
		device = "wg0"
	}
	return &WireGuardHandler{
		logger: logger,
		device: device,
	}
}

// PublicKeyResponse represents the response for the public key endpoint
type PublicKeyResponse struct {
	PublicKey string `json:"publicKey"`
	Device    string `json:"device"`
}

// CreatePeerRequest represents the request to create a peer
type CreatePeerRequest struct {
	PublicKey           string   `json:"publicKey"`
	AllowedIPs          []string `json:"allowedIPs"`
	PersistentKeepalive int      `json:"persistentKeepalive,omitempty"`
	Endpoint            string   `json:"endpoint,omitempty"`
	PresharedKey        string   `json:"presharedKey,omitempty"`
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

	if req.PublicKey == "" {
		http.Error(w, "publicKey is required", http.StatusBadRequest)
		return
	}

	if len(req.AllowedIPs) == 0 {
		http.Error(w, "allowedIPs is required", http.StatusBadRequest)
		return
	}

	if err := h.createPeer(device, &req); err != nil {
		h.logger.Error("failed to create peer",
			zap.String("device", device),
			zap.String("publicKey", req.PublicKey),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("failed to create peer: %v", err), http.StatusInternalServerError)
		return
	}

	h.logger.Info("peer created successfully",
		zap.String("device", device),
		zap.String("publicKey", req.PublicKey),
		zap.Strings("allowedIPs", req.AllowedIPs))

	response := PeerResponse{
		Success: true,
		Message: "peer created successfully",
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

// createPeer adds a new peer to the WireGuard device
func (h *WireGuardHandler) createPeer(device string, req *CreatePeerRequest) error {
	// Build wg set command
	// wg set <device> peer <public-key> allowed-ips <ip1>,<ip2> [persistent-keepalive <seconds>] [endpoint <endpoint>] [preshared-key <key>]
	args := []string{"set", device, "peer", req.PublicKey, "allowed-ips", strings.Join(req.AllowedIPs, ",")}

	if req.PersistentKeepalive > 0 {
		args = append(args, "persistent-keepalive", fmt.Sprintf("%d", req.PersistentKeepalive))
	}

	if req.Endpoint != "" {
		args = append(args, "endpoint", req.Endpoint)
	}

	if req.PresharedKey != "" {
		args = append(args, "preshared-key", "/dev/stdin")
	}

	cmd := exec.Command("wg", args...)

	if req.PresharedKey != "" {
		cmd.Stdin = strings.NewReader(req.PresharedKey)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wg set failed: %s: %w", string(output), err)
	}

	return nil
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
