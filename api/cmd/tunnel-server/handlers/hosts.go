package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	Namespace        string // Namespace to query for ingresses (defaults to current namespace)
	RouterServiceRef string // Reference to the router service (format: namespace/name, defaults to current-ns/wm-ingress-controller)
}

// HostEntry represents a single hosts file entry
type HostEntry struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
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

	// Get all ingresses in the namespace
	var ingressList networkingv1.IngressList
	if err := h.k8sClient.List(ctx, &ingressList, client.InNamespace(h.namespace)); err != nil {
		h.logger.Error("failed to list ingresses",
			zap.String("namespace", h.namespace),
			zap.Error(err))
		http.Error(w, "Failed to get hosts configuration", http.StatusInternalServerError)
		return
	}

	// Get the router service to obtain ClusterIP
	routerSvc := &corev1.Service{}
	if err := h.k8sClient.Get(ctx, client.ObjectKey{
		Namespace: h.namespace,
		Name:      h.routerServiceRef,
	}, routerSvc); err != nil {
		h.logger.Error("failed to get router service",
			zap.String("namespace", h.namespace),
			zap.String("service", h.routerServiceRef),
			zap.Error(err))
		http.Error(w, "Router service not available", http.StatusInternalServerError)
		return
	}

	clusterIP := routerSvc.Spec.ClusterIP
	if clusterIP == "" || clusterIP == "None" {
		h.logger.Error("router service has no ClusterIP",
			zap.String("service", h.routerServiceRef))
		http.Error(w, "Router service has no ClusterIP", http.StatusInternalServerError)
		return
	}

	// Build host entries from ingresses
	hosts := make([]HostEntry, 0)
	for i := range ingressList.Items {
		for j := range ingressList.Items[i].Spec.Rules {
			host := ingressList.Items[i].Spec.Rules[j].Host
			if host != "" {
				hosts = append(hosts, HostEntry{
					Hostname: host,
					IP:       clusterIP,
				})
			}
		}
	}

	h.logger.Debug("returning hosts",
		zap.Int("count", len(hosts)),
		zap.String("namespace", h.namespace))

	response := HostsResponse{
		Hosts: hosts,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode hosts response", zap.Error(err))
	}
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
				})
			}
		}
	}

	return hosts, nil
}
