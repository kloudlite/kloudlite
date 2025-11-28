package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	domainrequestv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	toolscache "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HostsCache maintains a cached list of hosts entries updated via watches
type HostsCache struct {
	mu    sync.RWMutex
	hosts []HostEntry

	// Configuration
	logger           *zap.Logger
	namespace        string // Namespace where router service is located
	routerServiceRef string
	wgServerIP       string // WireGuard server IP for vpn-check host

	// Kubernetes client for initial fetch and rebuilds
	k8sClient client.Client

	// Controller-runtime cache for watches
	cache cache.Cache
}

// HostsCacheConfig holds configuration for the hosts cache
type HostsCacheConfig struct {
	Namespace        string
	RouterServiceRef string
	WgServerIP       string // WireGuard server IP (default: 10.17.0.1)
	RestConfig       *rest.Config
	Scheme           *runtime.Scheme
}

// NewHostsCache creates a new HostsCache
func NewHostsCache(logger *zap.Logger, k8sClient client.Client, cfg HostsCacheConfig) (*HostsCache, error) {
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}
	if cfg.RouterServiceRef == "" {
		cfg.RouterServiceRef = "wm-ingress-controller"
	}
	if cfg.WgServerIP == "" {
		cfg.WgServerIP = "10.17.0.1"
	}

	// Create controller-runtime cache
	c, err := cache.New(cfg.RestConfig, cache.Options{
		Scheme: cfg.Scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	return &HostsCache{
		hosts:            make([]HostEntry, 0),
		logger:           logger,
		namespace:        cfg.Namespace,
		routerServiceRef: cfg.RouterServiceRef,
		wgServerIP:       cfg.WgServerIP,
		k8sClient:        k8sClient,
		cache:            c,
	}, nil
}

// Start starts the cache and sets up watches
func (hc *HostsCache) Start(ctx context.Context) error {
	hc.logger.Info("starting hosts cache")

	// Start the cache in background
	go func() {
		if err := hc.cache.Start(ctx); err != nil {
			hc.logger.Error("cache stopped with error", zap.Error(err))
		}
	}()

	// Wait for cache to sync
	if !hc.cache.WaitForCacheSync(ctx) {
		return fmt.Errorf("failed to sync cache")
	}

	hc.logger.Info("cache synced, setting up informers")

	// Setup informers for each resource type
	if err := hc.setupInformers(ctx); err != nil {
		return fmt.Errorf("failed to setup informers: %w", err)
	}

	// Do initial rebuild
	hc.rebuild(ctx)

	hc.logger.Info("hosts cache started successfully")
	return nil
}

// setupInformers sets up informers for all watched resources
func (hc *HostsCache) setupInformers(ctx context.Context) error {
	// Get informer for Ingresses
	ingressInformer, err := hc.cache.GetInformer(ctx, &networkingv1.Ingress{})
	if err != nil {
		return fmt.Errorf("failed to get ingress informer: %w", err)
	}

	// Get informer for Services
	serviceInformer, err := hc.cache.GetInformer(ctx, &corev1.Service{})
	if err != nil {
		return fmt.Errorf("failed to get service informer: %w", err)
	}

	// Get informer for Namespaces
	namespaceInformer, err := hc.cache.GetInformer(ctx, &corev1.Namespace{})
	if err != nil {
		return fmt.Errorf("failed to get namespace informer: %w", err)
	}

	// Get informer for DomainRequests
	domainRequestInformer, err := hc.cache.GetInformer(ctx, &domainrequestv1.DomainRequest{})
	if err != nil {
		return fmt.Errorf("failed to get domainrequest informer: %w", err)
	}

	// Get informer for Workspaces
	workspaceInformer, err := hc.cache.GetInformer(ctx, &workspacev1.Workspace{})
	if err != nil {
		return fmt.Errorf("failed to get workspace informer: %w", err)
	}

	// Create debounced rebuild function
	var rebuildTimer *time.Timer
	var rebuildMu sync.Mutex
	debouncedRebuild := func() {
		rebuildMu.Lock()
		defer rebuildMu.Unlock()

		if rebuildTimer != nil {
			rebuildTimer.Stop()
		}
		rebuildTimer = time.AfterFunc(500*time.Millisecond, func() {
			hc.rebuild(ctx)
		})
	}

	// Event handler for all resources
	handler := toolscache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			hc.logger.Debug("resource added, triggering rebuild")
			debouncedRebuild()
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			hc.logger.Debug("resource updated, triggering rebuild")
			debouncedRebuild()
		},
		DeleteFunc: func(obj interface{}) {
			hc.logger.Debug("resource deleted, triggering rebuild")
			debouncedRebuild()
		},
	}

	// Add event handlers
	if _, err := ingressInformer.AddEventHandler(handler); err != nil {
		return fmt.Errorf("failed to add ingress event handler: %w", err)
	}
	if _, err := serviceInformer.AddEventHandler(handler); err != nil {
		return fmt.Errorf("failed to add service event handler: %w", err)
	}
	if _, err := namespaceInformer.AddEventHandler(handler); err != nil {
		return fmt.Errorf("failed to add namespace event handler: %w", err)
	}
	if _, err := domainRequestInformer.AddEventHandler(handler); err != nil {
		return fmt.Errorf("failed to add domainrequest event handler: %w", err)
	}
	if _, err := workspaceInformer.AddEventHandler(handler); err != nil {
		return fmt.Errorf("failed to add workspace event handler: %w", err)
	}

	return nil
}

// rebuild rebuilds the entire hosts cache
func (hc *HostsCache) rebuild(ctx context.Context) {
	hc.logger.Debug("rebuilding hosts cache")

	hosts := make([]HostEntry, 0)

	// Get subdomain and domain from DomainRequest
	subdomain, domain, err := hc.getDomainInfo(ctx)
	if err != nil {
		hc.logger.Warn("failed to get domain info from DomainRequest, service hosts will not be added",
			zap.Error(err))
	}

	// Get the router service ClusterIP for ingresses
	routerIP, err := hc.getRouterIP(ctx)
	if err != nil {
		hc.logger.Warn("failed to get router service IP, ingress hosts will not be added",
			zap.String("service", hc.routerServiceRef),
			zap.Error(err))
	}

	// Get all ingresses cluster-wide and add their hosts
	if routerIP != "" {
		ingressHosts, err := hc.getIngressHosts(ctx, routerIP)
		if err != nil {
			hc.logger.Error("failed to get ingress hosts", zap.Error(err))
		} else {
			hosts = append(hosts, ingressHosts...)
		}
	}

	// Get services from kloudlite environment namespaces
	if subdomain != "" && domain != "" {
		serviceHosts, err := hc.getServiceHosts(ctx, subdomain, domain)
		if err != nil {
			hc.logger.Error("failed to get service hosts", zap.Error(err))
		} else {
			hosts = append(hosts, serviceHosts...)
		}

		// Get workspace services for VPN access (SSH, etc.)
		workspaceHosts, err := hc.getWorkspaceHosts(ctx, subdomain, domain)
		if err != nil {
			hc.logger.Error("failed to get workspace hosts", zap.Error(err))
		} else {
			hosts = append(hosts, workspaceHosts...)
		}

		// Add VPN check host entry (points to WireGuard server IP for connectivity check)
		vpnCheckHost := fmt.Sprintf("vpn-check.%s.%s", subdomain, domain)
		hosts = append(hosts, HostEntry{
			Hostname: vpnCheckHost,
			IP:       hc.wgServerIP,
			Type:     "vpn-check",
		})
	}

	// Update cache
	hc.mu.Lock()
	hc.hosts = hosts
	hc.mu.Unlock()

	hc.logger.Info("hosts cache rebuilt",
		zap.Int("count", len(hosts)),
		zap.String("subdomain", subdomain),
		zap.String("domain", domain))
}

// GetHosts returns the cached hosts entries (thread-safe)
func (hc *HostsCache) GetHosts() []HostEntry {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	// Return a copy to prevent race conditions
	result := make([]HostEntry, len(hc.hosts))
	copy(result, hc.hosts)
	return result
}

// getDomainInfo retrieves subdomain and domain from DomainRequest CR
// It extracts them from spec.domainRoutes[0].domain (e.g., "beanbag.khost.dev" -> subdomain="beanbag", domain="khost.dev")
func (hc *HostsCache) getDomainInfo(ctx context.Context) (subdomain, domain string, err error) {
	var domainRequestList domainrequestv1.DomainRequestList
	if err := hc.cache.List(ctx, &domainRequestList); err != nil {
		return "", "", fmt.Errorf("failed to list DomainRequests: %w", err)
	}

	if len(domainRequestList.Items) == 0 {
		return "", "", fmt.Errorf("no DomainRequest found")
	}

	// Use the first DomainRequest's spec.domainRoutes
	dr := domainRequestList.Items[0]
	if len(dr.Spec.DomainRoutes) == 0 {
		return "", "", fmt.Errorf("DomainRequest has no domain routes")
	}

	// Parse the full domain (e.g., "beanbag.khost.dev") into subdomain and domain
	fullDomain := dr.Spec.DomainRoutes[0].Domain
	parts := strings.SplitN(fullDomain, ".", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid domain format: %s (expected subdomain.domain)", fullDomain)
	}

	return parts[0], parts[1], nil
}

// getRouterIP gets the ClusterIP of the ingress controller service
func (hc *HostsCache) getRouterIP(ctx context.Context) (string, error) {
	routerSvc := &corev1.Service{}
	if err := hc.cache.Get(ctx, client.ObjectKey{
		Namespace: hc.namespace,
		Name:      hc.routerServiceRef,
	}, routerSvc); err != nil {
		return "", fmt.Errorf("failed to get router service: %w", err)
	}

	clusterIP := routerSvc.Spec.ClusterIP
	if clusterIP == "" || clusterIP == "None" {
		return "", fmt.Errorf("router service has no ClusterIP")
	}

	return clusterIP, nil
}

// getIngressHosts gets all ingress hosts cluster-wide
func (hc *HostsCache) getIngressHosts(ctx context.Context, routerIP string) ([]HostEntry, error) {
	var ingressList networkingv1.IngressList
	if err := hc.cache.List(ctx, &ingressList); err != nil {
		return nil, fmt.Errorf("failed to list ingresses: %w", err)
	}

	hosts := make([]HostEntry, 0)
	for i := range ingressList.Items {
		for j := range ingressList.Items[i].Spec.Rules {
			host := ingressList.Items[i].Spec.Rules[j].Host
			if host != "" {
				hosts = append(hosts, HostEntry{
					Hostname: host,
					IP:       routerIP,
					Type:     "ingress",
				})
			}
		}
	}

	return hosts, nil
}

// getServiceHosts gets all services from kloudlite environment namespaces
func (hc *HostsCache) getServiceHosts(ctx context.Context, subdomain, domain string) ([]HostEntry, error) {
	// List namespaces with kloudlite environment label
	var namespaceList corev1.NamespaceList
	if err := hc.cache.List(ctx, &namespaceList, client.HasLabels{kloudliteEnvironmentLabel}); err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	hosts := make([]HostEntry, 0)
	for i := range namespaceList.Items {
		ns := &namespaceList.Items[i]

		// Get environment name from label
		envName := ns.Labels[kloudliteEnvironmentLabel]
		if envName == "" {
			continue
		}

		// Try to get hash from Environment status first
		var envOwnerHash string
		var env environmentv1.Environment
		if err := hc.cache.Get(ctx, client.ObjectKey{Name: envName}, &env); err != nil {
			// Fall back to computing hash if Environment not found
			owner := ns.Annotations[kloudliteCreatedByAnnotation]
			if owner == "" {
				owner = "unknown"
			}
			envOwnerHash = generateHash(fmt.Sprintf("%s-%s", envName, owner))
			hc.logger.Debug("Environment not found, using computed hash",
				zap.String("environment", envName),
				zap.Error(err))
		} else if env.Status.Hash != "" {
			// Use hash from status
			envOwnerHash = env.Status.Hash
		} else {
			// Status.Hash not set yet, compute it
			envOwnerHash = generateHash(fmt.Sprintf("%s-%s", env.Spec.Name, env.Spec.OwnedBy))
		}

		// List services in this namespace
		var serviceList corev1.ServiceList
		if err := hc.cache.List(ctx, &serviceList, client.InNamespace(ns.Name)); err != nil {
			hc.logger.Warn("failed to list services in namespace",
				zap.String("namespace", ns.Name),
				zap.Error(err))
			continue
		}

		for j := range serviceList.Items {
			svc := &serviceList.Items[j]

			// Skip services without ClusterIP (headless services)
			if svc.Spec.ClusterIP == "" || svc.Spec.ClusterIP == "None" {
				continue
			}

			// Skip kubernetes default service
			if svc.Name == "kubernetes" && svc.Namespace == "default" {
				continue
			}

			// Build hostname: {service}-{hash}.{subdomain}.{domain}
			hostname := fmt.Sprintf("%s-%s.%s.%s", svc.Name, envOwnerHash, subdomain, domain)

			hosts = append(hosts, HostEntry{
				Hostname: hostname,
				IP:       svc.Spec.ClusterIP,
				Type:     "service",
			})
		}
	}

	return hosts, nil
}

// getWorkspaceHosts gets all workspace services and creates host entries for VPN access
func (hc *HostsCache) getWorkspaceHosts(ctx context.Context, subdomain, domain string) ([]HostEntry, error) {
	// List all workspaces
	var workspaceList workspacev1.WorkspaceList
	if err := hc.cache.List(ctx, &workspaceList); err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}

	hosts := make([]HostEntry, 0)
	for i := range workspaceList.Items {
		ws := &workspaceList.Items[i]

		// Get owner from workspace spec
		owner := ws.Spec.OwnedBy
		if owner == "" {
			continue
		}

		// Use hash from status if available, otherwise compute it
		var wsHash string
		if ws.Status.Hash != "" {
			wsHash = ws.Status.Hash
		} else {
			// Compute hash if not in status yet
			wsHash = generateHash(fmt.Sprintf("%s-%s", owner, ws.Name))
		}

		// Get workspace service ClusterIP
		// The service has the same name as the workspace
		svc := &corev1.Service{}
		if err := hc.cache.Get(ctx, client.ObjectKey{
			Namespace: ws.Namespace,
			Name:      ws.Name,
		}, svc); err != nil {
			hc.logger.Debug("workspace service not found",
				zap.String("workspace", ws.Name),
				zap.String("namespace", ws.Namespace),
				zap.Error(err))
			continue
		}

		// Skip services without ClusterIP
		if svc.Spec.ClusterIP == "" || svc.Spec.ClusterIP == "None" {
			continue
		}

		// Build hostname: {workspaceName}-{hash}.{subdomain}.{domain}
		hostname := fmt.Sprintf("%s-%s.%s.%s", ws.Name, wsHash, subdomain, domain)

		hosts = append(hosts, HostEntry{
			Hostname: hostname,
			IP:       svc.Spec.ClusterIP,
			Type:     "workspace",
		})
	}

	return hosts, nil
}

// generateHash generates an 8-character hash from the input string
func generateHash(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])[:8]
}
