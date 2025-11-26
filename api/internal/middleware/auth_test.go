package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"github.com/kloudlite/kloudlite/api/internal/services"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// mockAuthService implements services.AuthService for testing
type mockAuthService struct {
	validateFunc func(ctx context.Context, token string) (*services.UserClaims, error)
}

func (m *mockAuthService) ValidateToken(ctx context.Context, token string) (*services.UserClaims, error) {
	// Simple mock implementation
	if token == "valid-token" {
		return &services.UserClaims{
			Email: "test@example.com",
			Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			},
		}, nil
	}
	return nil, jwt.ErrSignatureInvalid
}

func (m *mockAuthService) VerifyPassword(ctx context.Context, email, password string) (*platformv1alpha1.User, error) {
	return nil, nil
}

func TestJWTMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("should reject request without Authorization header", func(t *testing.T) {
		authService := &mockAuthService{}
		middleware := JWTMiddleware(authService, logger, false)

		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header required")
	})

	t.Run("should reject request with invalid Authorization format", func(t *testing.T) {
		authService := &mockAuthService{}
		middleware := JWTMiddleware(authService, logger, false)

		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid authorization header format")
	})

	t.Run("should reject request with empty token", func(t *testing.T) {
		authService := &mockAuthService{}
		middleware := JWTMiddleware(authService, logger, false)

		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer ")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Empty token")
	})

	t.Run("should reject request with invalid token", func(t *testing.T) {
		authService := &mockAuthService{}
		middleware := JWTMiddleware(authService, logger, false)

		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")
	})

	t.Run("should accept request with valid token", func(t *testing.T) {
		authService := &mockAuthService{}
		middleware := JWTMiddleware(authService, logger, false)

		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			username, email, roles, exists := GetUserFromContext(c)
			assert.True(t, exists)
			assert.NotEmpty(t, username)
			assert.Equal(t, "test@example.com", email)
			assert.Equal(t, []platformv1alpha1.RoleType{platformv1alpha1.RoleUser}, roles)
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})

	t.Run("should skip authentication when skipAuth is true", func(t *testing.T) {
		authService := &mockAuthService{}
		middleware := JWTMiddleware(authService, logger, true)

		router := gin.New()
		router.Use(middleware)
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		// No Authorization header
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	})
}

func TestGetUserFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should return user data when present", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			c.Set("user_username", "testuser")
			c.Set("user_email", "test@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})

			username, email, roles, exists := GetUserFromContext(c)
			assert.True(t, exists)
			assert.Equal(t, "testuser", username)
			assert.Equal(t, "test@example.com", email)
			assert.Equal(t, []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin}, roles)
			c.JSON(200, gin.H{"ok": true})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should return false when user data not present", func(t *testing.T) {
		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			username, email, roles, exists := GetUserFromContext(c)
			assert.False(t, exists)
			assert.Empty(t, username)
			assert.Empty(t, email)
			assert.Nil(t, roles)
			c.JSON(200, gin.H{"ok": true})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestRequireRoles(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should allow user with required role", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})
		router.Use(RequireRoles(platformv1alpha1.RoleAdmin))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should reject user without required role", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_email", "user@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		})
		router.Use(RequireRoles(platformv1alpha1.RoleAdmin))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "Insufficient permissions")
	})

	t.Run("should reject when user not authenticated", func(t *testing.T) {
		router := gin.New()
		router.Use(RequireRoles(platformv1alpha1.RoleAdmin))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "User not authenticated")
	})
}

func TestRequireSuperAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should allow super admin", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_email", "superadmin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleSuperAdmin})
			c.Next()
		})
		router.Use(RequireSuperAdmin())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should reject regular admin", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})
		router.Use(RequireSuperAdmin())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRequireAdminOrSuperAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should allow admin", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_email", "admin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleAdmin})
			c.Next()
		})
		router.Use(RequireAdminOrSuperAdmin())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should allow super admin", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_email", "superadmin@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleSuperAdmin})
			c.Next()
		})
		router.Use(RequireAdminOrSuperAdmin())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("should reject regular user", func(t *testing.T) {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_email", "user@example.com")
			c.Set("user_roles", []platformv1alpha1.RoleType{platformv1alpha1.RoleUser})
			c.Next()
		})
		router.Use(RequireAdminOrSuperAdmin())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
