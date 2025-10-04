package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func setupOAuthHandlerTest() (*OAuthHandlers, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()

	handlers := NewOAuthHandlers(k8sClient, "default")
	router := gin.New()

	return handlers, router
}

func TestGetOAuthProviders(t *testing.T) {
	handlers, router := setupOAuthHandlerTest()

	t.Run("should return default providers when ConfigMap does not exist", func(t *testing.T) {
		router.GET("/oauth/providers", handlers.GetOAuthProviders)

		req := httptest.NewRequest(http.MethodGet, "/oauth/providers", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]OAuthProvider
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Len(t, response, 3)
		assert.Contains(t, response, "google")
		assert.Contains(t, response, "github")
		assert.Contains(t, response, "microsoft")

		assert.Equal(t, "google", response["google"].Type)
		assert.False(t, response["google"].Enabled)
		assert.Equal(t, "github", response["github"].Type)
		assert.False(t, response["github"].Enabled)
		assert.Equal(t, "microsoft", response["microsoft"].Type)
		assert.False(t, response["microsoft"].Enabled)
	})

	t.Run("should return providers from ConfigMap", func(t *testing.T) {
		// Create ConfigMap with OAuth providers
		googleProvider := OAuthProvider{
			Type:         "google",
			Enabled:      true,
			ClientID:     "test-google-client-id",
			ClientSecret: "test-google-secret",
		}
		googleData, _ := json.Marshal(googleProvider)

		githubProvider := OAuthProvider{
			Type:         "github",
			Enabled:      true,
			ClientID:     "test-github-client-id",
			ClientSecret: "test-github-secret",
		}
		githubData, _ := json.Marshal(githubProvider)

		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "oauth-providers-config",
				Namespace: "default",
			},
			Data: map[string]string{
				"google": string(googleData),
				"github": string(githubData),
			},
		}
		_ = handlers.k8sClient.Create(context.Background(), configMap)

		router.GET("/oauth/providers-with-config", handlers.GetOAuthProviders)

		req := httptest.NewRequest(http.MethodGet, "/oauth/providers-with-config", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]OAuthProvider
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Len(t, response, 3) // google, github, microsoft (default added)

		// Check google provider
		assert.Equal(t, "google", response["google"].Type)
		assert.True(t, response["google"].Enabled)
		assert.Equal(t, "test-google-client-id", response["google"].ClientID)
		assert.Equal(t, "test-google-secret", response["google"].ClientSecret)

		// Check github provider
		assert.Equal(t, "github", response["github"].Type)
		assert.True(t, response["github"].Enabled)
		assert.Equal(t, "test-github-client-id", response["github"].ClientID)

		// Microsoft should be default (disabled)
		assert.Equal(t, "microsoft", response["microsoft"].Type)
		assert.False(t, response["microsoft"].Enabled)
	})
}

func TestUpdateOAuthProvider(t *testing.T) {
	handlers, router := setupOAuthHandlerTest()

	t.Run("should create new provider configuration", func(t *testing.T) {
		router.PUT("/oauth/providers/:type", handlers.UpdateOAuthProvider)

		provider := OAuthProvider{
			Enabled:      true,
			ClientID:     "new-client-id",
			ClientSecret: "new-client-secret",
		}
		body, _ := json.Marshal(provider)

		req := httptest.NewRequest(http.MethodPut, "/oauth/providers/google", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]bool
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.True(t, response["success"])

		// Verify ConfigMap was created
		configMap := &corev1.ConfigMap{}
		err = handlers.k8sClient.Get(context.Background(), client.ObjectKey{
			Name:      "oauth-providers-config",
			Namespace: "default",
		}, configMap)
		assert.NoError(t, err)
		assert.Contains(t, configMap.Data, "google")
	})

	t.Run("should update existing provider configuration", func(t *testing.T) {
		// Create initial ConfigMap
		googleProvider := OAuthProvider{
			Type:         "google",
			Enabled:      false,
			ClientID:     "old-client-id",
			ClientSecret: "old-secret",
		}
		googleData, _ := json.Marshal(googleProvider)

		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "oauth-providers-config",
				Namespace: "default",
			},
			Data: map[string]string{
				"google": string(googleData),
			},
		}
		_ = handlers.k8sClient.Create(context.Background(), configMap)

		router.PUT("/oauth/providers/:type/update", handlers.UpdateOAuthProvider)

		updatedProvider := OAuthProvider{
			Enabled:      true,
			ClientID:     "updated-client-id",
			ClientSecret: "updated-secret",
		}
		body, _ := json.Marshal(updatedProvider)

		req := httptest.NewRequest(http.MethodPut, "/oauth/providers/google/update", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify ConfigMap was updated
		updatedConfigMap := &corev1.ConfigMap{}
		err := handlers.k8sClient.Get(context.Background(), client.ObjectKey{
			Name:      "oauth-providers-config",
			Namespace: "default",
		}, updatedConfigMap)
		assert.NoError(t, err)

		var savedProvider OAuthProvider
		err = json.Unmarshal([]byte(updatedConfigMap.Data["google"]), &savedProvider)
		assert.NoError(t, err)
		assert.True(t, savedProvider.Enabled)
		assert.Equal(t, "updated-client-id", savedProvider.ClientID)
		assert.Equal(t, "updated-secret", savedProvider.ClientSecret)
	})

	t.Run("should return 400 for invalid provider type", func(t *testing.T) {
		router.PUT("/oauth/providers/:type/invalid", handlers.UpdateOAuthProvider)

		provider := OAuthProvider{
			Enabled:      true,
			ClientID:     "test-id",
			ClientSecret: "test-secret",
		}
		body, _ := json.Marshal(provider)

		req := httptest.NewRequest(http.MethodPut, "/oauth/providers/invalid-provider/invalid", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Invalid provider type", response["error"])
	})

	t.Run("should return 400 with invalid JSON", func(t *testing.T) {
		router.PUT("/oauth/providers/:type/bad-json", handlers.UpdateOAuthProvider)

		req := httptest.NewRequest(http.MethodPut, "/oauth/providers/google/bad-json", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetPublicOAuthProviders(t *testing.T) {
	handlers, router := setupOAuthHandlerTest()

	t.Run("should return default providers when ConfigMap does not exist", func(t *testing.T) {
		router.GET("/oauth/public", handlers.GetPublicOAuthProviders)

		req := httptest.NewRequest(http.MethodGet, "/oauth/public", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]PublicOAuthProvider
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Len(t, response, 3)
		assert.Contains(t, response, "google")
		assert.Contains(t, response, "github")
		assert.Contains(t, response, "microsoft")

		assert.Equal(t, "google", response["google"].Type)
		assert.False(t, response["google"].Enabled)
	})

	t.Run("should return public providers without secrets", func(t *testing.T) {
		// Create ConfigMap with OAuth providers
		googleProvider := OAuthProvider{
			Type:         "google",
			Enabled:      true,
			ClientID:     "test-google-client-id",
			ClientSecret: "test-google-secret",
		}
		googleData, _ := json.Marshal(googleProvider)

		githubProvider := OAuthProvider{
			Type:         "github",
			Enabled:      false,
			ClientID:     "test-github-client-id",
			ClientSecret: "test-github-secret",
		}
		githubData, _ := json.Marshal(githubProvider)

		// Provider with no ClientID should be disabled
		microsoftProvider := OAuthProvider{
			Type:         "microsoft",
			Enabled:      true,
			ClientID:     "",
			ClientSecret: "",
		}
		microsoftData, _ := json.Marshal(microsoftProvider)

		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "oauth-providers-config",
				Namespace: "default",
			},
			Data: map[string]string{
				"google":    string(googleData),
				"github":    string(githubData),
				"microsoft": string(microsoftData),
			},
		}
		_ = handlers.k8sClient.Create(context.Background(), configMap)

		router.GET("/oauth/public-with-config", handlers.GetPublicOAuthProviders)

		req := httptest.NewRequest(http.MethodGet, "/oauth/public-with-config", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]PublicOAuthProvider
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		assert.Len(t, response, 3)

		// Google should be enabled (has ClientID)
		assert.Equal(t, "google", response["google"].Type)
		assert.True(t, response["google"].Enabled)

		// GitHub should be disabled (Enabled=false)
		assert.Equal(t, "github", response["github"].Type)
		assert.False(t, response["github"].Enabled)

		// Microsoft should be disabled (no ClientID)
		assert.Equal(t, "microsoft", response["microsoft"].Type)
		assert.False(t, response["microsoft"].Enabled)

		// Verify no secrets are exposed
		responseJSON := w.Body.String()
		assert.NotContains(t, responseJSON, "ClientID")
		assert.NotContains(t, responseJSON, "ClientSecret")
		assert.NotContains(t, responseJSON, "test-google-client-id")
		assert.NotContains(t, responseJSON, "test-google-secret")
	})
}
