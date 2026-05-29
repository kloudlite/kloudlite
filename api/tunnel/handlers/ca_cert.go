package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CACertHandler handles CA certificate requests
type CACertHandler struct {
	logger     *zap.Logger
	k8sClient  client.Client
	namespace  string
	secretName string
}

// CACertHandlerConfig holds configuration for the CA cert handler
type CACertHandlerConfig struct {
	K8sClient  client.Client
	Namespace  string
	SecretName string
}

// NewCACertHandler creates a new CACertHandler
func NewCACertHandler(logger *zap.Logger, cfg CACertHandlerConfig) *CACertHandler {
	return &CACertHandler{
		logger:     logger,
		k8sClient:  cfg.K8sClient,
		namespace:  cfg.Namespace,
		secretName: cfg.SecretName,
	}
}

// CACertResponse represents the CA certificate response
type CACertResponse struct {
	CACert string `json:"ca_cert"`
}

// ServeHTTP implements http.Handler for CA certificate endpoint
func (h *CACertHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read CA certificate from Kubernetes secret
	secret := &corev1.Secret{}
	if err := h.k8sClient.Get(context.Background(), client.ObjectKey{
		Namespace: h.namespace,
		Name:      h.secretName,
	}, secret); err != nil {
		h.logger.Error("failed to get CA certificate secret",
			zap.String("namespace", h.namespace),
			zap.String("secret", h.secretName),
			zap.Error(err))
		http.Error(w, "CA certificate not available", http.StatusInternalServerError)
		return
	}

	certData, ok := secret.Data["ca.crt"]
	if !ok {
		h.logger.Error("CA certificate secret missing ca.crt",
			zap.String("namespace", h.namespace),
			zap.String("secret", h.secretName))
		http.Error(w, "CA certificate not available", http.StatusInternalServerError)
		return
	}

	response := CACertResponse{
		CACert: string(certData),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode CA cert response", zap.Error(err))
	}
}
