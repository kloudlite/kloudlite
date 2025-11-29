package api

import (
	"bytes"
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
	Config         string `json:"config"`          // IPC format configuration
	AssignedIP     string `json:"assigned_ip"`     // Device IP address (e.g., "10.17.0.2")
	PublicKey      string `json:"public_key"`      // Device public key
	ServerEndpoint string `json:"server_endpoint"` // WorkMachine endpoint (e.g., "203.0.113.1:443")
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

// TunnelEndpointResponse contains tunnel endpoint info with hostname and IP
type TunnelEndpointResponse struct {
	TunnelEndpoint string `json:"tunnel_endpoint"` // hostname:443
	Hostname       string `json:"hostname"`        // vpn-connect.{subdomain}.{domain}
	IP             string `json:"ip"`              // Public IP for /etc/hosts
}

// TokenExchangeResponse contains the permanent token from token exchange
type TokenExchangeResponse struct {
	ConnectionToken string `json:"connection_token"` // Long-lived (1 year) permanent token
}

// GetTunnelEndpoint calls the VPN tunnel endpoint API
// Returns hostname, IP for /etc/hosts configuration, and the full endpoint for connection
func (c *Client) GetTunnelEndpoint() (*TunnelEndpointResponse, error) {
	url := fmt.Sprintf("%s/api/vpn/tunnel-endpoint", c.BaseURL)

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

	var result TunnelEndpointResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// ExchangeToken exchanges a short-lived temporary token for a long-lived permanent token
// This calls the /api/vpn/exchange endpoint which returns a 1-year token for VPN connections
func (c *Client) ExchangeToken(temporaryToken string) (*TokenExchangeResponse, error) {
	url := fmt.Sprintf("%s/api/vpn/exchange", c.BaseURL)

	reqBody := map[string]string{"token": temporaryToken}
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
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result TokenExchangeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
