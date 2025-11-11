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

type SubdomainPoller struct {
	config        *config.InstallationConfig
	k8sClient     client.Client
	logger        *zap.Logger
	caInitializer *CAInitializer
	httpClient    *http.Client
	stopCh        chan struct{}
	stopped       bool
	stopOnce      sync.Once
	readyCh       chan struct{}
	readyOnce     sync.Once
}

// NewSubdomainPoller creates a new subdomain poller
func NewSubdomainPoller(cfg *config.InstallationConfig, k8sClient client.Client, caInitializer *CAInitializer, logger *zap.Logger) *SubdomainPoller {
	return &SubdomainPoller{
		config:        cfg,
		k8sClient:     k8sClient,
		logger:        logger.Named("subdomain-poller"),
		caInitializer: caInitializer,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		stopCh:  make(chan struct{}),
		readyCh: make(chan struct{}),
	}
}

func (sp *SubdomainPoller) Start(ctx context.Context) {
	if sp.config.InstallationKey == "" {
		sp.logger.Info("Installation key not configured, skipping subdomain polling")
		return
	}

	sp.logger.Info("Starting subdomain poller",
		zap.String("console_url", sp.config.ConsoleURL),
		zap.Int("interval_seconds", sp.config.PollingIntervalSeconds))

	select {
	case <-time.After(10 * time.Second):
	case <-ctx.Done():
		return
	}

	if err := sp.ensureDomainRequestOnStartup(ctx); err != nil {
		sp.logger.Error("Failed to ensure DomainRequest on startup", zap.Error(err))
	}

	ticker := time.NewTicker(time.Duration(sp.config.PollingIntervalSeconds) * time.Second)
	defer ticker.Stop()

	if err := sp.poll(ctx); err != nil {
		sp.logger.Error("Initial poll failed", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-sp.stopCh:
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

func (sp *SubdomainPoller) Stop() {
	sp.stopOnce.Do(func() {
		sp.stopped = true
		close(sp.stopCh)
	})
}

func (sp *SubdomainPoller) markReady() {
	sp.readyOnce.Do(func() {
		close(sp.readyCh)
		sp.logger.Info("Subdomain obtained, controllers can start")
	})
}

func (sp *SubdomainPoller) WaitUntilReady(ctx context.Context) error {
	if sp.config.InstallationKey == "" {
		sp.markReady()
		return nil
	}

	select {
	case <-sp.readyCh:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("context cancelled while waiting for subdomain: %w", ctx.Err())
	}
}

func (sp *SubdomainPoller) handleSubdomainDetected(ctx context.Context, subdomain string) error {
	if err := sp.createOrUpdateDomainRequest(ctx, subdomain); err != nil {
		return fmt.Errorf("failed to create/update domain request: %w", err)
	}

	os.Setenv("HOSTED_SUBDOMAIN", subdomain+".khost.dev")
	sp.markReady()
	return nil
}

func (sp *SubdomainPoller) ensureDomainRequestOnStartup(ctx context.Context) error {
	verifyResp, err := sp.verifyInstallationKey(ctx)
	if err != nil {
		return nil
	}

	if verifyResp.Subdomain == "" || verifyResp.Subdomain == "0.0.0.0" {
		return nil
	}

	sp.logger.Info("Subdomain detected", zap.String("subdomain", verifyResp.Subdomain))
	return sp.handleSubdomainDetected(ctx, verifyResp.Subdomain)
}

func (sp *SubdomainPoller) poll(ctx context.Context) error {
	verifyResp, err := sp.verifyInstallationKey(ctx)
	if err != nil {
		return fmt.Errorf("failed to verify installation key: %w", err)
	}

	if verifyResp.Subdomain == "" || verifyResp.Subdomain == "0.0.0.0" {
		return nil
	}

	sp.logger.Info("Subdomain detected", zap.String("subdomain", verifyResp.Subdomain))

	if err := sp.handleSubdomainDetected(ctx, verifyResp.Subdomain); err != nil {
		return err
	}

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
	fqdn := fmt.Sprintf("%s.khost.dev", subdomain)

	if err := sp.caInitializer.ensureCA(ctx, fqdn); err != nil {
		return err
	}

	if err := sp.caInitializer.ensureWildcardCertificate(ctx); err != nil {
		return err
	}

	existingDR := &domainrequestv1.DomainRequest{}
	err := sp.k8sClient.Get(ctx, client.ObjectKey{Name: domainRequestName}, existingDR)
	if err == nil {
		existingDR.Spec.IPAddress = sp.config.PublicIP
		if err := sp.k8sClient.Update(ctx, existingDR); err != nil {
			return fmt.Errorf("failed to update domain request: %w", err)
		}
		return nil
	}

	if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check if domain request exists: %w", err)
	}

	domainRequest := &domainrequestv1.DomainRequest{
		ObjectMeta: metav1.ObjectMeta{
			Name: domainRequestName,
		},
		Spec: domainrequestv1.DomainRequestSpec{
			NodeName:          os.Getenv("NODE_NAME"),
			WorkloadNamespace: "kloudlite-ingress",
			IPAddress:         sp.config.PublicIP,
			CertificateScope:  "installation",
			// Covers both {subdomain}.khost.dev and *.{subdomain}.khost.dev
			OriginCertificateHostnames: []string{
				fqdn,
				"*." + fqdn,
			},
			DomainRoutes: []domainrequestv1.DomainRoute{
				{
					Domain:           fqdn,
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

	return nil
}
