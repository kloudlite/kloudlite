package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/kloudlite/kloudlite/api/internal/config"
	domainrequestv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	domainRequestName = "installation-domain"
)

// VerifyKeyResponse represents the response from the console verify-key API
type VerifyKeyResponse struct {
	Success         bool   `json:"success"`
	SecretKey       string `json:"secretKey"`
	Subdomain       string `json:"subdomain"`
	DeploymentReady bool   `json:"deploymentReady"`
}

// SubdomainPoller periodically polls for subdomain configuration
type SubdomainPoller struct {
	config     *config.InstallationConfig
	k8sClient  client.Client
	logger     *zap.Logger
	httpClient *http.Client
	stopCh     chan struct{}
	stopped    bool
	stopOnce   sync.Once
}

// NewSubdomainPoller creates a new subdomain poller
func NewSubdomainPoller(cfg *config.InstallationConfig, k8sClient client.Client, logger *zap.Logger) *SubdomainPoller {
	return &SubdomainPoller{
		config:    cfg,
		k8sClient: k8sClient,
		logger:    logger.Named("subdomain-poller"),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopCh: make(chan struct{}),
	}
}

// Start begins the polling loop
func (sp *SubdomainPoller) Start(ctx context.Context) {
	// Skip if installation key is not configured
	if sp.config.InstallationKey == "" {
		sp.logger.Info("Installation key not configured, skipping subdomain polling")
		return
	}

	sp.logger.Info("Starting subdomain poller",
		zap.String("console_url", sp.config.ConsoleURL),
		zap.Int("interval_seconds", sp.config.PollingIntervalSeconds))

	// Wait for service endpoints to be ready before making webhook calls
	// This prevents "no endpoints available" errors during startup
	sp.logger.Debug("Waiting 10 seconds for service endpoints to be ready")
	select {
	case <-time.After(10 * time.Second):
		// Continue after delay
	case <-ctx.Done():
		sp.logger.Info("Subdomain poller stopped during startup delay")
		return
	}

	// Ensure DomainRequest is created/updated idempotently on startup
	if err := sp.ensureDomainRequestOnStartup(ctx); err != nil {
		sp.logger.Error("Failed to ensure DomainRequest on startup", zap.Error(err))
	}

	ticker := time.NewTicker(time.Duration(sp.config.PollingIntervalSeconds) * time.Second)
	defer ticker.Stop()

	// Try polling immediately on startup
	if err := sp.poll(ctx); err != nil {
		sp.logger.Error("Initial poll failed", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			sp.logger.Info("Subdomain poller stopped due to context cancellation")
			return
		case <-sp.stopCh:
			sp.logger.Info("Subdomain poller stopped")
			return
		case <-ticker.C:
			if sp.stopped {
				return
			}
			if err := sp.poll(ctx); err != nil {
				sp.logger.Error("Poll failed", zap.Error(err))
			}
		}
	}
}

// Stop stops the poller (safe to call multiple times)
func (sp *SubdomainPoller) Stop() {
	sp.stopOnce.Do(func() {
		sp.stopped = true
		close(sp.stopCh)
	})
}

// ensureDomainRequestOnStartup ensures the DomainRequest is created/updated on API server startup
func (sp *SubdomainPoller) ensureDomainRequestOnStartup(ctx context.Context) error {
	sp.logger.Info("Ensuring DomainRequest on startup")

	// Call verify-key API to get subdomain
	verifyResp, err := sp.verifyInstallationKey(ctx)
	if err != nil {
		sp.logger.Debug("Failed to verify installation key on startup", zap.Error(err))
		// Don't fail startup if verify-key fails, just log and continue
		return nil
	}

	// Check if subdomain is set and valid
	if verifyResp.Subdomain == "" || verifyResp.Subdomain == "0.0.0.0" {
		sp.logger.Debug("Subdomain not yet configured on startup")
		return nil
	}

	sp.logger.Info("Subdomain detected on startup, ensuring DomainRequest", zap.String("subdomain", verifyResp.Subdomain))

	// Create or update DomainRequest
	if err := sp.createOrUpdateDomainRequest(ctx, verifyResp.Subdomain); err != nil {
		return fmt.Errorf("failed to create/update domain request on startup: %w", err)
	}

	os.Setenv("HOSTED_SUBDOMAIN", verifyResp.Subdomain)

	sp.logger.Info("DomainRequest ensured successfully on startup")
	return nil
}

// poll performs a single poll attempt
func (sp *SubdomainPoller) poll(ctx context.Context) error {
	// Call verify-key API
	verifyResp, err := sp.verifyInstallationKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to verify installation key: %w", err)
	}

	// Check if subdomain is set and valid
	if verifyResp.Subdomain == "" || verifyResp.Subdomain == "0.0.0.0" {
		sp.logger.Debug("Subdomain not yet configured")
		return nil
	}

	sp.logger.Info("Subdomain detected", zap.String("subdomain", verifyResp.Subdomain))

	// Create or update DomainRequest
	if err := sp.createOrUpdateDomainRequest(ctx, verifyResp.Subdomain); err != nil {
		return fmt.Errorf("failed to create/update domain request: %w", err)
	}

	sp.logger.Info("DomainRequest created/updated successfully, stopping poller")
	sp.Stop()

	return nil
}

// verifyInstallationKey calls the console verify-key API
func (sp *SubdomainPoller) verifyInstallationKey(ctx context.Context) (*VerifyKeyResponse, error) {
	verifyURL := fmt.Sprintf("%s/api/installations/verify-key", sp.config.ConsoleURL)

	reqBody := map[string]string{
		"installationKey": sp.config.InstallationKey,
	}

	reqJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, verifyURL, bytes.NewBuffer(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := sp.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var verifyResp VerifyKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &verifyResp, nil
}

// createOrUpdateDomainRequest creates or updates the DomainRequest resource
func (sp *SubdomainPoller) createOrUpdateDomainRequest(ctx context.Context, subdomain string) error {
	// Check if DomainRequest already exists
	existingDR := &domainrequestv1.DomainRequest{}
	err := sp.k8sClient.Get(ctx, client.ObjectKey{Name: domainRequestName, Namespace: "kloudlite"}, existingDR)

	if err == nil {
		// DomainRequest exists, update it
		sp.logger.Info("Updating existing DomainRequest", zap.String("name", domainRequestName))

		existingDR.Spec.IPAddress = sp.config.PublicIP

		if err := sp.k8sClient.Update(ctx, existingDR); err != nil {
			return fmt.Errorf("failed to update domain request: %w", err)
		}

		sp.logger.Info("DomainRequest updated successfully")
		return nil
	}

	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check if domain request exists: %w", err)
	}

	// DomainRequest doesn't exist, create it
	sp.logger.Info("Creating new DomainRequest", zap.String("name", domainRequestName))

	domainRequest := &domainrequestv1.DomainRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name:      domainRequestName,
			Namespace: "kloudlite",
		},
		Spec: domainrequestv1.DomainRequestSpec{
			NodeName:         os.Getenv("NODE_NAME"),
			IPAddress:        sp.config.PublicIP,
			CertificateScope: "installation",
			// Specify origin certificate hostnames to include the base domain
			// This ensures the certificate covers both {subdomain}.khost.dev and *.{subdomain}.khost.dev
			OriginCertificateHostnames: []string{
				fmt.Sprintf("%s.khost.dev", subdomain),
				fmt.Sprintf("*.%s.khost.dev", subdomain),
			},
			DomainRoutes: []domainrequestv1.DomainRoute{
				{
					Domain:           fmt.Sprintf("%s.khost.dev", subdomain),
					ServiceName:      "frontend",
					ServiceNamespace: "kloudlite",
					ServicePort:      80,
				},
			},
		},
	}

	if err := sp.k8sClient.Create(ctx, domainRequest); err != nil {
		return fmt.Errorf("failed to create domain request: %w", err)
	}

	sp.logger.Info("DomainRequest created successfully")
	return nil
}
