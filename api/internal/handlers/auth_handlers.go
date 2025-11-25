package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"github.com/kloudlite/kloudlite/api/internal/dto"
	"github.com/kloudlite/kloudlite/api/internal/services"
	"go.uber.org/zap"
)

// AuthHandlers provides HTTP handlers for authentication
type AuthHandlers struct {
	authService services.AuthService
	userService services.UserService
	logger      *zap.Logger
}

// NewAuthHandlers creates a new AuthHandlers instance
func NewAuthHandlers(authService services.AuthService, userService services.UserService, logger *zap.Logger) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		userService: userService,
		logger:      logger,
	}
}

// LoginRequest represents the payload for login requests
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// TokenRequest represents the payload for OAuth token generation
type TokenRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// AuthResponse represents the response for successful authentication
type AuthResponse struct {
	User  UserInfo                    `json:"user"`
	Roles []platformv1alpha1.RoleType `json:"roles"`
}

// UserInfo represents user information in auth response
type UserInfo struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName,omitempty"`
	IsActive    bool   `json:"isActive"`
}

// Login handles credential-based authentication
func (h *AuthHandlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request",
			zap.Error(err),
			zap.String("error_details", err.Error()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: fmt.Sprintf("Invalid request payload: %v", err),
		})
		return
	}

	h.logger.Info("Processing login request", zap.String("email", req.Email))

	// Verify password and get user
	user, err := h.authService.VerifyPassword(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Warn("Login failed", zap.String("email", req.Email), zap.Error(err))

		// Check if this is a connection error
		if isConnectionError(err) {
			c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{Error: "Authentication service temporarily unavailable - please try again later"})
			return
		}

		// Check for specific error messages
		errMsg := err.Error()
		if contains(errMsg, "failed to connect to authentication service") {
			c.JSON(http.StatusServiceUnavailable, dto.ErrorResponse{Error: "Authentication service temporarily unavailable - please try again later"})
			return
		}
		if contains(errMsg, "authentication failed: no password set") {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "Account not properly configured - please contact administrator"})
			return
		}
		if contains(errMsg, "authentication failed: invalid password") {
			c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid credentials"})
			return
		}
		if contains(errMsg, "user account is inactive") {
			c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "User account is inactive"})
			return
		}

		// Default authentication error
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Authentication failed"})
		return
	}

	// Update last login time
	if err := h.userService.UpdateUserLastLogin(c.Request.Context(), user.Name); err != nil {
		h.logger.Warn("Failed to update last login time", zap.String("email", req.Email), zap.Error(err))
		// Don't fail the login if we can't update last login time
	}

	// Prepare response - return user info only (NextAuth will generate JWT)
	response := AuthResponse{
		User: UserInfo{
			Username:    user.Name,
			Email:       user.Spec.Email,
			Name:        user.Spec.DisplayName,
			DisplayName: user.Spec.DisplayName,
			IsActive:    user.Spec.Active != nil && *user.Spec.Active,
		},
		Roles: user.Spec.Roles,
	}

	h.logger.Info("Login successful - user info returned", zap.String("email", req.Email))
	c.JSON(http.StatusOK, response)
}

// GenerateToken handles OAuth-based token generation
func (h *AuthHandlers) GenerateToken(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid token request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	h.logger.Info("Processing token generation request", zap.String("email", req.Email))

	// Get user by email (OAuth users should already exist)
	user, err := h.userService.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		h.logger.Warn("User not found for token generation", zap.String("email", req.Email), zap.Error(err))
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "User not found"})
		return
	}

	// Check if user is active
	if user.Spec.Active != nil && !*user.Spec.Active {
		h.logger.Warn("Inactive user attempted token generation", zap.String("email", req.Email))
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "User account is inactive"})
		return
	}

	// Update last login time
	if err := h.userService.UpdateUserLastLogin(c.Request.Context(), user.Name); err != nil {
		h.logger.Warn("Failed to update last login time", zap.String("email", req.Email), zap.Error(err))
		// Don't fail the login if we can't update last login time
	}

	// Prepare response - return user info only (NextAuth will generate JWT)
	response := AuthResponse{
		User: UserInfo{
			Username:    user.Name,
			Email:       user.Spec.Email,
			Name:        user.Spec.DisplayName,
			DisplayName: user.Spec.DisplayName,
			IsActive:    user.Spec.Active != nil && *user.Spec.Active,
		},
		Roles: user.Spec.Roles,
	}

	h.logger.Info("OAuth user info returned", zap.String("email", req.Email))
	c.JSON(http.StatusOK, response)
}

// ValidateToken handles token validation requests
func (h *AuthHandlers) ValidateToken(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Authorization header required"})
		return
	}

	// Extract token from Bearer header
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid authorization header format"})
		return
	}

	tokenString := authHeader[len(bearerPrefix):]

	// Validate token
	claims, err := h.authService.ValidateToken(c.Request.Context(), tokenString)
	if err != nil {
		h.logger.Warn("Token validation failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid token"})
		return
	}

	// Return user information
	c.JSON(http.StatusOK, dto.ValidateTokenResponse{
		Valid: true,
		User: map[string]interface{}{
			"email": claims.Email,
			"roles": claims.Roles,
		},
	})
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
		if strings.Contains(strings.ToLower(errStr), strings.ToLower(connStr)) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
