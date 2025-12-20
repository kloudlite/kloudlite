package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// TLSCertHandler handles TLS certificate requests for kltun HTTPS server
type TLSCertHandler struct {
	logger     *zap.Logger
	k8sClient  client.Client
	namespace  string
	secretName string
}

// TLSCertHandlerConfig holds configuration for the TLS cert handler
type TLSCertHandlerConfig struct {
	K8sClient  client.Client
	Namespace  string
	SecretName string
}

// NewTLSCertHandler creates a new TLSCertHandler
func NewTLSCertHandler(logger *zap.Logger, cfg TLSCertHandlerConfig) *TLSCertHandler {
	return &TLSCertHandler{
		logger:     logger,
		k8sClient:  cfg.K8sClient,
		namespace:  cfg.Namespace,
		secretName: cfg.SecretName,
	}
}

// TLSCertResponse represents the TLS certificate response
type TLSCertResponse struct {
	TLSCert string `json:"tls_cert"` // Server certificate PEM
	TLSKey  string `json:"tls_key"`  // Server private key PEM
	CACert  string `json:"ca_cert"`  // CA certificate PEM
}

// ServeHTTP implements http.Handler for TLS certificate endpoint
func (h *TLSCertHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read TLS certificate from Kubernetes secret
	secret := &corev1.Secret{}
	if err := h.k8sClient.Get(context.Background(), client.ObjectKey{
		Namespace: h.namespace,
		Name:      h.secretName,
	}, secret); err != nil {
		h.logger.Error("failed to get TLS certificate secret",
			zap.String("namespace", h.namespace),
			zap.String("secret", h.secretName),
			zap.Error(err))
		http.Error(w, "TLS certificate not available", http.StatusInternalServerError)
		return
	}

	// Extract certificate components
	tlsCert, hasCert := secret.Data["tls.crt"]
	tlsKey, hasKey := secret.Data["tls.key"]
	caCert, hasCA := secret.Data["ca.crt"]

	if !hasCert {
		h.logger.Error("TLS certificate secret missing tls.crt",
			zap.String("namespace", h.namespace),
			zap.String("secret", h.secretName))
		http.Error(w, "TLS certificate not available", http.StatusInternalServerError)
		return
	}

	if !hasKey {
		h.logger.Error("TLS certificate secret missing tls.key",
			zap.String("namespace", h.namespace),
			zap.String("secret", h.secretName))
		http.Error(w, "TLS certificate not available", http.StatusInternalServerError)
		return
	}

	if !hasCA {
		h.logger.Error("TLS certificate secret missing ca.crt",
			zap.String("namespace", h.namespace),
			zap.String("secret", h.secretName))
		http.Error(w, "TLS certificate not available", http.StatusInternalServerError)
		return
	}

	response := TLSCertResponse{
		TLSCert: string(tlsCert),
		TLSKey:  string(tlsKey),
		CACert:  string(caCert),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode TLS cert response", zap.Error(err))
	}
}
