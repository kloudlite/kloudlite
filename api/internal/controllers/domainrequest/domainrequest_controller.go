package domainrequest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kloudlite/kloudlite/api/internal/cloud"
	domainrequestsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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
	Scheme             *runtime.Scheme
	Logger             *zap.Logger
	InstallationKey    string
	InstallationSecret string
}

// configureIPRequest represents the request body for /api/installations/configure-ips
type configureIPRequest struct {
	InstallationKey   string                         `json:"installationKey"`
	IP                string                         `json:"ip"`
	DomainRequestName string                         `json:"domainRequestName"`
	DomainRoutes      []domainrequestsv1.DomainRoute `json:"domainRoutes,omitempty"`
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

// generateHAProxyConfig generates HAProxy configuration for routing traffic
func (r *DomainRequestReconciler) generateHAProxyConfig(domainRequest *domainrequestsv1.DomainRequest) string {
	config := `global
    maxconn 4096
    stats socket /var/run/haproxy.sock mode 660 level admin expose-fd listeners

defaults
    mode http
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
    option forwardfor
    option http-server-close

frontend http_frontend
    bind *:80
    redirect scheme https code 301

frontend https_frontend
    bind *:443 ssl crt /etc/haproxy/certs/tls.pem
`

	// Add domain-based routing if DomainRoutes are configured
	if len(domainRequest.Spec.DomainRoutes) > 0 {
		for i, route := range domainRequest.Spec.DomainRoutes {
			aclName := fmt.Sprintf("is_domain_%d", i)
			backendName := fmt.Sprintf("domain_backend_%d", i)
			config += fmt.Sprintf("\n    # Route for domain: %s\n", route.Domain)
			config += fmt.Sprintf("    acl %s hdr(host) -i %s\n", aclName, route.Domain)
			config += fmt.Sprintf("    use_backend %s if %s\n", backendName, aclName)
		}
		// Use first domain route as default backend
		config += "\n    default_backend domain_backend_0\n"
	} else if domainRequest.Spec.IngressBackend != nil {
		// Use IngressBackend as default if no domain routes
		config += "\n    default_backend service_backend\n"
	} else {
		// Fallback to default backend
		config += "\n    default_backend default_backend\n"
	}

	// Add backends for domain routes
	for i, route := range domainRequest.Spec.DomainRoutes {
		backendName := fmt.Sprintf("domain_backend_%d", i)
		config += fmt.Sprintf("\nbackend %s\n", backendName)
		config += fmt.Sprintf("    server backend%d %s.%s.svc.cluster.local:%d check\n",
			i, route.ServiceName, route.ServiceNamespace, route.ServicePort)
	}

	// Add IngressBackend if configured
	if domainRequest.Spec.IngressBackend != nil {
		backend := domainRequest.Spec.IngressBackend
		config += "\nbackend service_backend\n"
		config += fmt.Sprintf("    server backend1 %s.%s.svc.cluster.local:%d check\n",
			backend.ServiceName, backend.ServiceNamespace, backend.ServicePort)
	}

	// Add default backend if no other backends configured
	if len(domainRequest.Spec.DomainRoutes) == 0 && domainRequest.Spec.IngressBackend == nil {
		config += "\nbackend default_backend\n"
		config += "    server default1 frontend.kloudlite.svc.cluster.local:3000 check\n"
	}

	return config
}

// createHAProxyConfigMap creates or updates the HAProxy ConfigMap
func (r *DomainRequestReconciler) createHAProxyConfigMap(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) error {
	configMapName := fmt.Sprintf("%s-haproxy-config", domainRequest.Name)
	haproxyConfig := r.generateHAProxyConfig(domainRequest)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: "kloudlite",
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, configMap, func() error {
		if configMap.Data == nil {
			configMap.Data = make(map[string]string)
		}
		configMap.Data["haproxy.cfg"] = haproxyConfig

		// Set owner reference
		if err := controllerutil.SetControllerReference(domainRequest, configMap, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}

// createHAProxyPod creates or updates the HAProxy pod with hostNetwork
func (r *DomainRequestReconciler) createHAProxyPod(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) error {
	if domainRequest.Status.OriginCertificateSecretName == "" {
		return fmt.Errorf("origin certificate secret not yet created")
	}

	podName := fmt.Sprintf("%s-haproxy", domainRequest.Name)
	configMapName := fmt.Sprintf("%s-haproxy-config", domainRequest.Name)

	// Check if pod already exists
	existingPod := &corev1.Pod{}
	err := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: "kloudlite"}, existingPod)
	if err == nil {
		// Pod exists, delete it to force recreate (pods are immutable)
		logger.Info("Deleting existing HAProxy pod to recreate with new configuration")
		if err := r.Delete(ctx, existingPod); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete existing HAProxy pod: %w", err)
		}
		// Return and let the next reconcile create the new pod
		return nil
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check if HAProxy pod exists: %w", err)
	}

	// Create new pod
	podSpec := corev1.PodSpec{
		HostNetwork: true,
		DNSPolicy:   corev1.DNSClusterFirstWithHostNet,
		Containers: []corev1.Container{
			{
				Name:  "haproxy",
				Image: "haproxy:2.8-alpine",
				SecurityContext: &corev1.SecurityContext{
					RunAsUser:  pointerInt64(0),
					RunAsGroup: pointerInt64(0),
					Capabilities: &corev1.Capabilities{
						Add: []corev1.Capability{
							"NET_BIND_SERVICE",
						},
					},
				},
				Ports: []corev1.ContainerPort{
					{
						Name:          "http",
						ContainerPort: 80,
						HostPort:      80,
						Protocol:      corev1.ProtocolTCP,
					},
					{
						Name:          "https",
						ContainerPort: 443,
						HostPort:      443,
						Protocol:      corev1.ProtocolTCP,
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "haproxy-config",
						MountPath: "/usr/local/etc/haproxy",
						ReadOnly:  true,
					},
					{
						Name:      "tls-certs",
						MountPath: "/etc/haproxy/certs",
						ReadOnly:  true,
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    *NewQuantity("100m"),
						corev1.ResourceMemory: *NewQuantity("128Mi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    *NewQuantity("500m"),
						corev1.ResourceMemory: *NewQuantity("256Mi"),
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "haproxy-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: configMapName,
						},
					},
				},
			},
			{
				Name: "tls-certs",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: domainRequest.Status.OriginCertificateSecretName,
					},
				},
			},
		},
	}

	// Add nodeSelector if NodeName is specified
	if domainRequest.Spec.NodeName != "" {
		podSpec.NodeSelector = map[string]string{
			"kubernetes.io/hostname": domainRequest.Spec.NodeName,
		}
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: "kloudlite",
			Labels: map[string]string{
				"app":                     "haproxy",
				"domain-request":          domainRequest.Name,
				"kloudlite.io/managed-by": "domainrequest-controller",
			},
		},
		Spec: podSpec,
	}

	// Set owner reference so pod is automatically cleaned up when DomainRequest is deleted
	if err := controllerutil.SetControllerReference(domainRequest, pod, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference: %w", err)
	}

	// Create the pod (cannot use CreateOrUpdate because pods are immutable)
	if err := r.Create(ctx, pod); err != nil {
		return fmt.Errorf("failed to create HAProxy pod: %w", err)
	}

	logger.Info("HAProxy pod created successfully", zap.String("podName", podName))
	return nil
}

// checkHAProxyReady checks if the HAProxy pod is ready
func (r *DomainRequestReconciler) checkHAProxyReady(ctx context.Context, podName string, logger *zap.Logger) (bool, error) {
	pod := &corev1.Pod{}
	err := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: "kloudlite"}, pod)
	if err != nil {
		return false, err
	}

	// Check if pod is running and all containers are ready
	if pod.Status.Phase != corev1.PodRunning {
		return false, nil
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true, nil
		}
	}

	return false, nil
}

// deleteHAProxyResources deletes HAProxy pod and ConfigMap
func (r *DomainRequestReconciler) deleteHAProxyResources(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) error {
	podName := fmt.Sprintf("%s-haproxy", domainRequest.Name)
	configMapName := fmt.Sprintf("%s-haproxy-config", domainRequest.Name)

	// Delete pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: "kloudlite",
		},
	}
	if err := r.Delete(ctx, pod); err != nil && !errors.IsNotFound(err) {
		logger.Error("Failed to delete HAProxy pod", zap.Error(err))
		return err
	}

	// Delete ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: "kloudlite",
		},
	}
	if err := r.Delete(ctx, configMap); err != nil && !errors.IsNotFound(err) {
		logger.Error("Failed to delete HAProxy ConfigMap", zap.Error(err))
		return err
	}

	return nil
}

// Helper function to create resource quantities
func NewQuantity(value string) *resource.Quantity {
	q, _ := resource.ParseQuantity(value)
	return &q
}

// pointerInt64 returns a pointer to an int64 value
func pointerInt64(i int64) *int64 {
	return &i
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
		if errors.IsNotFound(err) {
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
		// New flow: Download installation's origin certificate first
		return r.handleOriginCertificateDownload(ctx, domainRequest, logger)
	case "CertificateDownloading":
		return r.handleOriginCertificateDownload(ctx, domainRequest, logger)
	case "CertificateReady":
		// Create HAProxy with landing page (or configured backend)
		return r.handleHAProxyCreation(ctx, domainRequest, logger)
	case "HAProxyCreating":
		return r.handleHAProxyStatusCheck(ctx, domainRequest, logger)
	case "HAProxyReady":
		// HAProxy is ready, now configure DNS/IP
		return r.handleIPRegistration(ctx, domainRequest, logger)
	case "IPRegistering":
		return r.handleIPRegistration(ctx, domainRequest, logger)
	case "Ready":
		logger.Info("DomainRequest is ready, no action needed")
		return reconcile.Result{RequeueAfter: 24 * time.Hour}, nil
	case "Failed":
		// Retry the failed operation after 30 seconds
		logger.Info("DomainRequest failed, determining which step to retry")

		// Determine which step failed based on what's been completed
		if domainRequest.Status.CertificateID != "" && domainRequest.Status.CertificateSecretName == "" {
			// Certificate was generated but not downloaded
			logger.Info("Retrying certificate download")
			return r.handleCertificateDownload(ctx, domainRequest, logger)
		} else if domainRequest.Status.LastIPRegistrationTime != nil && domainRequest.Status.CertificateID == "" {
			// IP was registered but certificate generation failed
			logger.Info("Retrying certificate generation")
			return r.handleCertificateGeneration(ctx, domainRequest, logger)
		} else {
			// IP registration failed or unknown state
			logger.Info("Retrying IP registration")
			return r.handleIPRegistration(ctx, domainRequest, logger)
		}
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

	// Call console API to register IP and create DNS records
	reqBody := configureIPRequest{
		InstallationKey:   r.InstallationKey,
		IP:                ipAddress,
		DomainRequestName: domainRequest.Name,
		DomainRoutes:      domainRequest.Spec.DomainRoutes,
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

	// Update status with registration details
	// New flow: Transition to Ready state after IP registration (certificate and HAProxy already set up)
	if err := statusutil.UpdateStatusWithRetry(ctx, r.Client, domainRequest, func() error {
		domainRequest.Status.State = "Ready"
		domainRequest.Status.Message = fmt.Sprintf("DomainRequest is ready - origin cert downloaded, HAProxy created, DNS configured (%d routes)", configResp.RouteRecordsCreated)
		domainRequest.Status.Domain = configResp.SSHDomain
		domainRequest.Status.Subdomain = configResp.Subdomain
		domainRequest.Status.DNSRecordIDs = dnsRecordIDs
		now := metav1.Now()
		domainRequest.Status.LastIPRegistrationTime = &now
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
			Namespace: "kloudlite",
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

// handleOriginCertificateDownload downloads the installation's origin certificate
func (r *DomainRequestReconciler) handleOriginCertificateDownload(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) (reconcile.Result, error) {
	logger.Info("Downloading installation origin certificate")

	// Call console API to get origin certificate
	url := fmt.Sprintf("%s/api/installations/get-origin-certificate?installationKey=%s",
		consoleAPIBaseURL, r.InstallationKey)

	resp, err := r.callConsoleAPI(ctx, url, "GET", nil, r.InstallationSecret, logger)
	if err != nil {
		logger.Error("Failed to download origin certificate", zap.Error(err))
		domainRequest.Status.State = "Failed"
		domainRequest.Status.Message = fmt.Sprintf("Failed to download origin certificate: %v", err)
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
	secretName := fmt.Sprintf("%s-origin-cert", domainRequest.Name)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: domainRequest.Namespace,
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": []byte(certResp.Certificate),
			"tls.key": []byte(certResp.PrivateKey),
		},
	}

	// Set owner reference
	if err := controllerutil.SetControllerReference(domainRequest, secret, r.Scheme); err != nil {
		logger.Error("Failed to set owner reference on Secret", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Create or update secret
	existingSecret := &corev1.Secret{}
	err = r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: domainRequest.Namespace}, existingSecret)
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

	// Update status
	domainRequest.Status.State = "CertificateReady"
	domainRequest.Status.Message = "Origin certificate downloaded and stored"
	domainRequest.Status.OriginCertificateSecretName = secretName

	if err := r.Status().Update(ctx, domainRequest); err != nil {
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
	ready, err := r.checkHAProxyReady(ctx, domainRequest.Status.HAProxyPodName, logger)
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
		// Delete HAProxy resources if they exist
		if domainRequest.Status.HAProxyPodName != "" {
			logger.Info("Deleting HAProxy resources")
			if err := r.deleteHAProxyResources(ctx, domainRequest, logger); err != nil {
				logger.Error("Failed to delete HAProxy resources", zap.Error(err))
				// Continue with deletion even if cleanup fails
			}
		}

		// Perform cleanup here if needed (e.g., delete DNS records via API)
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

// SetupWithManager sets up the controller with the Manager
func (r *DomainRequestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&domainrequestsv1.DomainRequest{}).
		Owns(&corev1.Secret{}).    // Watch Secrets owned by DomainRequest
		Owns(&corev1.Pod{}).       // Watch HAProxy Pods owned by DomainRequest
		Owns(&corev1.ConfigMap{}). // Watch HAProxy ConfigMaps owned by DomainRequest
		Complete(r)
}
