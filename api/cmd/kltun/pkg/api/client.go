package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HostEntry represents a host entry from the API
type HostEntry struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
}

// ConnectResponse represents the response from the connect API
type ConnectResponse struct {
	CACert   string      `json:"ca_cert"`
	WGConfig string      `json:"wg_config"`
	Hosts    []HostEntry `json:"hosts"`
}

// Client represents an API client
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Connect calls the VPN connect API endpoint
func (c *Client) Connect() (*ConnectResponse, error) {
	url := fmt.Sprintf("%s/api/vpn/connect", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var connectResp ConnectResponse
	if err := json.NewDecoder(resp.Body).Decode(&connectResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &connectResp, nil
}

// WireGuardConfigResponse represents WireGuard configuration with metadata
type WireGuardConfigResponse struct {
	Config     string `json:"config"`      // IPC format configuration
	AssignedIP string `json:"assigned_ip"` // Device IP address (e.g., "10.17.0.2")
	PublicKey  string `json:"public_key"`  // Device public key
}

// GetWireGuardConfig calls the VPN WireGuard config API endpoint
func (c *Client) GetWireGuardConfig(deviceID string) (*WireGuardConfigResponse, error) {
	url := fmt.Sprintf("%s/api/vpn/wireguard-config?device_id=%s", c.BaseURL, deviceID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result WireGuardConfigResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// GetCACert calls the VPN CA certificate API endpoint
func (c *Client) GetCACert() (string, error) {
	url := fmt.Sprintf("%s/api/vpn/ca-cert", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		CACert string `json:"ca_cert"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.CACert, nil
}

// GetHosts calls the VPN hosts API endpoint
func (c *Client) GetHosts() ([]HostEntry, error) {
	url := fmt.Sprintf("%s/api/vpn/hosts", c.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Hosts []HostEntry `json:"hosts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Hosts, nil
}
