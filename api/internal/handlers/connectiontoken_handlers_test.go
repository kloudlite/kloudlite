package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	connectiontokenv1 "github.com/kloudlite/kloudlite/api/internal/controllers/connectiontoken/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	testJWTSecret   = "test-secret-key-for-testing"
	testSSHJumpHost = "test.kloudlite.io"
	testSSHPort     = 2222
	testAPIURL      = "https://api.test.kloudlite.io"
	testUserEmail   = "test@example.com"
)

func setupConnectionTokenHandlerTest() (*ConnectionTokenHandlers, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	scheme := runtime.NewScheme()
	_ = connectiontokenv1.AddToScheme(scheme)

	k8sClient := fake.NewClientBuilder().WithScheme(scheme).Build()
	logger, _ := zap.NewDevelopment()

	handlers := NewConnectionTokenHandlers(k8sClient, logger, testJWTSecret, testSSHJumpHost, testAPIURL, testSSHPort)
	router := gin.New()

	return handlers, router
}

// addUserContext adds test user to gin context (simulating JWT middleware)
func addUserContext(c *gin.Context, email string) {
	c.Set("user_email", email)
	c.Set("user_roles", []string{"user"})
}

func TestCreateConnectionToken(t *testing.T) {
	t.Run("should create connection token successfully", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		router.POST("/connection-tokens", func(c *gin.Context) {
			addUserContext(c, testUserEmail)
			handlers.CreateConnectionToken(c)
		})

		reqBody := CreateConnectionTokenRequest{
			DisplayName: "Test Token",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/connection-tokens", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ConnectionTokenResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify token was created
		assert.NotNil(t, response.Token)
		assert.Equal(t, "Test Token", response.Token.Spec.DisplayName)
		assert.Equal(t, testUserEmail, response.Token.Spec.UserID)
		assert.Equal(t, testSSHJumpHost, response.Token.Spec.SSHJumpHost)
		assert.Equal(t, testSSHPort, response.Token.Spec.SSHPort)
		assert.Equal(t, testAPIURL, response.Token.Spec.APIURL)

		// Verify JWT was generated
		assert.NotEmpty(t, response.JWT)

		// Verify JWT claims
		token, err := jwt.ParseWithClaims(response.JWT, &ConnectionTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(testJWTSecret), nil
		})
		require.NoError(t, err)
		require.True(t, token.Valid)

		claims, ok := token.Claims.(*ConnectionTokenClaims)
		require.True(t, ok)
		assert.Equal(t, testUserEmail, claims.Email)
		assert.Equal(t, testSSHJumpHost, claims.SSHJumpHost)
		assert.Equal(t, testSSHPort, claims.SSHPort)
		assert.Equal(t, testAPIURL, claims.APIURL)
	})

	t.Run("should create connection token with custom webUrl", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		router.POST("/connection-tokens", func(c *gin.Context) {
			addUserContext(c, testUserEmail)
			handlers.CreateConnectionToken(c)
		})

		customWebURL := "https://custom.kloudlite.io"
		reqBody := CreateConnectionTokenRequest{
			DisplayName: "Custom URL Token",
			WebURL:      customWebURL,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/connection-tokens", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ConnectionTokenResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Verify custom webUrl was used
		assert.Equal(t, customWebURL, response.Token.Spec.APIURL)

		// Verify JWT has custom URL
		token, _ := jwt.ParseWithClaims(response.JWT, &ConnectionTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(testJWTSecret), nil
		})
		claims := token.Claims.(*ConnectionTokenClaims)
		assert.Equal(t, customWebURL, claims.APIURL)
	})

	t.Run("should return 400 when displayName is missing", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		router.POST("/connection-tokens", func(c *gin.Context) {
			addUserContext(c, testUserEmail)
			handlers.CreateConnectionToken(c)
		})

		reqBody := CreateConnectionTokenRequest{
			DisplayName: "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/connection-tokens", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 when displayName exceeds max length", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		router.POST("/connection-tokens", func(c *gin.Context) {
			addUserContext(c, testUserEmail)
			handlers.CreateConnectionToken(c)
		})

		// Create a displayName > 100 characters
		longName := ""
		for i := 0; i < 110; i++ {
			longName += "a"
		}

		reqBody := CreateConnectionTokenRequest{
			DisplayName: longName,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/connection-tokens", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 with invalid JSON", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		router.POST("/connection-tokens", func(c *gin.Context) {
			addUserContext(c, testUserEmail)
			handlers.CreateConnectionToken(c)
		})

		req := httptest.NewRequest(http.MethodPost, "/connection-tokens", bytes.NewBuffer([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 401 when user not in context", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		router.POST("/connection-tokens", handlers.CreateConnectionToken)

		reqBody := CreateConnectionTokenRequest{
			DisplayName: "Test Token",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/connection-tokens", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestListConnectionTokens(t *testing.T) {
	t.Run("should list connection tokens for current user", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		// Create test tokens for different users
		token1 := &connectiontokenv1.ConnectionToken{
			ObjectMeta: metav1.ObjectMeta{
				Name: "token-1",
			},
			Spec: connectiontokenv1.ConnectionTokenSpec{
				DisplayName: "Token 1",
				UserID:      testUserEmail,
				SSHJumpHost: testSSHJumpHost,
				SSHPort:     testSSHPort,
				APIURL:      testAPIURL,
			},
			Status: connectiontokenv1.ConnectionTokenStatus{
				IsReady: true,
				Token:   "secret-jwt-1", // Should be cleared
			},
		}
		token2 := &connectiontokenv1.ConnectionToken{
			ObjectMeta: metav1.ObjectMeta{
				Name: "token-2",
			},
			Spec: connectiontokenv1.ConnectionTokenSpec{
				DisplayName: "Token 2",
				UserID:      testUserEmail,
				SSHJumpHost: testSSHJumpHost,
				SSHPort:     testSSHPort,
				APIURL:      testAPIURL,
			},
			Status: connectiontokenv1.ConnectionTokenStatus{
				IsReady: true,
				Token:   "secret-jwt-2", // Should be cleared
			},
		}
		otherUserToken := &connectiontokenv1.ConnectionToken{
			ObjectMeta: metav1.ObjectMeta{
				Name: "other-token",
			},
			Spec: connectiontokenv1.ConnectionTokenSpec{
				DisplayName: "Other User Token",
				UserID:      "other@example.com",
				SSHJumpHost: testSSHJumpHost,
				SSHPort:     testSSHPort,
				APIURL:      testAPIURL,
			},
		}

		_ = handlers.k8sClient.Create(context.Background(), token1)
		_ = handlers.k8sClient.Create(context.Background(), token2)
		_ = handlers.k8sClient.Create(context.Background(), otherUserToken)

		router.GET("/connection-tokens", func(c *gin.Context) {
			addUserContext(c, testUserEmail)
			handlers.ListConnectionTokens(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/connection-tokens", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response connectiontokenv1.ConnectionTokenList
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should only return current user's tokens
		assert.Len(t, response.Items, 2)

		// Verify tokens belong to current user
		for _, token := range response.Items {
			assert.Equal(t, testUserEmail, token.Spec.UserID)
			// Verify JWT was cleared
			assert.Empty(t, token.Status.Token)
		}
	})

	t.Run("should return empty list when no tokens exist", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		router.GET("/connection-tokens", func(c *gin.Context) {
			addUserContext(c, testUserEmail)
			handlers.ListConnectionTokens(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/connection-tokens", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response connectiontokenv1.ConnectionTokenList
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Len(t, response.Items, 0)
	})

	t.Run("should return 401 when user not in context", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		router.GET("/connection-tokens", handlers.ListConnectionTokens)

		req := httptest.NewRequest(http.MethodGet, "/connection-tokens", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestDeleteConnectionToken(t *testing.T) {
	t.Run("should delete connection token successfully", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		// Create test token
		token := &connectiontokenv1.ConnectionToken{
			ObjectMeta: metav1.ObjectMeta{
				Name: "delete-token",
			},
			Spec: connectiontokenv1.ConnectionTokenSpec{
				DisplayName: "Delete Token",
				UserID:      testUserEmail,
				SSHJumpHost: testSSHJumpHost,
				SSHPort:     testSSHPort,
				APIURL:      testAPIURL,
			},
		}
		_ = handlers.k8sClient.Create(context.Background(), token)

		router.DELETE("/connection-tokens/:name", func(c *gin.Context) {
			addUserContext(c, testUserEmail)
			handlers.DeleteConnectionToken(c)
		})

		req := httptest.NewRequest(http.MethodDelete, "/connection-tokens/delete-token", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify token was deleted
		var deletedToken connectiontokenv1.ConnectionToken
		err := handlers.k8sClient.Get(context.Background(), getObjectKey("delete-token"), &deletedToken)
		assert.Error(t, err) // Should not exist
	})

	t.Run("should return 404 for non-existent token", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		router.DELETE("/connection-tokens/:name", func(c *gin.Context) {
			addUserContext(c, testUserEmail)
			handlers.DeleteConnectionToken(c)
		})

		req := httptest.NewRequest(http.MethodDelete, "/connection-tokens/nonexistent", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return 403 when trying to delete another user's token", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		// Create token owned by another user
		otherToken := &connectiontokenv1.ConnectionToken{
			ObjectMeta: metav1.ObjectMeta{
				Name: "other-user-token",
			},
			Spec: connectiontokenv1.ConnectionTokenSpec{
				DisplayName: "Other User Token",
				UserID:      "other@example.com",
				SSHJumpHost: testSSHJumpHost,
				SSHPort:     testSSHPort,
				APIURL:      testAPIURL,
			},
		}
		_ = handlers.k8sClient.Create(context.Background(), otherToken)

		router.DELETE("/connection-tokens/:name", func(c *gin.Context) {
			addUserContext(c, testUserEmail)
			handlers.DeleteConnectionToken(c)
		})

		req := httptest.NewRequest(http.MethodDelete, "/connection-tokens/other-user-token", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)

		// Verify token was NOT deleted
		var stillExists connectiontokenv1.ConnectionToken
		err := handlers.k8sClient.Get(context.Background(), getObjectKey("other-user-token"), &stillExists)
		assert.NoError(t, err) // Should still exist
	})

	t.Run("should return 401 when user not in context", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		router.DELETE("/connection-tokens/:name", handlers.DeleteConnectionToken)

		req := httptest.NewRequest(http.MethodDelete, "/connection-tokens/some-token", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestValidateConnectionToken(t *testing.T) {
	t.Run("should validate connection token successfully", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		// Create test connection token
		tokenName := "test-validation-token"
		connectionToken := &connectiontokenv1.ConnectionToken{
			ObjectMeta: metav1.ObjectMeta{
				Name: tokenName,
			},
			Spec: connectiontokenv1.ConnectionTokenSpec{
				DisplayName: "Validation Test Token",
				UserID:      testUserEmail,
				SSHJumpHost: testSSHJumpHost,
				SSHPort:     testSSHPort,
				APIURL:      testAPIURL,
			},
			Status: connectiontokenv1.ConnectionTokenStatus{
				IsReady: true,
			},
		}
		_ = handlers.k8sClient.Create(context.Background(), connectionToken)

		// Generate valid JWT
		claims := &ConnectionTokenClaims{
			Email:       testUserEmail,
			TokenID:     tokenName,
			SSHJumpHost: testSSHJumpHost,
			SSHPort:     testSSHPort,
			APIURL:      testAPIURL,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				Issuer:    "kloudlite-api",
				Subject:   testUserEmail,
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		jwtString, _ := token.SignedString([]byte(testJWTSecret))

		router.POST("/validate", handlers.ValidateConnectionToken)

		req := httptest.NewRequest(http.MethodPost, "/validate", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtString))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["valid"].(bool))
		assert.Equal(t, testUserEmail, response["email"])
		assert.Equal(t, tokenName, response["tokenId"])
		assert.Equal(t, testSSHJumpHost, response["sshJumpHost"])
		assert.Equal(t, float64(testSSHPort), response["sshPort"])
		assert.Equal(t, testAPIURL, response["apiUrl"])

		// Verify LastUsed was updated
		var updatedToken connectiontokenv1.ConnectionToken
		err = handlers.k8sClient.Get(context.Background(), getObjectKey(tokenName), &updatedToken)
		require.NoError(t, err)
		assert.NotNil(t, updatedToken.Status.LastUsed)
	})

	t.Run("should return 400 when Authorization header is missing", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		router.POST("/validate", handlers.ValidateConnectionToken)

		req := httptest.NewRequest(http.MethodPost, "/validate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 when Authorization header format is invalid", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		router.POST("/validate", handlers.ValidateConnectionToken)

		req := httptest.NewRequest(http.MethodPost, "/validate", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 401 when JWT is invalid", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		router.POST("/validate", handlers.ValidateConnectionToken)

		req := httptest.NewRequest(http.MethodPost, "/validate", nil)
		req.Header.Set("Authorization", "Bearer invalid.jwt.token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 401 when JWT is expired", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		// Generate expired JWT
		claims := &ConnectionTokenClaims{
			Email:       testUserEmail,
			TokenID:     "expired-token",
			SSHJumpHost: testSSHJumpHost,
			SSHPort:     testSSHPort,
			APIURL:      testAPIURL,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-24 * time.Hour)), // Expired
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-48 * time.Hour)),
				Issuer:    "kloudlite-api",
				Subject:   testUserEmail,
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		jwtString, _ := token.SignedString([]byte(testJWTSecret))

		router.POST("/validate", handlers.ValidateConnectionToken)

		req := httptest.NewRequest(http.MethodPost, "/validate", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtString))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return 401 when connection token has been revoked", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		// Generate JWT for non-existent token (revoked)
		claims := &ConnectionTokenClaims{
			Email:       testUserEmail,
			TokenID:     "revoked-token",
			SSHJumpHost: testSSHJumpHost,
			SSHPort:     testSSHPort,
			APIURL:      testAPIURL,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "kloudlite-api",
				Subject:   testUserEmail,
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		jwtString, _ := token.SignedString([]byte(testJWTSecret))

		router.POST("/validate", handlers.ValidateConnectionToken)

		req := httptest.NewRequest(http.MethodPost, "/validate", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtString))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Contains(t, response["error"], "revoked")
	})

	t.Run("should return 401 when JWT signing method is invalid", func(t *testing.T) {
		handlers, router := setupConnectionTokenHandlerTest()

		// Generate JWT with different signing method (RS256 instead of HS256)
		claims := &ConnectionTokenClaims{
			Email:   testUserEmail,
			TokenID: "test-token",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			},
		}

		// Create a JWT with wrong signing algorithm
		token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
		jwtString, _ := token.SignedString([]byte(testJWTSecret))

		router.POST("/validate", handlers.ValidateConnectionToken)

		req := httptest.NewRequest(http.MethodPost, "/validate", nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtString))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestGenerateTokenName(t *testing.T) {
	t.Run("should generate unique token names", func(t *testing.T) {
		name1, err := generateTokenName(testUserEmail)
		require.NoError(t, err)
		assert.NotEmpty(t, name1)
		assert.Contains(t, name1, "ct-")

		// Generate another and verify they're different
		time.Sleep(1 * time.Millisecond)
		name2, err := generateTokenName(testUserEmail)
		require.NoError(t, err)
		assert.NotEmpty(t, name2)
		assert.NotEqual(t, name1, name2)
	})

	t.Run("token name should have correct format", func(t *testing.T) {
		name, err := generateTokenName(testUserEmail)
		require.NoError(t, err)

		// Verify format: ct-{timestamp}-{random}
		assert.Regexp(t, `^ct-\d+-[0-9a-f]{16}$`, name)
	})
}

// Helper function to create object key
func getObjectKey(name string) types.NamespacedName {
	return types.NamespacedName{Name: name}
}
