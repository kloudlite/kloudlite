package wmingress

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ReconciliationMetrics tracks reconciliation performance
type ReconciliationMetrics struct {
	mu                    sync.Mutex
	TotalReconciliations  int64
	SkippedReconciliations int64
	FullRebuilds          int64
	IncrementalRebuilds   int64
	AverageReconcileTime  time.Duration
	totalReconcileTime    time.Duration

	// Race condition tracking
	ConcurrentReconciliations int64
	ReconciliationConflicts   int64
	RouteUpdateConflicts      int64
	CertUpdateConflicts       int64
}

// RecordReconcile records a reconciliation attempt
func (m *ReconciliationMetrics) RecordReconcile(skipped bool, fullRebuild bool, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalReconciliations++
	if skipped {
		m.SkippedReconciliations++
	}
	if fullRebuild {
		m.FullRebuilds++
	} else {
		m.IncrementalRebuilds++
	}
	m.totalReconcileTime += duration
	m.AverageReconcileTime = m.totalReconcileTime / time.Duration(m.TotalReconciliations)
}

// GetStats returns current metrics
func (m *ReconciliationMetrics) GetStats() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	return map[string]interface{}{
		"total_reconciliations":    m.TotalReconciliations,
		"skipped_reconciliations":  m.SkippedReconciliations,
		"full_rebuilds":           m.FullRebuilds,
		"incremental_rebuilds":    m.IncrementalRebuilds,
		"average_reconcile_time":  m.AverageReconcileTime.String(),
		"concurrent_reconciliations": m.ConcurrentReconciliations,
		"reconciliation_conflicts":   m.ReconciliationConflicts,
	}
}

// IngressReconciler reconciles Ingress resources and updates routing configuration
type IngressReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	Logger           *zap.Logger
	IngressClassName string
	HTTPPort         int
	HTTPSPort        int

	// Wildcard TLS configuration
	WildcardDomain          string // e.g., "khost.dev"
	WildcardSecretName      string // e.g., "kloudlite-wildcard-cert-tls"
	WildcardSecretNamespace string // e.g., "kloudlite"

	// The namespace where this ingress controller is running
	OwnNamespace string

	// Registry access control - username for path restriction
	// When set, write operations to cr.* domains are restricted to /v2/{username}/*
	RegistryUsername string

	// Optimization: Force full rebuild on every event (for debugging)
	ForceFullRebuild bool

	// HTTP server components
	router      *Router
	tlsManager  *TLSManager
	configMutex sync.RWMutex
	currentHash string

	// Caching for efficient reconciliation
	ingressCache map[string]string // key: namespace/name, value: resourceVersion hash

	// Metrics tracking
	metrics ReconciliationMetrics
}

// SetupWithManager sets up the controller with the Manager
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize caches
	r.ingressCache = make(map[string]string)

	// Load configuration from environment if not already set
	if r.ForceFullRebuild == false {
		r.ForceFullRebuild = os.Getenv("WM_INGRESS_FORCE_FULL_REBUILD") == "true"
	}
	if r.ForceFullRebuild {
		r.Logger.Warn("Force full rebuild enabled - O(n) reconciliation mode")
	}

	// Watch all Ingress resources cluster-wide with configured ingressClassName
	// Use optimized predicate to filter events efficiently
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1.Ingress{}).
		Watches(&corev1.Secret{}, &secretEventHandler{reconciler: r}).
		WithEventFilter(r.ingressPredicate()).
		Complete(r)
}

// ingressPredicate creates an optimized event filter for Ingress resources
func (r *IngressReconciler) ingressPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return r.shouldProcessResource(e.Object)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Skip updates that don't change the spec (only metadata changes)
			if !r.hasSpecChanged(e.ObjectOld, e.ObjectNew) {
				return false
			}
			return r.shouldProcessResource(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return r.shouldProcessResource(e.Object)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return r.shouldProcessResource(e.Object)
		},
	}
}

// hasSpecChanged checks if the spec of a resource has changed
func (r *IngressReconciler) hasSpecChanged(oldObj, newObj client.Object) bool {
	switch old := oldObj.(type) {
	case *networkingv1.Ingress:
		new, ok := newObj.(*networkingv1.Ingress)
		if !ok {
			return true
		}
		// Check if Ingress spec changed
		if old.ResourceVersion == new.ResourceVersion {
			return false
		}
		// Compare specs
		oldSpec, _ := json.Marshal(old.Spec)
		newSpec, _ := json.Marshal(new.Spec)
		return string(oldSpec) != string(newSpec)
	case *corev1.Secret:
		new, ok := newObj.(*corev1.Secret)
		if !ok {
			return true
		}
		// Check if Secret data changed
		if old.ResourceVersion == new.ResourceVersion {
			return false
		}
		// Compare data
		oldData, _ := json.Marshal(old.Data)
		newData, _ := json.Marshal(new.Data)
		return string(oldData) != string(newData)
	default:
		return true
	}
}

func (r *IngressReconciler) shouldProcessIngress(ing *networkingv1.Ingress) bool {
	return r.IngressClassName == "" || ing.Spec.IngressClassName == nil || *ing.Spec.IngressClassName == r.IngressClassName
}

// isWorkmachineNamespace checks if the namespace is a workmachine namespace
func isWorkmachineNamespace(namespace string) bool {
	return strings.HasPrefix(namespace, "wm-")
}

// isExposedPortHost checks if the host is for an exposed port (p{port}-hash.subdomain)
// e.g., p3000-a1b2c3d4.beanbag.khost.dev
func isExposedPortHost(host string) bool {
	parts := strings.Split(host, ".")
	if len(parts) < 1 {
		return false
	}
	prefix := parts[0]
	// Check if prefix starts with 'p' followed by digits and a dash
	if !strings.HasPrefix(prefix, "p") {
		return false
	}
	// Find the dash after the port number
	dashIdx := strings.Index(prefix[1:], "-")
	if dashIdx == -1 {
		return false
	}
	portStr := prefix[1 : 1+dashIdx]
	// Verify it's a valid port number (all digits)
	for _, c := range portStr {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// isValidHost checks if the host matches the pattern *.*.{WildcardDomain}
// e.g., vscode-58890cd7.beanbag.khost.dev matches *.*.khost.dev
func (r *IngressReconciler) isValidHost(host string) bool {
	if r.WildcardDomain == "" {
		return true // No filtering if WildcardDomain is not set
	}
	// Host should be like: prefix.subdomain.khost.dev
	// Must end with .{WildcardDomain} and have at least 2 parts before it
	suffix := "." + r.WildcardDomain
	if !strings.HasSuffix(host, suffix) {
		return false
	}
	// Ensure there's a subdomain before the wildcard domain
	prefix := strings.TrimSuffix(host, suffix)
	return strings.Contains(prefix, ".")
}

// shouldProcessResource checks if the resource should be processed
func (r *IngressReconciler) shouldProcessResource(obj client.Object) bool {
	switch o := obj.(type) {
	case *networkingv1.Ingress:
		return r.shouldProcessIngress(o)
	case *corev1.Secret:
		// For Secret events, we need to process them to check if they're used
		// This is checked later in the secretEventHandler
		return o.Type == corev1.SecretTypeTLS
	default:
		return false
	}
}

// Reconcile reconciles an Ingress or Secret resource
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	startTime := time.Now()
	logger := r.Logger.With(
		zap.String("resource", req.Name),
		zap.String("namespace", req.Namespace),
	)

	logger.Info("Reconciling resource")

	// Track concurrent reconciliations
	r.metrics.mu.Lock()
	r.metrics.ConcurrentReconciliations++
	currentConcurrent := r.metrics.ConcurrentReconciliations
	if currentConcurrent > 1 {
		r.metrics.ReconciliationConflicts++
	}
	r.metrics.mu.Unlock()

	if currentConcurrent > 1 {
		logger.Debug("Concurrent reconciliation detected",
			zap.Int64("active_reconciliations", currentConcurrent),
		)
	}

	// Decrement concurrent counter when done (defer ensures this runs even on error)
	defer func() {
		r.metrics.mu.Lock()
		r.metrics.ConcurrentReconciliations--
		r.metrics.mu.Unlock()
	}()

	// Determine if we need a full rebuild or incremental update
	forceFullRebuild := r.ForceFullRebuild

	// Check if this is a Secret event (Secret events always need full rebuild)
	// because we don't know which Ingresses use the Secret without listing
	secret := &corev1.Secret{}
	if err := r.Get(ctx, req.NamespacedName, secret); err == nil {
		if secret.Type == corev1.SecretTypeTLS {
			forceFullRebuild = true
			logger.Debug("Secret event detected, forcing full rebuild")
		}
	}

	// If force full rebuild is disabled, try incremental reconciliation
	if !forceFullRebuild {
		// Try to get the Ingress resource
		ingress := &networkingv1.Ingress{}
		if err := r.Get(ctx, req.NamespacedName, ingress); err != nil {
			// Ingress not found, need full rebuild to handle deletion
			forceFullRebuild = true
			logger.Debug("Ingress not found, forcing full rebuild for deletion")
		} else {
			// Check if this Ingress should be processed
			if !r.shouldProcessIngress(ingress) {
				logger.Info("Ingress does not match ingress class, skipping")
				r.metrics.RecordReconcile(true, false, time.Since(startTime))
				return ctrl.Result{}, nil
			}

			// Calculate hash for this specific Ingress
			ingressKey := fmt.Sprintf("%s/%s", ingress.Namespace, ingress.Name)
			ingressHash := r.calculateIngressHash(ingress)

			// Check if the Ingress actually changed
			if cachedHash, exists := r.ingressCache[ingressKey]; exists && cachedHash == ingressHash {
				logger.Info("Ingress unchanged, skipping reconciliation")
				r.metrics.RecordReconcile(true, false, time.Since(startTime))
				return ctrl.Result{}, nil
			}

			logger.Debug("Ingress changed, proceeding with reconciliation")
		}
	}

	// List all Ingress resources cluster-wide
	ingressList := &networkingv1.IngressList{}
	if err := r.List(ctx, ingressList); err != nil {
		logger.Error("Failed to list Ingress resources", zap.Error(err))
		return ctrl.Result{}, err
	}

	// Filter only ingresses with matching ingressClassName
	var matchedIngresses []networkingv1.Ingress
	for _, ing := range ingressList.Items {
		if r.shouldProcessIngress(&ing) {
			matchedIngresses = append(matchedIngresses, ing)
		}
	}

	logger.Info("Found matching Ingresses", zap.Int("count", len(matchedIngresses)))

	// Build routing configuration
	routes, err := r.buildRoutes(ctx, matchedIngresses)
	if err != nil {
		logger.Error("Failed to build routes", zap.Error(err))
		return ctrl.Result{}, err
	}

	// Calculate config hash
	configHash := r.calculateConfigHash(routes)

	// Update router if config changed
	r.configMutex.Lock()
	defer r.configMutex.Unlock()

	if configHash != r.currentHash {
		logger.Info("Configuration changed, updating router", zap.String("hash", configHash))

		// Atomic update: Update routes first, then certificates
		// This ensures that if a request arrives between updates, it will either
		// see the old routes with old certs, or new routes with new certs
		// Both UpdateRoutes and UpdateCertificates use internal mutexes for safety
		startRouteUpdate := time.Now()
		if err := r.router.UpdateRoutes(routes); err != nil {
			logger.Error("Failed to update routes", zap.Error(err))
			return ctrl.Result{}, err
		}
		routeUpdateDuration := time.Since(startRouteUpdate)
		logger.Debug("Route update completed", zap.Duration("duration", routeUpdateDuration))

		// Update TLS certificates after routes are updated
		// This is safe because certificate lookup happens at connection time
		startCertUpdate := time.Now()
		if err := r.updateTLSCertificates(ctx, matchedIngresses); err != nil {
			logger.Error("Failed to update TLS certificates", zap.Error(err))
			return ctrl.Result{}, err
		}
		certUpdateDuration := time.Since(startCertUpdate)
		logger.Debug("Certificate update completed", zap.Duration("duration", certUpdateDuration))

		r.currentHash = configHash
		logger.Info("Router configuration updated successfully",
			zap.Int("route_count", len(routes)),
			zap.Duration("route_update_time", routeUpdateDuration),
			zap.Duration("cert_update_time", certUpdateDuration),
		)
	} else {
		logger.Info("Configuration unchanged, skipping update")
	}

	// Update cache with current Ingress hashes
	r.updateIngressCache(matchedIngresses)

	// Record metrics
	duration := time.Since(startTime)
	r.metrics.RecordReconcile(false, forceFullRebuild, duration)
	logger.Info("Reconciliation completed",
		zap.Duration("duration", duration),
		zap.Bool("fullRebuild", forceFullRebuild),
	)

	return ctrl.Result{}, nil
}

// calculateIngressHash computes a hash of an Ingress resource for change detection
func (r *IngressReconciler) calculateIngressHash(ingress *networkingv1.Ingress) string {
	// Hash the spec to detect actual changes
	data, _ := json.Marshal(ingress.Spec)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// updateIngressCache updates the Ingress cache with current hashes
func (r *IngressReconciler) updateIngressCache(ingresses []networkingv1.Ingress) {
	newCache := make(map[string]string)
	for _, ing := range ingresses {
		key := fmt.Sprintf("%s/%s", ing.Namespace, ing.Name)
		newCache[key] = r.calculateIngressHash(&ing)
	}

	r.configMutex.Lock()
	defer r.configMutex.Unlock()
	r.ingressCache = newCache
}

// StartServers starts the HTTP and HTTPS servers
func (r *IngressReconciler) StartServers(ctx context.Context) error {
	// Initialize router with registry access control
	r.router = NewRouter(r.Logger, r.RegistryUsername)

	// Initialize TLS manager
	r.tlsManager = NewTLSManager(r.Logger)

	// Start HTTP server
	if err := r.router.StartHTTP(ctx, r.HTTPPort); err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	// Start HTTPS server
	if err := r.router.StartHTTPS(ctx, r.HTTPSPort, r.tlsManager); err != nil {
		return fmt.Errorf("failed to start HTTPS server: %w", err)
	}

	r.Logger.Info("HTTP/HTTPS servers started successfully",
		zap.Int("http-port", r.HTTPPort),
		zap.Int("https-port", r.HTTPSPort),
	)

	return nil
}

// buildRoutes builds routing configuration from Ingress resources
func (r *IngressReconciler) buildRoutes(ctx context.Context, ingresses []networkingv1.Ingress) ([]*Route, error) {
	var routes []*Route

	for _, ingress := range ingresses {
		for _, rule := range ingress.Spec.Rules {
			if rule.HTTP == nil {
				continue
			}

			// Skip hosts that don't match the wildcard domain pattern
			if !r.isValidHost(rule.Host) {
				r.Logger.Debug("Skipping host - does not match wildcard domain",
					zap.String("host", rule.Host),
					zap.String("wildcardDomain", r.WildcardDomain),
				)
				continue
			}

			// Only filter workspace services from OTHER WM namespaces, not our own
			// This prevents proxying workspace services (vscode, ttyd, etc.) to other users
			// while still allowing the owner to access their own workspace services
			if isWorkmachineNamespace(ingress.Namespace) && ingress.Namespace != r.OwnNamespace && !isExposedPortHost(rule.Host) {
				r.Logger.Debug("Skipping non-exposed-port host from other workmachine namespace",
					zap.String("host", rule.Host),
					zap.String("namespace", ingress.Namespace),
				)
				continue
			}

			for _, path := range rule.HTTP.Paths {
				route, err := r.createRoute(ctx, &ingress, rule.Host, &path)
				if err != nil {
					r.Logger.Error("Failed to create route",
						zap.String("ingress", ingress.Name),
						zap.String("host", rule.Host),
						zap.Error(err),
					)
					continue
				}

				r.Logger.Info("Created Route",
					zap.String("ingress", ingress.Name),
					zap.String("host", rule.Host),
				)

				routes = append(routes, route)
			}
		}
	}

	return routes, nil
}

// createRoute creates a route from an Ingress path
func (r *IngressReconciler) createRoute(
	ctx context.Context,
	ingress *networkingv1.Ingress,
	host string,
	path *networkingv1.HTTPIngressPath,
) (*Route, error) {
	backend := path.Backend
	if backend.Service == nil {
		return nil, fmt.Errorf("backend service is nil")
	}

	var port int32
	if backend.Service.Port.Number != 0 {
		port = backend.Service.Port.Number
	} else {
		// Resolve named port
		svc := &corev1.Service{}
		if err := r.Get(ctx, client.ObjectKey{
			Name:      backend.Service.Name,
			Namespace: ingress.Namespace,
		}, svc); err != nil {
			return nil, fmt.Errorf("failed to get service %s: %w", backend.Service.Name, err)
		}

		for _, p := range svc.Spec.Ports {
			if p.Name == backend.Service.Port.Name {
				port = p.Port
				break
			}
		}

		if port == 0 {
			return nil, fmt.Errorf("port %s not found in service %s", backend.Service.Port.Name, backend.Service.Name)
		}
	}

	// Construct backend URL
	backendURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
		backend.Service.Name,
		ingress.Namespace,
		port,
	)

	// Determine path type
	pathType := networkingv1.PathTypePrefix
	if path.PathType != nil {
		pathType = *path.PathType
	}

	return &Route{
		Host:        host,
		Path:        path.Path,
		PathType:    pathType,
		BackendURL:  backendURL,
		IngressName: ingress.Name,
		Namespace:   ingress.Namespace,
	}, nil
}

// updateTLSCertificates updates TLS certificates in the TLS manager
// Uses a central wildcard certificate from the configured namespace
func (r *IngressReconciler) updateTLSCertificates(ctx context.Context, ingresses []networkingv1.Ingress) error {
	certificates := make(map[string]*TLSCertificate)

	// If wildcard secret is not configured, skip TLS setup
	if r.WildcardSecretName == "" || r.WildcardSecretNamespace == "" {
		r.Logger.Warn("Wildcard TLS secret not configured, skipping TLS setup")
		return r.tlsManager.UpdateCertificates(certificates)
	}

	// Fetch the wildcard secret from central namespace
	secret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{
		Name:      r.WildcardSecretName,
		Namespace: r.WildcardSecretNamespace,
	}, secret); err != nil {
		r.Logger.Error("Failed to get wildcard TLS secret",
			zap.String("secret", r.WildcardSecretName),
			zap.String("namespace", r.WildcardSecretNamespace),
			zap.Error(err),
		)
		return err
	}

	// Extract certificate and key
	certData, ok := secret.Data["tls.crt"]
	if !ok {
		r.Logger.Error("tls.crt not found in wildcard secret",
			zap.String("secret", r.WildcardSecretName),
		)
		return fmt.Errorf("tls.crt not found in wildcard secret %s/%s", r.WildcardSecretNamespace, r.WildcardSecretName)
	}

	keyData, ok := secret.Data["tls.key"]
	if !ok {
		r.Logger.Error("tls.key not found in wildcard secret",
			zap.String("secret", r.WildcardSecretName),
		)
		return fmt.Errorf("tls.key not found in wildcard secret %s/%s", r.WildcardSecretNamespace, r.WildcardSecretName)
	}

	// Collect all valid hosts from ingresses and assign the wildcard cert
	for _, ingress := range ingresses {
		for _, rule := range ingress.Spec.Rules {
			if r.isValidHost(rule.Host) {
				certificates[rule.Host] = &TLSCertificate{
					Hosts:    []string{rule.Host},
					CertPEM:  certData,
					KeyPEM:   keyData,
					SecretID: fmt.Sprintf("%s/%s", r.WildcardSecretNamespace, r.WildcardSecretName),
				}
			}
		}
	}

	r.Logger.Info("Updated TLS certificates from wildcard secret",
		zap.String("secret", fmt.Sprintf("%s/%s", r.WildcardSecretNamespace, r.WildcardSecretName)),
		zap.Int("hostCount", len(certificates)),
	)

	return r.tlsManager.UpdateCertificates(certificates)
}

// calculateConfigHash computes a hash of the routing configuration
func (r *IngressReconciler) calculateConfigHash(routes []*Route) string {
	data, _ := json.Marshal(routes)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// GetMetrics returns current reconciliation metrics
func (r *IngressReconciler) GetMetrics() map[string]interface{} {
	metrics := r.metrics.GetStats()

	// Add router metrics
	if r.router != nil {
		routerMetrics := r.router.GetMetrics()
		for k, v := range routerMetrics {
			metrics["router_"+k] = v
		}
	}

	// Add TLS manager metrics
	if r.tlsManager != nil {
		tlsMetrics := r.tlsManager.GetMetrics()
		for k, v := range tlsMetrics {
			metrics["tls_"+k] = v
		}
	}

	// Add current config hash
	r.configMutex.RLock()
	metrics["current_config_hash"] = r.currentHash
	r.configMutex.RUnlock()

	return metrics
}

// secretEventHandler handles Secret events
type secretEventHandler struct {
	reconciler *IngressReconciler
}

func (h *secretEventHandler) Create(ctx context.Context, e event.CreateEvent, q workqueue.TypedRateLimitingInterface[ctrl.Request]) {
	secret, ok := e.Object.(*corev1.Secret)
	if !ok {
		return
	}

	// Only enqueue if this secret is actually used by Ingress resources
	if h.isSecretRelevant(ctx, secret) {
		h.reconciler.Logger.Info("Secret created, triggering reconciliation",
			zap.String("secret", secret.Name),
			zap.String("namespace", secret.Namespace),
		)
		// Enqueue a single request to trigger full rebuild
		q.Add(ctrl.Request{
			NamespacedName: client.ObjectKey{
				Name:      "secret-event",
				Namespace: secret.Namespace,
			},
		})
	}
}

func (h *secretEventHandler) Update(ctx context.Context, e event.UpdateEvent, q workqueue.TypedRateLimitingInterface[ctrl.Request]) {
	newSecret, ok := e.ObjectNew.(*corev1.Secret)
	if !ok {
		return
	}

	oldSecret, ok := e.ObjectOld.(*corev1.Secret)
	if !ok {
		return
	}

	// Check if secret data actually changed
	if h.hasSecretDataChanged(oldSecret, newSecret) {
		h.reconciler.Logger.Info("Secret data changed, triggering reconciliation",
			zap.String("secret", newSecret.Name),
			zap.String("namespace", newSecret.Namespace),
		)
		// Enqueue a single request to trigger full rebuild
		q.Add(ctrl.Request{
			NamespacedName: client.ObjectKey{
				Name:      "secret-event",
				Namespace: newSecret.Namespace,
			},
		})
	}
}

func (h *secretEventHandler) Delete(ctx context.Context, e event.DeleteEvent, q workqueue.TypedRateLimitingInterface[ctrl.Request]) {
	secret, ok := e.Object.(*corev1.Secret)
	if !ok {
		return
	}

	// Only enqueue if this secret was actually used by Ingress resources
	if h.isSecretRelevant(ctx, secret) {
		h.reconciler.Logger.Info("Secret deleted, triggering reconciliation",
			zap.String("secret", secret.Name),
			zap.String("namespace", secret.Namespace),
		)
		// Enqueue a single request to trigger full rebuild
		q.Add(ctrl.Request{
			NamespacedName: client.ObjectKey{
				Name:      "secret-event",
				Namespace: secret.Namespace,
			},
		})
	}
}

func (h *secretEventHandler) Generic(ctx context.Context, e event.GenericEvent, q workqueue.TypedRateLimitingInterface[ctrl.Request]) {
	// Generic events are rare for secrets, handle conservatively
	secret, ok := e.Object.(*corev1.Secret)
	if !ok {
		return
	}

	if h.isSecretRelevant(ctx, secret) {
		q.Add(ctrl.Request{
			NamespacedName: client.ObjectKey{
				Name:      "secret-event",
				Namespace: secret.Namespace,
			},
		})
	}
}

// isSecretRelevant checks if a secret is used by any Ingress
func (h *secretEventHandler) isSecretRelevant(ctx context.Context, secret *corev1.Secret) bool {
	// Check if this is the wildcard secret
	if secret.Namespace == h.reconciler.WildcardSecretNamespace &&
		secret.Name == h.reconciler.WildcardSecretName {
		return true
	}

	// Check if any Ingress references this secret
	ingressList := &networkingv1.IngressList{}
	if err := h.reconciler.List(ctx, ingressList); err != nil {
		h.reconciler.Logger.Error("Failed to list Ingress resources", zap.Error(err))
		return false
	}

	for _, ing := range ingressList.Items {
		if !h.reconciler.shouldProcessIngress(&ing) {
			continue
		}
		// Check TLS hosts in Ingress spec
		for _, tls := range ing.Spec.TLS {
			if tls.SecretName == secret.Name {
				// TLS secrets are in the same namespace as the Ingress
				return true
			}
		}
	}

	return false
}

// hasSecretDataChanged checks if secret data has actually changed
func (h *secretEventHandler) hasSecretDataChanged(old, new *corev1.Secret) bool {
	// Quick check: resource version
	if old.ResourceVersion == new.ResourceVersion {
		return false
	}

	// Compare data fields
	if len(old.Data) != len(new.Data) {
		return true
	}

	for key, oldValue := range old.Data {
		newValue, exists := new.Data[key]
		if !exists {
			return true
		}
		if len(oldValue) != len(newValue) {
			return true
		}
		if string(oldValue) != string(newValue) {
			return true
		}
	}

	return false
}
