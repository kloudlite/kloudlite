package wmingress

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"

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

// IngressReconciler reconciles Ingress resources and updates routing configuration
type IngressReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	Logger           *zap.Logger
	IngressClassName string
	HTTPPort         int
	HTTPSPort        int

	// HTTP server components
	router      *Router
	tlsManager  *TLSManager
	configMutex sync.RWMutex
	currentHash string
}

// SetupWithManager sets up the controller with the Manager
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Watch all Ingress resources cluster-wide with configured ingressClassName
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkingv1.Ingress{}).
		Watches(&corev1.Secret{}, &secretEventHandler{reconciler: r}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return r.shouldProcessResource(e.Object)
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return r.shouldProcessResource(e.ObjectNew)
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return r.shouldProcessResource(e.Object)
			},
		}).
		Complete(r)
}

// shouldProcessResource checks if the resource should be processed
func (r *IngressReconciler) shouldProcessResource(obj client.Object) bool {
	switch o := obj.(type) {
	case *networkingv1.Ingress:
		return r.IngressClassName == "" || o.Spec.IngressClassName != nil || *o.Spec.IngressClassName == r.IngressClassName
	case *corev1.Secret:
		// Process all TLS secrets (we'll check if they're used by Ingress later)
		return o.Type == corev1.SecretTypeTLS
	default:
		return false
	}
}

// Reconcile reconciles an Ingress or Secret resource
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Logger.With(
		zap.String("resource", req.Name),
		zap.String("namespace", req.Namespace),
	)

	logger.Info("Reconciling resource")

	// List all Ingress resources cluster-wide
	ingressList := &networkingv1.IngressList{}
	if err := r.List(ctx, ingressList); err != nil {
		logger.Error("Failed to list Ingress resources", zap.Error(err))
		return ctrl.Result{}, err
	}

	// Filter only ingresses with matching ingressClassName
	var matchedIngresses []networkingv1.Ingress
	for _, ing := range ingressList.Items {
		if ing.Spec.IngressClassName != nil && *ing.Spec.IngressClassName == r.IngressClassName {
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

		if err := r.router.UpdateRoutes(routes); err != nil {
			logger.Error("Failed to update routes", zap.Error(err))
			return ctrl.Result{}, err
		}

		// Update TLS certificates
		if err := r.updateTLSCertificates(ctx, matchedIngresses); err != nil {
			logger.Error("Failed to update TLS certificates", zap.Error(err))
			return ctrl.Result{}, err
		}

		r.currentHash = configHash
		logger.Info("Router configuration updated successfully")
	} else {
		logger.Info("Configuration unchanged, skipping update")
	}

	return ctrl.Result{}, nil
}

// StartServers starts the HTTP and HTTPS servers
func (r *IngressReconciler) StartServers(ctx context.Context) error {
	// Initialize router
	r.router = NewRouter(r.Logger)

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
func (r *IngressReconciler) updateTLSCertificates(ctx context.Context, ingresses []networkingv1.Ingress) error {
	certificates := make(map[string]*TLSCertificate)

	for _, ingress := range ingresses {
		for _, tls := range ingress.Spec.TLS {
			if tls.SecretName == "" {
				continue
			}

			// Fetch the Secret
			secret := &corev1.Secret{}
			if err := r.Get(ctx, client.ObjectKey{
				Name:      tls.SecretName,
				Namespace: ingress.Namespace,
			}, secret); err != nil {
				r.Logger.Error("Failed to get TLS secret",
					zap.String("secret", tls.SecretName),
					zap.Error(err),
				)
				continue
			}

			// Extract certificate and key
			certData, ok := secret.Data["tls.crt"]
			if !ok {
				r.Logger.Error("tls.crt not found in secret", zap.String("secret", tls.SecretName))
				continue
			}

			keyData, ok := secret.Data["tls.key"]
			if !ok {
				r.Logger.Error("tls.key not found in secret", zap.String("secret", tls.SecretName))
				continue
			}

			cert := &TLSCertificate{
				Hosts:    tls.Hosts,
				CertPEM:  certData,
				KeyPEM:   keyData,
				SecretID: fmt.Sprintf("%s/%s", ingress.Namespace, tls.SecretName),
			}

			for _, host := range tls.Hosts {
				certificates[host] = cert
			}
		}
	}

	return r.tlsManager.UpdateCertificates(certificates)
}

// calculateConfigHash computes a hash of the routing configuration
func (r *IngressReconciler) calculateConfigHash(routes []*Route) string {
	data, _ := json.Marshal(routes)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// secretEventHandler handles Secret events
type secretEventHandler struct {
	reconciler *IngressReconciler
}

func (h *secretEventHandler) Create(ctx context.Context, e event.CreateEvent, q workqueue.TypedRateLimitingInterface[ctrl.Request]) {
	// Trigger reconciliation of all Ingress resources
	h.enqueueAll(ctx, q)
}

func (h *secretEventHandler) Update(ctx context.Context, e event.UpdateEvent, q workqueue.TypedRateLimitingInterface[ctrl.Request]) {
	h.enqueueAll(ctx, q)
}

func (h *secretEventHandler) Delete(ctx context.Context, e event.DeleteEvent, q workqueue.TypedRateLimitingInterface[ctrl.Request]) {
	h.enqueueAll(ctx, q)
}

func (h *secretEventHandler) Generic(ctx context.Context, e event.GenericEvent, q workqueue.TypedRateLimitingInterface[ctrl.Request]) {
	h.enqueueAll(ctx, q)
}

func (h *secretEventHandler) enqueueAll(ctx context.Context, q workqueue.TypedRateLimitingInterface[ctrl.Request]) {
	// List all Ingress resources and enqueue them
	ingressList := &networkingv1.IngressList{}
	if err := h.reconciler.List(ctx, ingressList); err != nil {
		h.reconciler.Logger.Error("Failed to list Ingress resources for secret event", zap.Error(err))
		return
	}

	for _, ing := range ingressList.Items {
		if ing.Spec.IngressClassName != nil && *ing.Spec.IngressClassName == h.reconciler.IngressClassName {
			q.Add(ctrl.Request{
				NamespacedName: client.ObjectKey{
					Name:      ing.Name,
					Namespace: ing.Namespace,
				},
			})
		}
	}
}
