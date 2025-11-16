package domainrequest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	domainrequestsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// computeDomainRoutesHash computes a SHA256 hash of the DomainRoutes to detect changes
func computeDomainRoutesHash(routes []domainrequestsv1.DomainRoute) (string, error) {
	// Convert routes to JSON for consistent hashing
	routesJSON, err := json.Marshal(routes)
	if err != nil {
		return "", fmt.Errorf("failed to marshal routes: %w", err)
	}

	// Compute SHA256 hash
	hash := sha256.Sum256(routesJSON)
	return hex.EncodeToString(hash[:]), nil
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
			if errors.IsNotFound(err) {
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

	// Extract domain names from DomainRoutes
	domains := make([]string, 0, len(domainRequest.Spec.DomainRoutes))
	for _, route := range domainRequest.Spec.DomainRoutes {
		domains = append(domains, route.Domain)
	}

	// Call console API to register IP and create DNS records
	reqBody := configureIPRequest{
		InstallationKey:   r.InstallationKey,
		IP:                ipAddress,
		DomainRequestName: domainRequest.Name,
		Domains:           domains,
	}

	resp, err := r.callConsoleAPI(ctx, "/api/installations/configure-ips", "POST", reqBody, r.InstallationSecret, logger)
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

	// Build DNS record IDs array (SSH record + route records)
	var dnsRecordIDs []string
	if configResp.SSHRecordCreated {
		// Note: API doesn't return individual SSH record ID, but we track it internally
		logger.Info("SSH DNS record created successfully", zap.String("sshDomain", configResp.SSHDomain))
	}

	// Compute hash of current routes to track for future reconciliation
	routesHash, err := computeDomainRoutesHash(domainRequest.Spec.DomainRoutes)
	if err != nil {
		logger.Error("Failed to compute routes hash", zap.Error(err))
		return r.updateStatus(ctx, domainRequest, "Failed", fmt.Sprintf("Failed to compute routes hash: %v", err), logger)
	}

	// Update status with registration details
	// New flow: Transition to Ready state after IP registration (certificate and HAProxy already set up)
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, domainRequest, func() error {
		// Build accurate status message based on actual HAProxy state
		haproxyStatus := "not created"
		if domainRequest.Status.HAProxyReady {
			haproxyStatus = "created and ready"
		} else if domainRequest.Status.HAProxyPodName != "" {
			haproxyStatus = "creating"
		}

		domainRequest.Status.State = "Ready"
		domainRequest.Status.Message = fmt.Sprintf("DomainRequest is ready - origin cert downloaded, HAProxy %s, DNS configured (%d routes)", haproxyStatus, configResp.RouteRecordsCreated)
		domainRequest.Status.Domain = configResp.SSHDomain
		domainRequest.Status.Subdomain = configResp.Subdomain
		domainRequest.Status.DNSRecordIDs = dnsRecordIDs
		now := metav1.Now()
		domainRequest.Status.LastIPRegistrationTime = &now
		domainRequest.Status.LastReconciledRoutesHash = routesHash
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("IP registration successful - DomainRequest is now ready",
		zap.String("sshDomain", configResp.SSHDomain),
		zap.String("subdomain", configResp.Subdomain),
		zap.Int("routeRecords", configResp.RouteRecordsCreated),
		zap.Int("edgeCertificates", configResp.EdgeCertificatesCreated))

	return reconcile.Result{Requeue: true}, nil
}

// handleCertificateGeneration generates TLS certificates
func (r *DomainRequestReconciler) handleCertificateGeneration(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Handling certificate generation")

	// Call console API to generate certificate
	reqBody := generateCertificateRequest{
		InstallationKey:       r.InstallationKey,
		Scope:                 domainRequest.Spec.CertificateScope,
		ScopeIdentifier:       domainRequest.Spec.CertificateScopeIdentifier,
		ParentScopeIdentifier: domainRequest.Spec.CertificateParentScopeIdentifier,
	}

	resp, err := r.callConsoleAPI(ctx, "/api/installations/generate-certificates", "POST", reqBody, r.InstallationSecret, logger)
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
		r.InstallationKey,
		domainRequest.Spec.CertificateScope)

	if domainRequest.Spec.CertificateScopeIdentifier != "" {
		downloadURL += "&scopeIdentifier=" + domainRequest.Spec.CertificateScopeIdentifier
	}
	if domainRequest.Spec.CertificateParentScopeIdentifier != "" {
		downloadURL += "&parentScopeIdentifier=" + domainRequest.Spec.CertificateParentScopeIdentifier
	}

	// Call console API to download certificate
	resp, err := r.callConsoleAPI(ctx, downloadURL, "GET", nil, r.InstallationSecret, logger)
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
			Namespace: domainRequest.Spec.WorkloadNamespace,
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
		// Add combined PEM file for HAProxy
		combinedPEM := downloadResp.Certificate + "\n" + downloadResp.PrivateKey
		secret.Data["tls.pem"] = []byte(combinedPEM)

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

// handleOriginCertificateDownload downloads the installation's origin certificate
func (r *DomainRequestReconciler) handleOriginCertificateDownload(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) (reconcile.Result, error) {
	// Check if secret already exists (e.g., from previous reconciliation)
	secretName := fmt.Sprintf("%s-origin-cert", domainRequest.Name)
	existingSecret := &corev1.Secret{}
	err := r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: domainRequest.Spec.WorkloadNamespace}, existingSecret)
	if err == nil {
		// Secret already exists, move to next state
		logger.Info("Origin certificate secret already exists, skipping download", zap.String("secretName", secretName))

		// Update status to move to next state
		latestDomainRequest := &domainrequestsv1.DomainRequest{}
		if err := r.Get(ctx, client.ObjectKey{Name: domainRequest.Name}, latestDomainRequest); err != nil {
			logger.Error("Failed to refetch DomainRequest", zap.Error(err))
			return reconcile.Result{}, err
		}

		latestDomainRequest.Status.State = "CertificateGenerated"
		latestDomainRequest.Status.Message = "Origin certificate already exists"
		latestDomainRequest.Status.OriginCertificateSecretName = secretName

		if err := r.Status().Update(ctx, latestDomainRequest); err != nil {
			logger.Error("Failed to update status", zap.Error(err))
			return reconcile.Result{}, err
		}

		return reconcile.Result{Requeue: true}, nil
	} else if !errors.IsNotFound(err) {
		logger.Error("Failed to check for existing secret", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Fetching or creating origin certificate",
		zap.String("domainRequestName", domainRequest.Name),
		zap.Strings("hostnames", domainRequest.Spec.OriginCertificateHostnames))

	// Prepare request body with identifier (domainRequestName) and hostnames
	// Console API will check Supabase for certificate with key (domainRequestName, installationKey):
	// - If exists -> return it
	// - If not -> create via Cloudflare API, store in Supabase with this key, return it
	reqBody := map[string]interface{}{
		"installationKey":   r.InstallationKey,
		"domainRequestName": domainRequest.Name, // Used as identifier in Supabase
	}

	// Include hostnames if specified in DomainRequest spec
	if len(domainRequest.Spec.OriginCertificateHostnames) > 0 {
		reqBody["hostnames"] = domainRequest.Spec.OriginCertificateHostnames
	} else {
		logger.Info("No custom hostnames specified, API will use defaults")
	}

	// Call create-origin-certificate endpoint
	// Idempotent - returns existing certificate for (domainRequestName, installationKey) if found in Supabase
	createPath := "/api/installations/create-origin-certificate"
	resp, createErr := r.callConsoleAPI(ctx, createPath, "POST", reqBody, r.InstallationSecret, logger)
	if createErr != nil {
		logger.Error("Failed to get or create origin certificate", zap.Error(createErr))
		domainRequest.Status.State = "Failed"
		domainRequest.Status.Message = fmt.Sprintf("Failed to get or create origin certificate: %v", createErr)
		if updateErr := r.Status().Update(ctx, domainRequest); updateErr != nil {
			logger.Error("Failed to update status", zap.Error(updateErr))
			return reconcile.Result{}, updateErr
		}
		return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
	}

	var certResp getOriginCertificateResponse
	if err := json.Unmarshal(resp, &certResp); err != nil {
		logger.Error("Failed to parse origin certificate response", zap.Error(err))
		domainRequest.Status.State = "Failed"
		domainRequest.Status.Message = "Failed to parse origin certificate response"
		if updateErr := r.Status().Update(ctx, domainRequest); updateErr != nil {
			logger.Error("Failed to update status", zap.Error(updateErr))
			return reconcile.Result{}, updateErr
		}
		return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Create a secret to store the origin certificate
	secretName = fmt.Sprintf("%s-origin-cert", domainRequest.Name)

	// Create combined PEM file for HAProxy (certificate + key)
	combinedPEM := certResp.Certificate + "\n" + certResp.PrivateKey

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: domainRequest.Spec.WorkloadNamespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": []byte(certResp.Certificate),
			"tls.key": []byte(certResp.PrivateKey),
			"tls.pem": []byte(combinedPEM), // HAProxy requires combined PEM
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(domainRequest, secret, r.Scheme); err != nil {
		logger.Error("Failed to set owner reference on Secret", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Create or update secret
	existingSecret = &corev1.Secret{}
	err = r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: domainRequest.Spec.WorkloadNamespace}, existingSecret)
	if err != nil {
		if errors.IsNotFound(err) {
			// Create new secret
			if err := r.Create(ctx, secret); err != nil {
				logger.Error("Failed to create Secret", zap.Error(err))
				return reconcile.Result{}, err
			}
			logger.Info("Created origin certificate Secret", zap.String("secretName", secretName))
		} else {
			logger.Error("Failed to check for existing Secret", zap.Error(err))
			return reconcile.Result{}, err
		}
	} else {
		// Update existing secret
		existingSecret.Data = secret.Data
		if err := r.Update(ctx, existingSecret); err != nil {
			logger.Error("Failed to update Secret", zap.Error(err))
			return reconcile.Result{}, err
		}
		logger.Info("Updated origin certificate Secret", zap.String("secretName", secretName))
	}

	// Refetch the latest DomainRequest to avoid resourceVersion conflicts
	latestDomainRequest := &domainrequestsv1.DomainRequest{}
	if err := r.Get(ctx, client.ObjectKey{Name: domainRequest.Name}, latestDomainRequest); err != nil {
		logger.Error("Failed to refetch DomainRequest before status update", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Update status
	latestDomainRequest.Status.State = "CertificateGenerated"
	latestDomainRequest.Status.Message = "Origin certificate downloaded and stored"
	latestDomainRequest.Status.OriginCertificateSecretName = secretName

	if err := r.Status().Update(ctx, latestDomainRequest); err != nil {
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Origin certificate download successful", zap.String("secretName", secretName))
	return reconcile.Result{Requeue: true}, nil
}

// handleHAProxyCreation creates HAProxy pod and ConfigMap for traffic routing
func (r *DomainRequestReconciler) handleHAProxyCreation(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Handling HAProxy creation")

	// Create HAProxy ConfigMap
	if err := r.createHAProxyConfigMap(ctx, domainRequest, logger); err != nil {
		logger.Error("Failed to create HAProxy ConfigMap", zap.Error(err))
		return r.updateStatus(ctx, domainRequest, "Failed", fmt.Sprintf("Failed to create HAProxy config: %v", err), logger)
	}

	// Create HAProxy Pod
	if err := r.createHAProxyPod(ctx, domainRequest, logger); err != nil {
		logger.Error("Failed to create HAProxy pod", zap.Error(err))
		return r.updateStatus(ctx, domainRequest, "Failed", fmt.Sprintf("Failed to create HAProxy pod: %v", err), logger)
	}

	podName := fmt.Sprintf("%s-haproxy", domainRequest.Name)

	// Update status to HAProxyCreating
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, domainRequest, func() error {
		domainRequest.Status.State = "HAProxyCreating"
		domainRequest.Status.Message = "HAProxy pod and config created, waiting for ready status"
		domainRequest.Status.HAProxyPodName = podName
		domainRequest.Status.HAProxyReady = false
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("HAProxy resources created successfully",
		zap.String("podName", podName))

	// Requeue immediately to check pod status
	return reconcile.Result{Requeue: true}, nil
}

// handleHAProxyStatusCheck checks if HAProxy pod is ready
func (r *DomainRequestReconciler) handleHAProxyStatusCheck(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Checking HAProxy pod status")

	if domainRequest.Status.HAProxyPodName == "" {
		logger.Error("HAProxy pod name not set in status")
		return r.updateStatus(ctx, domainRequest, "Failed", "HAProxy pod name not found in status", logger)
	}

	// Check if HAProxy pod is ready
	ready, err := r.checkHAProxyReady(ctx, domainRequest.Spec.WorkloadNamespace, domainRequest.Status.HAProxyPodName, logger)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Warn("HAProxy pod not found, recreating it")
			// Pod was deleted or never created, recreate it
			if err := r.createHAProxyPod(ctx, domainRequest, logger); err != nil {
				logger.Error("Failed to recreate HAProxy pod", zap.Error(err))
				return reconcile.Result{}, err
			}
			return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
		}
		logger.Error("Failed to check HAProxy pod status", zap.Error(err))
		return reconcile.Result{}, err
	}

	if !ready {
		logger.Info("HAProxy pod not ready yet, will retry")
		return reconcile.Result{RequeueAfter: 10 * time.Second}, nil
	}

	// Update status to HAProxyReady
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, domainRequest, func() error {
		domainRequest.Status.State = "HAProxyReady"
		domainRequest.Status.Message = "HAProxy is running and serving traffic"
		domainRequest.Status.HAProxyReady = true
		return nil
	}, logger); err != nil {
		logger.Error("Failed to update status", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("HAProxy pod is ready and serving traffic",
		zap.String("podName", domainRequest.Status.HAProxyPodName))

	// Requeue immediately to proceed to IP registration
	return reconcile.Result{Requeue: true}, nil
}

// handleDeletion handles cleanup when DomainRequest is deleted
func (r *DomainRequestReconciler) handleDeletion(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Handling DomainRequest deletion")

	if controllerutil.ContainsFinalizer(domainRequest, domainRequestFinalizer) {
		// Delete HAProxy resources if they exist
		if domainRequest.Status.HAProxyPodName != "" {
			logger.Info("Deleting HAProxy resources")
			if err := r.deleteHAProxyResources(ctx, domainRequest, logger); err != nil {
				logger.Error("Failed to delete HAProxy resources", zap.Error(err))
				// Continue with deletion even if cleanup fails
			}
		}

		// Delete all DomainRequest resources from Cloudflare and Supabase
		// This includes:
		// 1. Origin certificates (Cloudflare Origin CA)
		// 2. Edge certificates (Cloudflare Edge TLS)
		// 3. DNS records (A/AAAA records)
		// All stored in Supabase with key (domainRequestName, installationKey)
		logger.Info("Deleting DomainRequest resources from Cloudflare and Supabase",
			zap.String("domainRequestName", domainRequest.Name))

		deleteReqBody := map[string]interface{}{
			"installationKey":   r.InstallationKey,
			"domainRequestName": domainRequest.Name,
		}

		// Call delete-domain-request endpoint to remove all associated resources
		_, err := r.callConsoleAPI(ctx, "/api/installations/delete-domain-request", "POST", deleteReqBody, r.InstallationSecret, logger)
		if err != nil {
			logger.Error("Failed to delete DomainRequest resources via console API", zap.Error(err))
			// Continue with finalizer removal even if API call fails
		} else {
			logger.Info("Successfully deleted all DomainRequest resources from Cloudflare and Supabase")
		}

		// TLS Secret cleanup is automatic via owner references

		controllerutil.RemoveFinalizer(domainRequest, domainRequestFinalizer)
		if err := r.Update(ctx, domainRequest); err != nil {
			logger.Error("Failed to remove finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		logger.Info("Removed finalizer from DomainRequest")
	}

	return reconcile.Result{}, nil
}
