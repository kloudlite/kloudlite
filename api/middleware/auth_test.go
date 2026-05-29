package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kloudlite/kloudlite/api/config"
	"go.uber.org/zap"
)

func TestJWTAuthAllowsRequestWhenAuthenticationIsSkipped(t *testing.T) {
	router := newAuthTestRouter(config.AuthConfig{SkipAuthentication: true})

	response := performAuthRequest(router, "")

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
}

func TestJWTAuthRejectsMissingMalformedAndInvalidTokens(t *testing.T) {
	router := newAuthTestRouter(config.AuthConfig{JWTSecret: "secret"})

	tests := []struct {
		name   string
		header string
	}{
		{name: "missing"},
		{name: "malformed", header: "Basic token"},
		{name: "empty bearer", header: "Bearer "},
		{name: "invalid token", header: "Bearer not-a-jwt"},
		{name: "wrong secret", header: "Bearer " + signedAuthTestToken(t, "other-secret")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := performAuthRequest(router, tt.header)
			if response.Code != http.StatusUnauthorized {
				t.Fatalf("expected status %d, got %d: %s", http.StatusUnauthorized, response.Code, response.Body.String())
			}
		})
	}
}

func TestJWTAuthAllowsValidHMACToken(t *testing.T) {
	secret := "secret"
	router := newAuthTestRouter(config.AuthConfig{JWTSecret: secret})

	response := performAuthRequest(router, "Bearer "+signedAuthTestToken(t, secret))

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}
}

func newAuthTestRouter(auth config.AuthConfig) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(JWTAuth(auth, zap.NewNop()))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return router
}

func performAuthRequest(router http.Handler, authHeader string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	return response
}

func signedAuthTestToken(t *testing.T, secret string) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "test-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}
