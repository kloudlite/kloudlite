package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// OAuthProvider is the internal representation with secrets (never exposed via API)
type OAuthProvider struct {
	Type         string `json:"type"`
	Enabled      bool   `json:"enabled"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"` // NEVER return this in API responses
}

// OAuthProviderResponse is the safe API response without secrets
type OAuthProviderResponse struct {
	Type     string `json:"type"`
	Enabled  bool   `json:"enabled"`
	ClientID string `json:"clientId"` // Safe to expose
	// ClientSecret is intentionally omitted for security
}

type OAuthHandlers struct {
	k8sClient client.Client
	namespace string
}

func NewOAuthHandlers(k8sClient client.Client, namespace string) *OAuthHandlers {
	return &OAuthHandlers{
		k8sClient: k8sClient,
		namespace: namespace,
	}
}

const oauthConfigMapName = "oauth-providers-config"

type PublicOAuthProvider struct {
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

// GetOAuthProviders retrieves all OAuth provider configurations
func (h *OAuthHandlers) GetOAuthProviders(c *gin.Context) {
	// JWT middleware already handles authentication, user context is available

	// Get or create the ConfigMap
	configMap := &corev1.ConfigMap{}
	err := h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      oauthConfigMapName,
		Namespace: h.namespace,
	}, configMap)
	if err != nil {
		// Handle TLS/certificate errors gracefully for development environments
		if isTLSError(err) {
			fmt.Printf("TLS error when getting OAuth providers (development mode): %v\n", err)
		} else if errors.IsNotFound(err) {
			// Create default empty providers (without secrets)
			providers := map[string]OAuthProviderResponse{
				"google": {
					Type:    "google",
					Enabled: false,
				},
				"github": {
					Type:    "github",
					Enabled: false,
				},
				"microsoft": {
					Type:    "microsoft",
					Enabled: false,
				},
			}
			c.JSON(http.StatusOK, providers)
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Parse providers from ConfigMap and convert to safe response type
	providers := make(map[string]OAuthProviderResponse)
	for providerType, data := range configMap.Data {
		var provider OAuthProvider
		if err := json.Unmarshal([]byte(data), &provider); err != nil {
			continue
		}
		// Convert to response type WITHOUT secrets
		providers[providerType] = OAuthProviderResponse{
			Type:     provider.Type,
			Enabled:  provider.Enabled,
			ClientID: provider.ClientID,
			// ClientSecret is intentionally excluded
		}
	}

	// Ensure all provider types exist
	for _, providerType := range []string{"google", "github", "microsoft"} {
		if _, exists := providers[providerType]; !exists {
			providers[providerType] = OAuthProviderResponse{
				Type:    providerType,
				Enabled: false,
			}
		}
	}

	c.JSON(http.StatusOK, providers)
}

// UpdateOAuthProvider updates a single OAuth provider configuration
func (h *OAuthHandlers) UpdateOAuthProvider(c *gin.Context) {
	providerType := c.Param("type")

	// JWT middleware already handles authentication, user context is available

	// Validate provider type
	validTypes := []string{"google", "github", "microsoft"}
	isValid := false
	for _, t := range validTypes {
		if providerType == t {
			isValid = true
			break
		}
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider type"})
		return
	}

	var provider OAuthProvider
	if err := c.ShouldBindJSON(&provider); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	provider.Type = providerType

	// Get or create ConfigMap
	configMap := &corev1.ConfigMap{}
	err := h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      oauthConfigMapName,
		Namespace: h.namespace,
	}, configMap)
	if err != nil {
		if isTLSError(err) {
			fmt.Printf("TLS error when getting OAuth providers config for update (development mode): %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "TLS certificate error in development mode"})
			return
		} else if errors.IsNotFound(err) {
			// Create new ConfigMap
			configMap = &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      oauthConfigMapName,
					Namespace: h.namespace,
				},
				Data: make(map[string]string),
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Update provider data
	if configMap.Data == nil {
		configMap.Data = make(map[string]string)
	}

	providerData, err := json.Marshal(provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	configMap.Data[providerType] = string(providerData)

	// Save ConfigMap
	if err := h.k8sClient.Get(c.Request.Context(), client.ObjectKey{Name: oauthConfigMapName, Namespace: h.namespace}, &corev1.ConfigMap{}); err != nil {
		if isTLSError(err) {
			fmt.Printf("TLS error when checking OAuth providers config existence (development mode): %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "TLS certificate error in development mode"})
			return
		} else if errors.IsNotFound(err) {
			// Create
			if err := h.k8sClient.Create(c.Request.Context(), configMap); err != nil {
				if isTLSError(err) {
					fmt.Printf("TLS error when creating OAuth providers config (development mode): %v\n", err)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "TLS certificate error in development mode"})
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				}
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Update
		if err := h.k8sClient.Update(c.Request.Context(), configMap); err != nil {
			if isTLSError(err) {
				fmt.Printf("TLS error when updating OAuth providers config (development mode): %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "TLS certificate error in development mode"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetPublicOAuthProviders retrieves enabled OAuth providers without secrets (public endpoint)
func (h *OAuthHandlers) GetPublicOAuthProviders(c *gin.Context) {
	// Get the ConfigMap
	configMap := &corev1.ConfigMap{}
	err := h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      oauthConfigMapName,
		Namespace: h.namespace,
	}, configMap)
	if err != nil {
		// Handle TLS/certificate errors gracefully for development environments
		if isTLSError(err) {
			// Log at debug level for development environments
			fmt.Printf("TLS error when getting OAuth providers (development mode): %v\n", err)
		} else {
			// Log actual errors at error level
			fmt.Printf("Error getting OAuth providers ConfigMap: %v\n", err)
		}

		// Return default disabled providers to ensure signin page works
		providers := map[string]PublicOAuthProvider{
			"google": {
				Type:    "google",
				Enabled: false,
			},
			"github": {
				Type:    "github",
				Enabled: false,
			},
			"microsoft": {
				Type:    "microsoft",
				Enabled: false,
			},
		}
		c.JSON(http.StatusOK, providers)
		return
	}

	// Parse providers from ConfigMap (without secrets)
	providers := make(map[string]PublicOAuthProvider)
	for providerType, data := range configMap.Data {
		var provider OAuthProvider
		if err := json.Unmarshal([]byte(data), &provider); err != nil {
			continue
		}
		// Only return type and enabled status (no secrets)
		providers[providerType] = PublicOAuthProvider{
			Type:    provider.Type,
			Enabled: provider.Enabled && provider.ClientID != "",
		}
	}

	// Ensure all provider types exist
	for _, providerType := range []string{"google", "github", "microsoft"} {
		if _, exists := providers[providerType]; !exists {
			providers[providerType] = PublicOAuthProvider{
				Type:    providerType,
				Enabled: false,
			}
		}
	}

	c.JSON(http.StatusOK, providers)
}

// isTLSError checks if the error is related to TLS certificate verification
func isTLSError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())
	tlsErrorStrings := []string{
		"tls: failed to verify certificate",
		"x509: certificate signed by unknown authority",
		"certificate not trusted",
		"certificate has expired",
		"certificate is not yet valid",
		"tls handshake error",
		"certificate authority",
	}

	for _, tlsStr := range tlsErrorStrings {
		if strings.Contains(errStr, tlsStr) {
			return true
		}
	}

	return false
}
