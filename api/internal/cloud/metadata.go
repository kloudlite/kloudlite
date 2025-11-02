package cloud

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// MetadataProvider provides methods to fetch cloud provider metadata
type MetadataProvider interface {
	GetPublicIP(ctx context.Context) (string, error)
}

// AWSMetadataProvider fetches metadata from AWS EC2
type AWSMetadataProvider struct {
	client *http.Client
}

// NewAWSMetadataProvider creates a new AWS metadata provider
func NewAWSMetadataProvider() *AWSMetadataProvider {
	return &AWSMetadataProvider{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetPublicIP fetches the public IP from AWS EC2 metadata service using IMDSv2
func (p *AWSMetadataProvider) GetPublicIP(ctx context.Context) (string, error) {
	// Step 1: Get IMDSv2 token
	tokenURL := "http://169.254.169.254/latest/api/token"
	tokenReq, err := http.NewRequestWithContext(ctx, "PUT", tokenURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	tokenReq.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "21600")

	tokenResp, err := p.client.Do(tokenReq)
	if err != nil {
		return "", fmt.Errorf("failed to fetch IMDSv2 token: %w", err)
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token request returned status %d", tokenResp.StatusCode)
	}

	tokenBody, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}
	token := string(tokenBody)

	// Step 2: Use token to fetch public IP
	ipURL := "http://169.254.169.254/latest/meta-data/public-ipv4"
	req, err := http.NewRequestWithContext(ctx, "GET", ipURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("X-aws-ec2-metadata-token", token)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch public IP from EC2 metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("metadata service returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	publicIP := string(body)
	if publicIP == "" {
		return "", fmt.Errorf("metadata service returned empty IP")
	}

	return publicIP, nil
}
