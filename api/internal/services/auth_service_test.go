package services

import (
	"context"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// mockUserServiceForAuth implements minimal UserService for auth testing
type mockUserServiceForAuth struct {
	getUserByEmailFunc func(ctx context.Context, email string) (*platformv1alpha1.User, error)
}

func (m *mockUserServiceForAuth) GetUserByEmail(ctx context.Context, email string) (*platformv1alpha1.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, errors.New("not implemented")
}

// Stub implementations for other UserService methods
func (m *mockUserServiceForAuth) CreateUser(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserServiceForAuth) GetUser(ctx context.Context, name string) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserServiceForAuth) GetUserByUsername(ctx context.Context, username string) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserServiceForAuth) UpdateUser(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserServiceForAuth) DeleteUser(ctx context.Context, name string) error {
	return errors.New("not implemented")
}
func (m *mockUserServiceForAuth) ListUsers(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserServiceForAuth) ActivateUser(ctx context.Context, name string) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserServiceForAuth) DeactivateUser(ctx context.Context, name string) (*platformv1alpha1.User, error) {
	return nil, errors.New("not implemented")
}
func (m *mockUserServiceForAuth) ResetUserPassword(ctx context.Context, name, newPassword string) error {
	return errors.New("not implemented")
}
func (m *mockUserServiceForAuth) UpdateUserLastLogin(ctx context.Context, name string) error {
	return nil
}
func (m *mockUserServiceForAuth) ValidatePassword(ctx context.Context, user *platformv1alpha1.User, password string) error {
	return errors.New("not implemented")
}
func (m *mockUserServiceForAuth) HashPassword(password string) (string, error) {
	return "", errors.New("not implemented")
}

// Helper function to generate a test JWT token for validation tests
func generateTestToken(secret string, email string, roles []platformv1alpha1.RoleType, expiry time.Duration) (string, error) {
	claims := &UserClaims{
		Email: email,
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   email,
			Issuer:    "kloudlite-api",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func TestValidateToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	jwtSecret := "test-secret-key"

	userService := &mockUserServiceForAuth{}
	authService := NewAuthService(jwtSecret, userService, logger)

	t.Run("should validate valid token", func(t *testing.T) {
		ctx := context.Background()
		email := "test@example.com"
		roles := []platformv1alpha1.RoleType{platformv1alpha1.RoleUser}

		// Generate token using helper
		tokenString, err := generateTestToken(jwtSecret, email, roles, 1*time.Hour)
		assert.NoError(t, err)

		// Validate token
		claims, err := authService.ValidateToken(ctx, tokenString)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, roles, claims.Roles)
	})

	t.Run("should reject expired token", func(t *testing.T) {
		ctx := context.Background()
		email := "test@example.com"
		roles := []platformv1alpha1.RoleType{platformv1alpha1.RoleUser}

		// Generate token with very short expiry (already expired)
		tokenString, err := generateTestToken(jwtSecret, email, roles, -1*time.Hour)
		assert.NoError(t, err)

		// Try to validate expired token
		claims, err := authService.ValidateToken(ctx, tokenString)
		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Contains(t, err.Error(), "invalid token")
	})

	t.Run("should reject token signed with wrong secret", func(t *testing.T) {
		ctx := context.Background()
		email := "test@example.com"
		roles := []platformv1alpha1.RoleType{platformv1alpha1.RoleUser}

		// Generate token with different secret
		tokenString, err := generateTestToken("wrong-secret", email, roles, 1*time.Hour)
		assert.NoError(t, err)

		// Try to validate with correct secret
		claims, err := authService.ValidateToken(ctx, tokenString)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("should reject malformed token", func(t *testing.T) {
		ctx := context.Background()

		claims, err := authService.ValidateToken(ctx, "invalid.token.string")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("should reject empty token", func(t *testing.T) {
		ctx := context.Background()

		claims, err := authService.ValidateToken(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("should reject token with invalid format", func(t *testing.T) {
		ctx := context.Background()

		claims, err := authService.ValidateToken(ctx, "not-a-jwt-token")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})

	t.Run("should reject tampered token", func(t *testing.T) {
		ctx := context.Background()
		email := "test@example.com"
		roles := []platformv1alpha1.RoleType{platformv1alpha1.RoleUser}

		tokenString, err := generateTestToken(jwtSecret, email, roles, 1*time.Hour)
		assert.NoError(t, err)

		// Tamper with the token
		tamperedToken := tokenString + "tampered"

		claims, err := authService.ValidateToken(ctx, tamperedToken)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
}

func TestVerifyPassword(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	jwtSecret := "test-secret-key"
	activeTrue := true
	activeFalse := false

	t.Run("should verify valid credentials", func(t *testing.T) {
		ctx := context.Background()
		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		userService := &mockUserServiceForAuth{
			getUserByEmailFunc: func(ctx context.Context, e string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
					Spec: platformv1alpha1.UserSpec{
						Email:    email,
						Password: base64.StdEncoding.EncodeToString(hashedPassword),
						Active:   &activeTrue,
						Roles:    []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
		}

		authService := NewAuthService(jwtSecret, userService, logger)
		user, err := authService.VerifyPassword(ctx, email, password)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, email, user.Spec.Email)
	})

	t.Run("should reject wrong password", func(t *testing.T) {
		ctx := context.Background()
		email := "test@example.com"
		correctPassword := "password123"
		wrongPassword := "wrongpassword"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

		userService := &mockUserServiceForAuth{
			getUserByEmailFunc: func(ctx context.Context, e string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
					Spec: platformv1alpha1.UserSpec{
						Email:    email,
						Password: base64.StdEncoding.EncodeToString(hashedPassword),
						Active:   &activeTrue,
						Roles:    []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
		}

		authService := NewAuthService(jwtSecret, userService, logger)
		user, err := authService.VerifyPassword(ctx, email, wrongPassword)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "authentication failed")
	})

	t.Run("should reject non-existent user", func(t *testing.T) {
		ctx := context.Background()
		email := "nonexistent@example.com"
		password := "password123"

		userService := &mockUserServiceForAuth{
			getUserByEmailFunc: func(ctx context.Context, e string) (*platformv1alpha1.User, error) {
				return nil, errors.New("user not found")
			},
		}

		authService := NewAuthService(jwtSecret, userService, logger)
		user, err := authService.VerifyPassword(ctx, email, password)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "authentication failed")
	})

	t.Run("should reject user with no password set", func(t *testing.T) {
		ctx := context.Background()
		email := "nopassword@example.com"
		password := "password123"

		userService := &mockUserServiceForAuth{
			getUserByEmailFunc: func(ctx context.Context, e string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: "oauth-user"},
					Spec: platformv1alpha1.UserSpec{
						Email:    email,
						Password: "", // No password set (OAuth user)
						Active:   &activeTrue,
						Roles:    []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
		}

		authService := NewAuthService(jwtSecret, userService, logger)
		user, err := authService.VerifyPassword(ctx, email, password)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "no password set")
	})

	t.Run("should reject inactive user", func(t *testing.T) {
		ctx := context.Background()
		email := "inactive@example.com"
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		userService := &mockUserServiceForAuth{
			getUserByEmailFunc: func(ctx context.Context, e string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: "inactive-user"},
					Spec: platformv1alpha1.UserSpec{
						Email:    email,
						Password: base64.StdEncoding.EncodeToString(hashedPassword),
						Active:   &activeFalse,
						Roles:    []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
		}

		authService := NewAuthService(jwtSecret, userService, logger)
		user, err := authService.VerifyPassword(ctx, email, password)
		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Contains(t, err.Error(), "inactive")
	})

	t.Run("should accept inactive user with nil Active field", func(t *testing.T) {
		ctx := context.Background()
		email := "test@example.com"
		password := "password123"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		userService := &mockUserServiceForAuth{
			getUserByEmailFunc: func(ctx context.Context, e string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
					Spec: platformv1alpha1.UserSpec{
						Email:    email,
						Password: base64.StdEncoding.EncodeToString(hashedPassword),
						Active:   nil, // nil means active by default
						Roles:    []platformv1alpha1.RoleType{platformv1alpha1.RoleUser},
					},
				}, nil
			},
		}

		authService := NewAuthService(jwtSecret, userService, logger)
		user, err := authService.VerifyPassword(ctx, email, password)
		assert.NoError(t, err)
		assert.NotNil(t, user)
	})

	t.Run("should verify password for user with multiple roles", func(t *testing.T) {
		ctx := context.Background()
		email := "admin@example.com"
		password := "adminpass"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

		userService := &mockUserServiceForAuth{
			getUserByEmailFunc: func(ctx context.Context, e string) (*platformv1alpha1.User, error) {
				return &platformv1alpha1.User{
					ObjectMeta: metav1.ObjectMeta{Name: "admin-user"},
					Spec: platformv1alpha1.UserSpec{
						Email:    email,
						Password: base64.StdEncoding.EncodeToString(hashedPassword),
						Active:   &activeTrue,
						Roles: []platformv1alpha1.RoleType{
							platformv1alpha1.RoleAdmin,
							platformv1alpha1.RoleUser,
						},
					},
				}, nil
			},
		}

		authService := NewAuthService(jwtSecret, userService, logger)
		user, err := authService.VerifyPassword(ctx, email, password)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, 2, len(user.Spec.Roles))
	})
}

func TestRolesToStrings(t *testing.T) {
	t.Run("should convert roles to strings", func(t *testing.T) {
		roles := []platformv1alpha1.RoleType{
			platformv1alpha1.RoleUser,
			platformv1alpha1.RoleAdmin,
			platformv1alpha1.RoleSuperAdmin,
		}

		result := rolesToStrings(roles)
		assert.Equal(t, 3, len(result))
		assert.Equal(t, "user", result[0])
		assert.Equal(t, "admin", result[1])
		assert.Equal(t, "super-admin", result[2])
	})

	t.Run("should handle empty roles", func(t *testing.T) {
		roles := []platformv1alpha1.RoleType{}
		result := rolesToStrings(roles)
		assert.Equal(t, 0, len(result))
	})

	t.Run("should handle single role", func(t *testing.T) {
		roles := []platformv1alpha1.RoleType{platformv1alpha1.RoleUser}
		result := rolesToStrings(roles)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, "user", result[0])
	})
}
