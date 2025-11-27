package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TunnelClient is a client for direct communication with the tunnel server
type TunnelClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewTunnelClient creates a new tunnel server client
// endpoint should be in format "ip:port" (e.g., "203.0.113.1:443")
func NewTunnelClient(endpoint string) *TunnelClient {
	return &TunnelClient{
		BaseURL: fmt.Sprintf("https://%s", endpoint),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, // Tunnel server uses self-signed cert initially
					MinVersion:         tls.VersionTLS13,
				},
			},
		},
	}
}

// CreatePeerRequest represents the request to create a WireGuard peer
type CreatePeerRequest struct {
	DeviceName string `json:"deviceName"`
}

// CreatePeerResponse represents the response from creating a peer
type CreatePeerResponse struct {
	Success   bool   `json:"success"`
	PublicKey string `json:"publicKey"`
	IP        string `json:"ip"`
	Config    string `json:"config"` // Full WireGuard config for the client
}

// CreatePeer creates a new WireGuard peer on the tunnel server
func (c *TunnelClient) CreatePeer(deviceName string) (*CreatePeerResponse, error) {
	url := fmt.Sprintf("%s/wg/peer", c.BaseURL)

	reqBody := CreatePeerRequest{DeviceName: deviceName}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tunnel server returned status %d: %s", resp.StatusCode, string(body))
	}

	var result CreatePeerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DeletePeerRequest represents the request to delete a WireGuard peer
type DeletePeerRequest struct {
	PublicKey string `json:"publicKey"`
}

// DeletePeer deletes a WireGuard peer from the tunnel server
func (c *TunnelClient) DeletePeer(publicKey string) error {
	url := fmt.Sprintf("%s/wg/peer", c.BaseURL)

	reqBody := DeletePeerRequest{PublicKey: publicKey}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tunnel server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetPublicKeyResponse represents the response from getting the server's public key
type GetPublicKeyResponse struct {
	PublicKey string `json:"publicKey"`
	Device    string `json:"device"`
}

// GetPublicKey gets the tunnel server's WireGuard public key
func (c *TunnelClient) GetPublicKey() (*GetPublicKeyResponse, error) {
	url := fmt.Sprintf("%s/wg/public-key", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tunnel server returned status %d: %s", resp.StatusCode, string(body))
	}

	var result GetPublicKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetCACert gets the CA certificate from the tunnel server
func (c *TunnelClient) GetCACert() (string, error) {
	url := fmt.Sprintf("%s/ca-cert", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("tunnel server returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		CACert string `json:"ca_cert"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.CACert, nil
}

// GetHosts gets the hosts entries from the tunnel server
func (c *TunnelClient) GetHosts() ([]HostEntry, error) {
	url := fmt.Sprintf("%s/hosts", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tunnel server returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Hosts []HostEntry `json:"hosts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Hosts, nil
}

// Health checks if the tunnel server is healthy
func (c *TunnelClient) Health() error {
	url := fmt.Sprintf("%s/health", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("tunnel server returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
