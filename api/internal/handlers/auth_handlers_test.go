package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"github.com/kloudlite/kloudlite/api/internal/services"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// mockAuthService implements services.AuthService for testing
type mockAuthService struct {
	verifyPasswordFunc func(ctx context.Context, email, password string) (*platformv1alpha1.User, error)
	validateTokenFunc  func(ctx context.Context, token string) (*services.UserClaims, error)
}

func (m *mockAuthService) VerifyPassword(ctx context.Context, email, password string) (*platformv1alpha1.User, error) {
	if m.verifyPasswordFunc != nil {
		return m.verifyPasswordFunc(ctx, email, password)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAuthService) ValidateToken(ctx context.Context, token string) (*services.UserClaims, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAuthService) GenerateToken(username string, email string, roles []platformv1alpha1.RoleType, expiryHours int) (string, error) {
	return "mock-token", nil
}

// mockUserService implements services.UserService for testing
type mockUserService struct {
	getUserByEmailFunc      func(ctx context.Context, email string) (*platformv1alpha1.User, error)
	updateUserLastLoginFunc func(ctx context.Context, name string) error
}

func (m *mockUserService) GetUserByEmail(ctx context.Context, email string) (*platformv1alpha1.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, errors.New("not implemented")
}

func (m *mockUserService) UpdateUserLastLogin(ctx context.Context, name string) error {
	if m.updateUserLastLoginFunc != nil {
		return m.updateUserLastLoginFunc(ctx, name)
	}
	return nil
}

// Stub implementations for other UserService methods
func (m *mockUserService) CreateUser(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserService) GetUser(ctx context.Context, name string) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserService) GetUserByUsername(ctx context.Context, username string) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserService) UpdateUser(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserService) DeleteUser(ctx context.Context, name string) error {
	return errors.New("not implemented")
}
func (m *mockUserService) ListUsers(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserService) ValidatePassword(ctx context.Context, user *platformv1alpha1.User, password string) error {
	return errors.New("not implemented")
}
func (m *mockUserService) HashPassword(password string) (string, error) {
	return "", errors.New("not implemented")
}
func (m *mockUserService) ActivateUser(ctx context.Context, name string) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserService) DeactivateUser(ctx context.Context, name string) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserService) ResetUserPassword(ctx context.Context, name, newPassword string) error {
	return errors.New("not implemented")
}

func TestLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	activeTrue := true

	t.Run("should login successfully with valid credentials", func(t *testing.T) {
		authService := &mockAuthService{
			verifyPasswordFunc: func(ctx context.Context, email, password string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
					Spec: platformv1alpha1.UserSpec{
						Email:       "test@example.com",
						DisplayName: "Test User",
						Active:      &activeTrue,
						Roles:       []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
		}
		userService := &mockUserService{
			updateUserLastLoginFunc: func(ctx context.Context, name string) error {
				return nil
			},
		}
		handlers := NewAuthHandlers(authService, userService, logger)

		router := gin.New()
		router.POST("/login", handlers.Login)

		loginReq := LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		// No token in response anymore - NextAuth will generate it
		assert.Equal(t, "test-user", response.User.Username)
		assert.Equal(t, "test@example.com", response.User.Email)
		assert.Equal(t, "Test User", response.User.Name)
		assert.Equal(t, "Test User", response.User.DisplayName)
		assert.True(t, response.User.IsActive)
		assert.Equal(t, []platformv1alpha1.RoleType{platformv1alpha1.RoleUser}, response.Roles)
	})

	t.Run("should reject login with invalid credentials", func(t *testing.T) {
		authService := &mockAuthService{
			verifyPasswordFunc: func(ctx context.Context, email, password string) (*platformv1alpha1.User, error) {
				return nil, errors.New("authentication failed: invalid password")
			},
		}
		userService := &mockUserService{}
		handlers := NewAuthHandlers(authService, userService, logger)

		router := gin.New()
		router.POST("/login", handlers.Login)

		loginReq := LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid credentials")
	})

	t.Run("should reject login with invalid request payload", func(t *testing.T) {
		authService := &mockAuthService{}
		userService := &mockUserService{}
		handlers := NewAuthHandlers(authService, userService, logger)

		router := gin.New()
		router.POST("/login", handlers.Login)

		// Missing required fields
		body := []byte(`{"email": "invalid-email"}`)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request payload")
	})

	t.Run("should reject login with malformed JSON", func(t *testing.T) {
		authService := &mockAuthService{}
		userService := &mockUserService{}
		handlers := NewAuthHandlers(authService, userService, logger)

		router := gin.New()
		router.POST("/login", handlers.Login)

		body := []byte(`{invalid json}`)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should succeed even if updating last login fails", func(t *testing.T) {
		authService := &mockAuthService{
			verifyPasswordFunc: func(ctx context.Context, email, password string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
					Spec: platformv1alpha1.UserSpec{
						Email:  "test@example.com",
						Active: &activeTrue,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
		}
		userService := &mockUserService{
			updateUserLastLoginFunc: func(ctx context.Context, name string) error {
				return errors.New("update failed")
			},
		}
		handlers := NewAuthHandlers(authService, userService, logger)

		router := gin.New()
		router.POST("/login", handlers.Login)

		loginReq := LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		body, _ := json.Marshal(loginReq)
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should still succeed - frontend handles token generation
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestGenerateToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()
	activeTrue := true
	activeFalse := false

	t.Run("should return user info for existing user (OAuth)", func(t *testing.T) {
		authService := &mockAuthService{}
		userService := &mockUserService{
			getUserByEmailFunc: func(ctx context.Context, email string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: "oauth-user"},
					Spec: platformv1alpha1.UserSpec{
						Email:       "oauth@example.com",
						DisplayName: "OAuth User",
						Active:      &activeTrue,
						Roles:       []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
			updateUserLastLoginFunc: func(ctx context.Context, name string) error {
				return nil
			},
		}
		handlers := NewAuthHandlers(authService, userService, logger)

		router := gin.New()
		router.POST("/token", handlers.GenerateToken)

		tokenReq := TokenRequest{
			Email: "oauth@example.com",
		}
		body, _ := json.Marshal(tokenReq)
		req, _ := http.NewRequest("POST", "/token", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response AuthResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		// No token in response anymore - NextAuth will generate it
		assert.Equal(t, "oauth-user", response.User.Username)
		assert.Equal(t, "oauth@example.com", response.User.Email)
		assert.Equal(t, "OAuth User", response.User.Name)
		assert.True(t, response.User.IsActive)
		assert.Equal(t, []platformv1alpha1.RoleType{platformv1alpha1.RoleUser}, response.Roles)
	})

	t.Run("should reject token generation for non-existent user", func(t *testing.T) {
		authService := &mockAuthService{}
		userService := &mockUserService{
			getUserByEmailFunc: func(ctx context.Context, email string) (*platformv1alpha1.User, error) {
				return nil, errors.New("user not found")
			},
		}
		handlers := NewAuthHandlers(authService, userService, logger)

		router := gin.New()
		router.POST("/token", handlers.GenerateToken)

		tokenReq := TokenRequest{
			Email: "nonexistent@example.com",
		}
		body, _ := json.Marshal(tokenReq)
		req, _ := http.NewRequest("POST", "/token", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Contains(t, w.Body.String(), "User not found")
	})

	t.Run("should reject token generation for inactive user", func(t *testing.T) {
		authService := &mockAuthService{}
		userService := &mockUserService{
			getUserByEmailFunc: func(ctx context.Context, email string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: "inactive-user"},
					Spec: platformv1alpha1.UserSpec{
						Email:  "inactive@example.com",
						Active: &activeFalse,
						Roles:  []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
		}
		handlers := NewAuthHandlers(authService, userService, logger)

		router := gin.New()
		router.POST("/token", handlers.GenerateToken)

		tokenReq := TokenRequest{
			Email: "inactive@example.com",
		}
		body, _ := json.Marshal(tokenReq)
		req, _ := http.NewRequest("POST", "/token", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "User account is inactive")
	})

	t.Run("should reject token generation with invalid request payload", func(t *testing.T) {
		authService := &mockAuthService{}
		userService := &mockUserService{}
		handlers := NewAuthHandlers(authService, userService, logger)

		router := gin.New()
		router.POST("/token", handlers.GenerateToken)

		// Invalid email format
		body := []byte(`{"email": "not-an-email"}`)
		req, _ := http.NewRequest("POST", "/token", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid request payload")
	})
}

func TestValidateToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger, _ := zap.NewDevelopment()

	t.Run("should validate valid token", func(t *testing.T) {
		authService := &mockAuthService{
			validateTokenFunc: func(ctx context.Context, token string) (*services.UserClaims, error) {
				return &services.UserClaims{
					Email: "test@example.com",
					Roles: []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
					},
				}, nil
			},
		}
		userService := &mockUserService{}
		handlers := NewAuthHandlers(authService, userService, logger)

		router := gin.New()
		router.POST("/validate", handlers.ValidateToken)

		req, _ := http.NewRequest("POST", "/validate", nil)
		req.Header.Set("Authorization", "Bearer valid-token-string")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "test@example.com")
		assert.Contains(t, w.Body.String(), "\"valid\":true")
	})

	t.Run("should reject invalid token", func(t *testing.T) {
		authService := &mockAuthService{
			validateTokenFunc: func(ctx context.Context, token string) (*services.UserClaims, error) {
				return nil, errors.New("invalid token")
			},
		}
		userService := &mockUserService{}
		handlers := NewAuthHandlers(authService, userService, logger)

		router := gin.New()
		router.POST("/validate", handlers.ValidateToken)

		req, _ := http.NewRequest("POST", "/validate", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid token")
	})

	t.Run("should reject request without Authorization header", func(t *testing.T) {
		authService := &mockAuthService{}
		userService := &mockUserService{}
		handlers := NewAuthHandlers(authService, userService, logger)

		router := gin.New()
		router.POST("/validate", handlers.ValidateToken)

		req, _ := http.NewRequest("POST", "/validate", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Authorization header required")
	})

	t.Run("should reject request with invalid Authorization header format", func(t *testing.T) {
		authService := &mockAuthService{}
		userService := &mockUserService{}
		handlers := NewAuthHandlers(authService, userService, logger)

		router := gin.New()
		router.POST("/validate", handlers.ValidateToken)

		req, _ := http.NewRequest("POST", "/validate", nil)
		req.Header.Set("Authorization", "InvalidFormat token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "Invalid authorization header format")
	})

	t.Run("should reject request with Bearer prefix only", func(t *testing.T) {
		authService := &mockAuthService{
			validateTokenFunc: func(ctx context.Context, token string) (*services.UserClaims, error) {
				return nil, errors.New("invalid token")
			},
		}
		userService := &mockUserService{}
		handlers := NewAuthHandlers(authService, userService, logger)

		router := gin.New()
		router.POST("/validate", handlers.ValidateToken)

		req, _ := http.NewRequest("POST", "/validate", nil)
		req.Header.Set("Authorization", "Bearer ")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Empty token gets passed to ValidateToken which returns 401
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
