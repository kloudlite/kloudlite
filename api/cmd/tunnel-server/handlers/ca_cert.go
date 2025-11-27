package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"go.uber.org/zap"
)

// CACertHandler handles CA certificate requests
type CACertHandler struct {
	logger   *zap.Logger
	certPath string
}

// CACertHandlerConfig holds configuration for the CA cert handler
type CACertHandlerConfig struct {
	CertPath string // Path to CA certificate file (defaults to /certs/ca.crt)
}

// NewCACertHandler creates a new CACertHandler
func NewCACertHandler(logger *zap.Logger, cfg CACertHandlerConfig) *CACertHandler {
	if cfg.CertPath == "" {
		cfg.CertPath = "/certs/ca.crt"
	}
	return &CACertHandler{
		logger:   logger,
		certPath: cfg.CertPath,
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

	// Read CA certificate from file
	certData, err := os.ReadFile(h.certPath)
	if err != nil {
		h.logger.Error("failed to read CA certificate",
			zap.String("path", h.certPath),
			zap.Error(err))
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
