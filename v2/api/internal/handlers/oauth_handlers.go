package handlers

import (
	"context"
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

// GetOAuthProviders retrieves all OAuth provider configurations
func (h *OAuthHandlers) GetOAuthProviders(c *gin.Context) {
	// Check authentication
	userEmail := c.GetHeader("X-User-Email")
	if userEmail == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Get or create the ConfigMap
	configMap := &corev1.ConfigMap{}
	err := h.k8sClient.Get(context.Background(), client.ObjectKey{
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

	// Check authentication
	userEmail := c.GetHeader("X-User-Email")
	if userEmail == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

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
	err := h.k8sClient.Get(context.Background(), client.ObjectKey{
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
	if err := h.k8sClient.Get(context.Background(), client.ObjectKey{Name: oauthConfigMapName, Namespace: h.namespace}, &corev1.ConfigMap{}); err != nil {
		if errors.IsNotFound(err) {
			// Create
			if err := h.k8sClient.Create(context.Background(), configMap); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Update
		if err := h.k8sClient.Update(context.Background(), configMap); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}