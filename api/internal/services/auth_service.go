package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AuthService handles JWT token generation and validation
type AuthService interface {
	GenerateToken(ctx context.Context, userEmail string, roles []platformv1alpha1.RoleType) (string, error)
	ValidateToken(ctx context.Context, tokenString string) (*UserClaims, error)
	VerifyPassword(ctx context.Context, email, password string) (*platformv1alpha1.User, error)
}

// UserClaims represents the JWT claims for a user
type UserClaims struct {
	Email string                      `json:"email"`
	Roles []platformv1alpha1.RoleType `json:"roles"`
	jwt.RegisteredClaims
}

// authService implements AuthService
type authService struct {
	jwtSecret   string
	tokenExpiry time.Duration
	userService UserService
	logger      *zap.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(jwtSecret string, tokenExpiry time.Duration, userService UserService, logger *zap.Logger) AuthService {
	return &authService{
		jwtSecret:   jwtSecret,
		tokenExpiry: tokenExpiry,
		userService: userService,
		logger:      logger,
	}
}

// GenerateToken creates a new JWT token for the user
func (s *authService) GenerateToken(ctx context.Context, userEmail string, roles []platformv1alpha1.RoleType) (string, error) {
	now := time.Now()
	expirationTime := now.Add(s.tokenExpiry)

	claims := &UserClaims{
		Email: userEmail,
		Roles: roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "kloudlite-api",
			Subject:   userEmail,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		s.logger.Error("Failed to sign JWT token", zap.Error(err))
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	s.logger.Debug("Generated JWT token",
		zap.String("email", userEmail),
		zap.Strings("roles", rolesToStrings(roles)),
		zap.Time("expires_at", expirationTime),
	)

	return tokenString, nil
}

// ValidateToken parses and validates a JWT token
func (s *authService) ValidateToken(ctx context.Context, tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
	if err != nil {
		s.logger.Warn("JWT token validation failed", zap.Error(err))
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		s.logger.Warn("Invalid JWT token claims")
		return nil, fmt.Errorf("invalid token claims")
	}

	s.logger.Debug("Validated JWT token",
		zap.String("email", claims.Email),
		zap.Strings("roles", rolesToStrings(claims.Roles)),
	)

	return claims, nil
}

// VerifyPassword authenticates a user with email and password
func (s *authService) VerifyPassword(ctx context.Context, email, password string) (*platformv1alpha1.User, error) {
	// Get user by email
	user, err := s.userService.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Warn("User not found for authentication", zap.String("email", email), zap.Error(err))
		// Check if this is a connection/TLS error
		if isConnectionError(err) {
			return nil, fmt.Errorf("failed to connect to authentication service: %w", err)
		}
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Check if user has a password set
	if user.Spec.Password == "" {
		s.logger.Warn("User has no password set", zap.String("email", email))
		return nil, fmt.Errorf("authentication failed: no password set")
	}

	// Decode base64-encoded password hash
	hashedPassword, err := base64.StdEncoding.DecodeString(user.Spec.Password)
	if err != nil {
		s.logger.Error("Failed to decode password hash", zap.String("email", email), zap.Error(err))
		return nil, fmt.Errorf("authentication failed: invalid password format")
	}

	// Verify password using bcrypt
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		s.logger.Warn("Password verification failed", zap.String("email", email), zap.Error(err))
		return nil, fmt.Errorf("authentication failed: invalid password")
	}

	// Check if user is active
	if user.Spec.Active != nil && !*user.Spec.Active {
		s.logger.Warn("Inactive user attempted login", zap.String("email", email))
		return nil, fmt.Errorf("user account is inactive")
	}

	s.logger.Info("User authenticated successfully", zap.String("email", email))
	return user, nil
}

// isConnectionError checks if the error is related to connection/TLS issues
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	connectionErrorStrings := []string{
		"tls: failed to verify certificate",
		"x509: certificate signed by unknown authority",
		"certificate not trusted",
		"certificate has expired",
		"certificate is not yet valid",
		"tls handshake error",
		"certificate authority",
		"failed to get server groups",
		"connection refused",
		"no such host",
		"timeout",
		"network is unreachable",
		"connection reset by peer",
	}

	for _, connStr := range connectionErrorStrings {
		if contains(errStr, connStr) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr)))
}

// containsSubstring checks if a string contains a substring
func containsSubstring(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}

// Helper function to convert roles to strings for logging
func rolesToStrings(roles []platformv1alpha1.RoleType) []string {
	strings := make([]string, len(roles))
	for i, role := range roles {
		strings[i] = string(role)
	}
	return strings
}
