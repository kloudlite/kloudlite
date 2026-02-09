package console

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// ConsoleBaseURL is the base URL for the console API
	// TODO: Make this configurable via flag or environment variable
	ConsoleBaseURL = "https://console.kloudlite.io"
	KhostDomain    = "khost.dev"
)

// Client is the console API client
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new console API client
func NewClient() *Client {
	return &Client{
		baseURL: ConsoleBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewClientWithBase creates a new console API client with a custom base URL
func NewClientWithBase(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CheckSubdomainAvailabilityResponse is the response from check-domain-kli
type CheckSubdomainAvailabilityResponse struct {
	Available bool   `json:"available"`
	Subdomain string `json:"subdomain"`
	Reason    string `json:"reason,omitempty"` // "reserved", "invalid", "taken"
	Error     string `json:"error,omitempty"`
}

// CheckSubdomainAvailability checks if a subdomain is available
func (c *Client) CheckSubdomainAvailability(ctx context.Context, subdomain string) (*CheckSubdomainAvailabilityResponse, error) {
	reqURL := fmt.Sprintf("%s/api/installations/check-domain-kli?subdomain=%s", c.baseURL, url.QueryEscape(subdomain))

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var result CheckSubdomainAvailabilityResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ReserveSubdomainRequest is the request to reserve-domain-kli
type ReserveSubdomainRequest struct {
	InstallationKey string `json:"installationKey"`
	Subdomain       string `json:"subdomain"`
}

// ReserveSubdomainResponse is the response from reserve-domain-kli
type ReserveSubdomainResponse struct {
	Success   bool   `json:"success"`
	Subdomain string `json:"subdomain"`
	URL       string `json:"url"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
}

// ReserveSubdomain reserves a subdomain for an installation
func (c *Client) ReserveSubdomain(ctx context.Context, installationKey, subdomain string) (*ReserveSubdomainResponse, error) {
	reqURL := fmt.Sprintf("%s/api/installations/reserve-domain-kli", c.baseURL)

	reqBody := ReserveSubdomainRequest{
		InstallationKey: installationKey,
		Subdomain:       subdomain,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response even on error status to get error message
	var result ReserveSubdomainResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && result.Error != "" {
		return nil, fmt.Errorf("API error: %s", result.Error)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return &result, nil
}

// ACMValidationRecord represents a DNS validation record for ACM
type ACMValidationRecord struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CreateACMValidationRequest is the request to create-acm-validation
type CreateACMValidationRequest struct {
	InstallationKey   string                `json:"installationKey"`
	ValidationRecords []ACMValidationRecord `json:"validationRecords"`
}

// CreateACMValidationResponse is the response from create-acm-validation
type CreateACMValidationResponse struct {
	Success   bool     `json:"success"`
	RecordIDs []string `json:"recordIds"`
	Created   int      `json:"created"`
	Total     int      `json:"total"`
	Errors    []string `json:"errors,omitempty"`
	Message   string   `json:"message,omitempty"`
	Error     string   `json:"error,omitempty"`
}

// CreateACMValidationRecords creates DNS validation records in Cloudflare for ACM
func (c *Client) CreateACMValidationRecords(ctx context.Context, installationKey, secretKey string, records []ACMValidationRecord) (*CreateACMValidationResponse, error) {
	reqURL := fmt.Sprintf("%s/api/installations/create-acm-validation", c.baseURL)

	reqBody := CreateACMValidationRequest{
		InstallationKey:   installationKey,
		ValidationRecords: records,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secretKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result CreateACMValidationResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && result.Error != "" {
		return nil, fmt.Errorf("API error: %s", result.Error)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return &result, nil
}

// ConfigureRootDNSRequest is the request to configure-root-dns
type ConfigureRootDNSRequest struct {
	InstallationKey string `json:"installationKey"`
	Target          string `json:"target"`  // DNS name for CNAME or IP for A record
	Type            string `json:"type"`    // "cname" or "a"
	Proxied         bool   `json:"proxied"` // Enable Cloudflare proxy mode for TLS termination
}

// ConfigureRootDNSResponse is the response from configure-root-dns
type ConfigureRootDNSResponse struct {
	Success  bool   `json:"success"`
	Domain   string `json:"domain"`
	Target   string `json:"target"`
	Type     string `json:"type"`
	RecordID string `json:"recordId"`
	Proxied  bool   `json:"proxied"`
	Error    string `json:"error,omitempty"`
}

// ConfigureRootDNS registers the root DNS record for an installation
// For load balancers (ALB): creates CNAME record pointing to load balancer DNS
// For direct IPs: creates A record pointing to IP address
// When proxied=true, Cloudflare will proxy the traffic and handle TLS termination
func (c *Client) ConfigureRootDNS(ctx context.Context, installationKey, secretKey, target, recordType string, proxied bool) (*ConfigureRootDNSResponse, error) {
	reqURL := fmt.Sprintf("%s/api/installations/configure-root-dns", c.baseURL)

	reqBody := ConfigureRootDNSRequest{
		InstallationKey: installationKey,
		Target:          target,
		Type:            recordType, // "cname" for load balancers, "a" for IPs
		Proxied:         proxied,
	}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", secretKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result ConfigureRootDNSResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && result.Error != "" {
		return nil, fmt.Errorf("API error: %s", result.Error)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return &result, nil
}

// GetFullDomain returns the full domain for a subdomain
func GetFullDomain(subdomain string) string {
	return fmt.Sprintf("%s.%s", subdomain, KhostDomain)
}
