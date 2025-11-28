package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	domainrequestv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Label used to identify kloudlite environment namespaces
	kloudliteEnvironmentLabel = "kloudlite.io/environment"
	// Annotation for environment owner
	kloudliteCreatedByAnnotation = "kloudlite.io/created-by"
)

// HostsHandler handles hosts configuration requests
type HostsHandler struct {
	logger           *zap.Logger
	k8sClient        client.Client
	namespace        string
	routerServiceRef string
}

// HostsHandlerConfig holds configuration for the hosts handler
type HostsHandlerConfig struct {
	Namespace        string // Namespace where router service is located
	RouterServiceRef string // Reference to the router service (defaults to wm-ingress-controller)
}

// HostEntry represents a single hosts file entry
type HostEntry struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Type     string `json:"type,omitempty"` // "ingress" or "service"
}

// HostsResponse represents the response from the hosts endpoint
type HostsResponse struct {
	Hosts []HostEntry `json:"hosts"`
}

// NewHostsHandler creates a new HostsHandler
func NewHostsHandler(logger *zap.Logger, k8sClient client.Client, cfg HostsHandlerConfig) *HostsHandler {
	if cfg.Namespace == "" {
		cfg.Namespace = "default"
	}
	if cfg.RouterServiceRef == "" {
		cfg.RouterServiceRef = "wm-ingress-controller"
	}
	return &HostsHandler{
		logger:           logger,
		k8sClient:        k8sClient,
		namespace:        cfg.Namespace,
		routerServiceRef: cfg.RouterServiceRef,
	}
}

// ServeHTTP implements http.Handler for hosts endpoint
func (h *HostsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	hosts := make([]HostEntry, 0)

	// Get subdomain and domain from DomainRequest
	subdomain, domain, err := h.getDomainInfo(ctx)
	if err != nil {
		h.logger.Warn("failed to get domain info from DomainRequest, service hosts will not be added",
			zap.Error(err))
	}

	// Get the router service ClusterIP for ingresses
	routerIP, err := h.getRouterIP(ctx)
	if err != nil {
		h.logger.Error("failed to get router service IP",
			zap.String("service", h.routerServiceRef),
			zap.Error(err))
		http.Error(w, "Router service not available", http.StatusInternalServerError)
		return
	}

	// Get all ingresses cluster-wide and add their hosts
	ingressHosts, err := h.getIngressHosts(ctx, routerIP)
	if err != nil {
		h.logger.Error("failed to get ingress hosts", zap.Error(err))
	} else {
		hosts = append(hosts, ingressHosts...)
	}

	// Get services from kloudlite environment namespaces
	if subdomain != "" && domain != "" {
		serviceHosts, err := h.getServiceHosts(ctx, subdomain, domain)
		if err != nil {
			h.logger.Error("failed to get service hosts", zap.Error(err))
		} else {
			hosts = append(hosts, serviceHosts...)
		}
	}

	h.logger.Debug("returning hosts",
		zap.Int("count", len(hosts)),
		zap.String("subdomain", subdomain),
		zap.String("domain", domain))

	response := HostsResponse{
		Hosts: hosts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode hosts response", zap.Error(err))
	}
}

// getDomainInfo retrieves subdomain and domain from DomainRequest CR
func (h *HostsHandler) getDomainInfo(ctx context.Context) (subdomain, domain string, err error) {
	var domainRequestList domainrequestv1.DomainRequestList
	if err := h.k8sClient.List(ctx, &domainRequestList); err != nil {
		return "", "", fmt.Errorf("failed to list DomainRequests: %w", err)
	}

	if len(domainRequestList.Items) == 0 {
		return "", "", fmt.Errorf("no DomainRequest found")
	}

	// Use the first DomainRequest's status
	dr := domainRequestList.Items[0]
	return dr.Status.Subdomain, dr.Status.Domain, nil
}

// getRouterIP gets the ClusterIP of the ingress controller service
func (h *HostsHandler) getRouterIP(ctx context.Context) (string, error) {
	routerSvc := &corev1.Service{}
	if err := h.k8sClient.Get(ctx, client.ObjectKey{
		Namespace: h.namespace,
		Name:      h.routerServiceRef,
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
func (h *HostsHandler) getIngressHosts(ctx context.Context, routerIP string) ([]HostEntry, error) {
	var ingressList networkingv1.IngressList
	// List all ingresses cluster-wide
	if err := h.k8sClient.List(ctx, &ingressList); err != nil {
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
func (h *HostsHandler) getServiceHosts(ctx context.Context, subdomain, domain string) ([]HostEntry, error) {
	// List namespaces with kloudlite environment label
	var namespaceList corev1.NamespaceList
	if err := h.k8sClient.List(ctx, &namespaceList, client.HasLabels{kloudliteEnvironmentLabel}); err != nil {
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

		// Get owner from annotation
		owner := ns.Annotations[kloudliteCreatedByAnnotation]
		if owner == "" {
			owner = "unknown"
		}

		// List services in this namespace
		var serviceList corev1.ServiceList
		if err := h.k8sClient.List(ctx, &serviceList, client.InNamespace(ns.Name)); err != nil {
			h.logger.Warn("failed to list services in namespace",
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

			// Build hostname: {service}-{envName}-{owner}.{subdomain}.{domain}
			hostname := fmt.Sprintf("%s-%s-%s.%s.%s", svc.Name, envName, owner, subdomain, domain)

			hosts = append(hosts, HostEntry{
				Hostname: hostname,
				IP:       svc.Spec.ClusterIP,
				Type:     "service",
			})
		}
	}

	return hosts, nil
}

// GetHostsForNamespace returns hosts for a specific namespace (utility method)
func (h *HostsHandler) GetHostsForNamespace(ctx context.Context, namespace string) ([]HostEntry, error) {
	var ingressList networkingv1.IngressList
	if err := h.k8sClient.List(ctx, &ingressList, client.InNamespace(namespace)); err != nil {
		return nil, err
	}

	routerSvc := &corev1.Service{}
	if err := h.k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      h.routerServiceRef,
	}, routerSvc); err != nil {
		return nil, err
	}

	clusterIP := routerSvc.Spec.ClusterIP
	hosts := make([]HostEntry, 0)
	for i := range ingressList.Items {
		for j := range ingressList.Items[i].Spec.Rules {
			host := ingressList.Items[i].Spec.Rules[j].Host
			if host != "" {
				hosts = append(hosts, HostEntry{
					Hostname: host,
					IP:       clusterIP,
					Type:     "ingress",
				})
			}
		}
	}

	return hosts, nil
}
