package k8s

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type VerifyInstallationRequest struct {
	InstallationKey string `json:"installationKey"`
}

type VerifyInstallationResponse struct {
	Success   bool   `json:"success"`
	SecretKey string `json:"secretKey"`
	Subdomain string `json:"subdomain,omitempty"`
	Error     string `json:"error,omitempty"`
}

func VerifyInstallation(ctx context.Context, installationKey string) (string, error) {
	// TODO: Make this configurable via flag or environment variable
	registrationAPIURL := "https://console.kloudlite.io/api/installations/verify-key"

	// Create request payload
	reqPayload := VerifyInstallationRequest{
		InstallationKey: installationKey,
	}
	reqBody, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", registrationAPIURL, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	var verifyResp VerifyInstallationResponse
	if err := json.Unmarshal(respBody, &verifyResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API error
	if verifyResp.Error != "" {
		return "", fmt.Errorf("API error: %s", verifyResp.Error)
	}

	// Validate secret key
	if verifyResp.SecretKey == "" {
		return "", fmt.Errorf("no secret key returned from API")
	}

	return verifyResp.SecretKey, nil
}
