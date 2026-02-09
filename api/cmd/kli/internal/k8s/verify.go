package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type VerifyInstallationRequest struct {
	InstallationKey string `json:"installationKey"`
	Provider        string `json:"provider,omitempty"`
	Region          string `json:"region,omitempty"`
}

type VerifyInstallationResponse struct {
	Success   bool   `json:"success"`
	SecretKey string `json:"secretKey"`
	Subdomain string `json:"subdomain,omitempty"`
	Error     string `json:"error,omitempty"`
}

// VerifyInstallationResult contains the result of verification
type VerifyInstallationResult struct {
	SecretKey string
	Subdomain string
}

// VerifyInstallationOptions contains optional parameters for verification
type VerifyInstallationOptions struct {
	Provider string // aws, gcp, azure
	Region   string // cloud region/location
}

func VerifyInstallation(ctx context.Context, installationKey string, opts *VerifyInstallationOptions) (*VerifyInstallationResult, error) {
	baseURL := os.Getenv("CONSOLE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://console.kloudlite.io"
	}
	registrationAPIURL := baseURL + "/api/installations/verify-key"

	// Create request payload
	reqPayload := VerifyInstallationRequest{
		InstallationKey: installationKey,
	}
	if opts != nil {
		reqPayload.Provider = opts.Provider
		reqPayload.Region = opts.Region
	}
	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", registrationAPIURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var verifyResp VerifyInstallationResponse
	if err := json.Unmarshal(respBody, &verifyResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error
	if verifyResp.Error != "" {
		return nil, fmt.Errorf("API error: %s", verifyResp.Error)
	}

	// Validate secret key
	if verifyResp.SecretKey == "" {
		return nil, fmt.Errorf("no secret key returned from API")
	}

	return &VerifyInstallationResult{
		SecretKey: verifyResp.SecretKey,
		Subdomain: verifyResp.Subdomain,
	}, nil
}
