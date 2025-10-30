package domainrequest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	domainrequestsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	domainRequestFinalizer = "domains.kloudlite.io/finalizer"
	consoleAPIBaseURL      = "https://console.kloudlite.io"
)

// DomainRequestReconciler reconciles DomainRequest objects
type DomainRequestReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
}

// configureIPRequest represents the request body for /api/installations/configure-ips
type configureIPRequest struct {
	InstallationKey string `json:"installationKey"`
	Type            string `json:"type"`
	IP              string `json:"ip"`
	WorkMachineName string `json:"workMachineName,omitempty"`
}

// configureIPResponse represents the response from /api/installations/configure-ips
type configureIPResponse struct {
	Success      bool     `json:"success"`
	Domain       string   `json:"domain"`
	Subdomain    string   `json:"subdomain"`
	DNSRecordIDs []string `json:"dnsRecordIds"`
}

// generateCertificateRequest represents the request body for /api/installations/generate-certificates
type generateCertificateRequest struct {
	InstallationKey            string `json:"installationKey"`
	Scope                      string `json:"scope"`
	ScopeIdentifier            string `json:"scopeIdentifier,omitempty"`
	ParentScopeIdentifier      string `json:"parentScopeIdentifier,omitempty"`
}

// generateCertificateResponse represents the response from /api/installations/generate-certificates
type generateCertificateResponse struct {
	Success     bool      `json:"success"`
	ID          string    `json:"id"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// downloadCertificatesResponse represents the response from /api/installations/download-certificates
type downloadCertificatesResponse struct {
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"privateKey"`
}

// Reconcile handles DomainRequest reconciliation
func (r *DomainRequestReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("name", req.Name),
		zap.String("namespace", req.Namespace),
	)

	logger.Info("Reconciling DomainRequest")

	// Fetch the DomainRequest instance
	domainRequest := &domainrequestsv1.DomainRequest{}
	if err := r.Get(ctx, req.NamespacedName, domainRequest); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("DomainRequest resource not found, ignoring")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get DomainRequest", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Handle deletion
	if !domainRequest.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, domainRequest, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(domainRequest, domainRequestFinalizer) {
		controllerutil.AddFinalizer(domainRequest, domainRequestFinalizer)
		if err := r.Update(ctx, domainRequest); err != nil {
			logger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		logger.Info("Added finalizer to DomainRequest")
	}

	// Reconcile based on current state
	switch domainRequest.Status.State {
	case "", "Pending":
		return r.handleIPRegistration(ctx, domainRequest, logger)
	case "IPRegistered":
		return r.handleCertificateGeneration(ctx, domainRequest, logger)
	case "CertificateGenerated":
		return r.handleCertificateDownload(ctx, domainRequest, logger)
	case "Ready":
		// Check if certificate needs renewal (e.g., within 30 days of expiry)
		if domainRequest.Status.CertificateExpiresAt != nil {
			timeUntilExpiry := time.Until(domainRequest.Status.CertificateExpiresAt.Time)
			if timeUntilExpiry < 30*24*time.Hour {
				logger.Info("Certificate expiring soon, regenerating",
					zap.Duration("timeUntilExpiry", timeUntilExpiry))
				return r.handleCertificateGeneration(ctx, domainRequest, logger)
			}
		}
		logger.Info("DomainRequest is ready, no action needed")
		return reconcile.Result{RequeueAfter: 24 * time.Hour}, nil
	case "Failed":
		// Retry after some time
		logger.Info("DomainRequest failed, retrying after 5 minutes")
		return reconcile.Result{RequeueAfter: 5 * time.Minute}, nil
	}

	return reconcile.Result{}, nil
}

// handleIPRegistration registers the IP address with console.kloudlite.io
func (r *DomainRequestReconciler) handleIPRegistration(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Handling IP registration")

	// Get IP address from spec or LoadBalancer service
	ipAddress := domainRequest.Spec.IPAddress
	if ipAddress == "" && domainRequest.Spec.LoadBalancerServiceName != "" {
		logger.Info("Fetching IP from LoadBalancer service",
			zap.String("serviceName", domainRequest.Spec.LoadBalancerServiceName),
			zap.String("serviceNamespace", domainRequest.Spec.LoadBalancerServiceNamespace))

		svc := &corev1.Service{}
		serviceKey := client.ObjectKey{
			Name:      domainRequest.Spec.LoadBalancerServiceName,
			Namespace: domainRequest.Spec.LoadBalancerServiceNamespace,
		}
		if err := r.Get(ctx, serviceKey, svc); err != nil {
			if apierrors.IsNotFound(err) {
				logger.Error("LoadBalancer service not found", zap.Error(err))
				return r.updateStatus(ctx, domainRequest, "Failed", "LoadBalancer service not found", logger)
			}
			logger.Error("Failed to get LoadBalancer service", zap.Error(err))
			return reconcile.Result{}, err
		}

		// Get LoadBalancer IP
		if len(svc.Status.LoadBalancer.Ingress) == 0 {
			logger.Info("LoadBalancer IP not yet assigned, requeuing")
			return r.updateStatus(ctx, domainRequest, "Pending", "Waiting for LoadBalancer IP assignment", logger)
		}

		ipAddress = svc.Status.LoadBalancer.Ingress[0].IP
		if ipAddress == "" {
			ipAddress = svc.Status.LoadBalancer.Ingress[0].Hostname
		}

		if ipAddress == "" {
			logger.Info("LoadBalancer IP not available yet, requeuing")
			return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
		}

		logger.Info("Got IP from LoadBalancer", zap.String("ip", ipAddress))
	}

	if ipAddress == "" {
		logger.Error("No IP address available")
		return r.updateStatus(ctx, domainRequest, "Failed", "No IP address provided or available from LoadBalancer", logger)
	}

	// Call console API to register IP
	reqBody := configureIPRequest{
		InstallationKey: domainRequest.Spec.InstallationKey,
		Type:            domainRequest.Spec.Type,
		IP:              ipAddress,
		WorkMachineName: domainRequest.Spec.WorkMachineName,
	}

	resp, err := r.callConsoleAPI(ctx, "/api/installations/configure-ips", "POST", reqBody, domainRequest.Spec.InstallationSecret, logger)
	if err != nil {
		logger.Error("Failed to register IP", zap.Error(err))
		return r.updateStatus(ctx, domainRequest, "Failed", fmt.Sprintf("Failed to register IP: %v", err), logger)
	}

	var configResp configureIPResponse
	if err := json.Unmarshal(resp, &configResp); err != nil {
		logger.Error("Failed to parse configure-ips response", zap.Error(err))
		return r.updateStatus(ctx, domainRequest, "Failed", fmt.Sprintf("Failed to parse response: %v", err), logger)
	}

	if !configResp.Success {
		logger.Error("IP registration failed")
		return r.updateStatus(ctx, domainRequest, "Failed", "IP registration API returned success=false", logger)
	}

	// Update status with registration details
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, domainRequest, func() error {
		domainRequest.Status.State = "IPRegistered"
		domainRequest.Status.Message = "IP address registered successfully"
		domainRequest.Status.Domain = configResp.Domain
		domainRequest.Status.Subdomain = configResp.Subdomain
		domainRequest.Status.DNSRecordIDs = configResp.DNSRecordIDs
		now := metav1.Now()
		domainRequest.Status.LastIPRegistrationTime = &now
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("IP registration successful",
		zap.String("domain", configResp.Domain),
		zap.String("subdomain", configResp.Subdomain))

	return reconcile.Result{Requeue: true}, nil
}

// handleCertificateGeneration generates TLS certificates
func (r *DomainRequestReconciler) handleCertificateGeneration(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Handling certificate generation")

	// Call console API to generate certificate
	reqBody := generateCertificateRequest{
		InstallationKey:       domainRequest.Spec.InstallationKey,
		Scope:                 domainRequest.Spec.CertificateScope,
		ScopeIdentifier:       domainRequest.Spec.CertificateScopeIdentifier,
		ParentScopeIdentifier: domainRequest.Spec.CertificateParentScopeIdentifier,
	}

	resp, err := r.callConsoleAPI(ctx, "/api/installations/generate-certificates", "POST", reqBody, domainRequest.Spec.InstallationSecret, logger)
	if err != nil {
		logger.Error("Failed to generate certificate", zap.Error(err))
		return r.updateStatus(ctx, domainRequest, "Failed", fmt.Sprintf("Failed to generate certificate: %v", err), logger)
	}

	var certResp generateCertificateResponse
	if err := json.Unmarshal(resp, &certResp); err != nil {
		logger.Error("Failed to parse generate-certificates response", zap.Error(err))
		return r.updateStatus(ctx, domainRequest, "Failed", fmt.Sprintf("Failed to parse response: %v", err), logger)
	}

	if !certResp.Success {
		logger.Error("Certificate generation failed")
		return r.updateStatus(ctx, domainRequest, "Failed", "Certificate generation API returned success=false", logger)
	}

	// Update status with certificate details
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, domainRequest, func() error {
		domainRequest.Status.State = "CertificateGenerated"
		domainRequest.Status.Message = "Certificate generated successfully"
		domainRequest.Status.CertificateID = certResp.ID
		expiresAt := metav1.NewTime(certResp.ExpiresAt)
		domainRequest.Status.CertificateExpiresAt = &expiresAt
		now := metav1.Now()
		domainRequest.Status.LastCertificateGenerationTime = &now
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Certificate generation successful",
		zap.String("certificateID", certResp.ID),
		zap.Time("expiresAt", certResp.ExpiresAt))

	return reconcile.Result{Requeue: true}, nil
}

// handleCertificateDownload downloads and stores the certificate in a Kubernetes Secret
func (r *DomainRequestReconciler) handleCertificateDownload(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Handling certificate download")

	// Build download URL with query parameters
	downloadURL := fmt.Sprintf("/api/installations/download-certificates?installationKey=%s&format=json&scope=%s",
		domainRequest.Spec.InstallationKey,
		domainRequest.Spec.CertificateScope)

	if domainRequest.Spec.CertificateScopeIdentifier != "" {
		downloadURL += "&scopeIdentifier=" + domainRequest.Spec.CertificateScopeIdentifier
	}
	if domainRequest.Spec.CertificateParentScopeIdentifier != "" {
		downloadURL += "&parentScopeIdentifier=" + domainRequest.Spec.CertificateParentScopeIdentifier
	}

	// Call console API to download certificate
	resp, err := r.callConsoleAPI(ctx, downloadURL, "GET", nil, domainRequest.Spec.InstallationSecret, logger)
	if err != nil {
		logger.Error("Failed to download certificate", zap.Error(err))
		return r.updateStatus(ctx, domainRequest, "Failed", fmt.Sprintf("Failed to download certificate: %v", err), logger)
	}

	var downloadResp downloadCertificatesResponse
	if err := json.Unmarshal(resp, &downloadResp); err != nil {
		logger.Error("Failed to parse download-certificates response", zap.Error(err))
		return r.updateStatus(ctx, domainRequest, "Failed", fmt.Sprintf("Failed to parse response: %v", err), logger)
	}

	// Create or update Kubernetes Secret with certificate
	secretName := fmt.Sprintf("%s-tls", domainRequest.Name)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: domainRequest.Namespace,
		},
	}

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		if secret.Type == "" {
			secret.Type = corev1.SecretTypeTLS
		}
		if secret.Data == nil {
			secret.Data = make(map[string][]byte)
		}
		secret.Data["tls.crt"] = []byte(downloadResp.Certificate)
		secret.Data["tls.key"] = []byte(downloadResp.PrivateKey)

		// Set owner reference
		if err := controllerutil.SetControllerReference(domainRequest, secret, r.Scheme); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		logger.Error("Failed to create/update certificate secret", zap.Error(err))
		return r.updateStatus(ctx, domainRequest, "Failed", fmt.Sprintf("Failed to store certificate: %v", err), logger)
	}

	// Update status to Ready
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, domainRequest, func() error {
		domainRequest.Status.State = "Ready"
		domainRequest.Status.Message = "Domain and certificate ready"
		domainRequest.Status.CertificateSecretName = secretName
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Certificate download and storage successful",
		zap.String("secretName", secretName))

	// Requeue after 24 hours to check for certificate renewal
	return reconcile.Result{RequeueAfter: 24 * time.Hour}, nil
}

// updateStatus is a helper to update DomainRequest status
func (r *DomainRequestReconciler) updateStatus(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, state, message string, logger *zap.Logger) (reconcile.Result, error) {
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, domainRequest, func() error {
		domainRequest.Status.State = state
		domainRequest.Status.Message = message
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Requeue failed requests after 5 minutes
	if state == "Failed" {
		return reconcile.Result{RequeueAfter: 5 * time.Minute}, nil
	}

	return reconcile.Result{Requeue: true}, nil
}

// handleDeletion handles cleanup when DomainRequest is deleted
func (r *DomainRequestReconciler) handleDeletion(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Handling DomainRequest deletion")

	if controllerutil.ContainsFinalizer(domainRequest, domainRequestFinalizer) {
		// Perform cleanup here if needed (e.g., delete DNS records via API)
		// For now, we just remove the finalizer

		controllerutil.RemoveFinalizer(domainRequest, domainRequestFinalizer)
		if err := r.Update(ctx, domainRequest); err != nil {
			logger.Error("Failed to remove finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		logger.Info("Removed finalizer from DomainRequest")
	}

	return reconcile.Result{}, nil
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

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

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

// SetupWithManager sets up the controller with the Manager
func (r *DomainRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&domainrequestsv1.DomainRequest{}).
		Owns(&corev1.Secret{}). // Watch Secrets owned by DomainRequest
		Complete(r)
}
