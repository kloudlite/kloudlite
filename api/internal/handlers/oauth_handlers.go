package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type OAuthProvider struct {
	Type         string `json:"type"`
	Enabled      bool   `json:"enabled"`
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
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
		if errors.IsNotFound(err) {
			// Create default empty providers
			providers := map[string]OAuthProvider{
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Parse providers from ConfigMap
	providers := make(map[string]OAuthProvider)
	for providerType, data := range configMap.Data {
		var provider OAuthProvider
		if err := json.Unmarshal([]byte(data), &provider); err != nil {
			continue
		}
		providers[providerType] = provider
	}

	// Ensure all provider types exist
	for _, providerType := range []string{"google", "github", "microsoft"} {
		if _, exists := providers[providerType]; !exists {
			providers[providerType] = OAuthProvider{
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
		if errors.IsNotFound(err) {
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
		if errors.IsNotFound(err) {
			// Create
			if err := h.k8sClient.Create(c.Request.Context(), configMap); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Update
		if err := h.k8sClient.Update(c.Request.Context(), configMap); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		if errors.IsNotFound(err) {
			// Return default disabled providers
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
