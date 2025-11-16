package domainrequest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// configureIPRequest represents the request body for /api/installations/configure-ips
type configureIPRequest struct {
	InstallationKey   string   `json:"installationKey"`
	IP                string   `json:"ip,omitempty"`
	DomainRequestName string   `json:"domainRequestName"`
	Domains           []string `json:"domains,omitempty"`
	Deleted           bool     `json:"deleted,omitempty"` // Set to true to delete all DNS records
}

// configureIPResponse represents the response from /api/installations/configure-ips
type configureIPResponse struct {
	Success                 bool   `json:"success"`
	DomainRequestName       string `json:"domainRequestName"`
	IP                      string `json:"ip"`
	SSHDomain               string `json:"sshDomain"`
	Subdomain               string `json:"subdomain"`
	SSHRecordCreated        bool   `json:"sshRecordCreated"`
	RouteRecordsCreated     int    `json:"routeRecordsCreated"`
	EdgeCertificatesCreated int    `json:"edgeCertificatesCreated"`
	DNSSuccess              bool   `json:"dnsSuccess"`
}

// generateCertificateRequest represents the request body for /api/installations/generate-certificates
type generateCertificateRequest struct {
	InstallationKey       string `json:"installationKey"`
	Scope                 string `json:"scope"`
	ScopeIdentifier       string `json:"scopeIdentifier,omitempty"`
	ParentScopeIdentifier string `json:"parentScopeIdentifier,omitempty"`
}

// generateCertificateResponse represents the response from /api/installations/generate-certificates
type generateCertificateResponse struct {
	Success   bool      `json:"success"`
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// downloadCertificatesResponse represents the response from /api/installations/download-certificates
type downloadCertificatesResponse struct {
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"privateKey"`
}

// getOriginCertificateResponse represents the response from /api/installations/get-origin-certificate
type getOriginCertificateResponse struct {
	Success       bool   `json:"success"`
	Certificate   string `json:"certificate"`
	PrivateKey    string `json:"privateKey"`
	CertificateID string `json:"certificateId"`
	ValidFrom     string `json:"validFrom"`
	ValidUntil    string `json:"validUntil"`
}

// callConsoleAPI makes HTTP requests to console.kloudlite.io API
func (r *DomainRequestReconciler) callConsoleAPI(ctx context.Context, path, method string, body interface{}, authToken string, logger *zap.Logger) ([]byte, error) {
	url := consoleAPIBaseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authToken)

	logger.Info("Calling console API",
		zap.String("method", method),
		zap.String("url", url))

	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call console API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Error("Console API returned error",
			zap.Int("statusCode", resp.StatusCode),
			zap.String("response", string(respBody)))
		return nil, fmt.Errorf("console API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
