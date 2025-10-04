package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/dto"
	"github.com/kloudlite/kloudlite/api/internal/services"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/pkg/apis/platform/v1alpha1"
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
	Token string                      `json:"token"`
	User  UserInfo                    `json:"user"`
	Roles []platformv1alpha1.RoleType `json:"roles"`
}

// UserInfo represents user information in auth response
type UserInfo struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName,omitempty"`
	IsActive    bool   `json:"isActive"`
}

// Login handles credential-based authentication
func (h *AuthHandlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid login request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	h.logger.Info("Processing login request", zap.String("email", req.Email))

	// Verify password and get user
	user, err := h.authService.VerifyPassword(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.Warn("Login failed", zap.String("email", req.Email), zap.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid credentials"})
		return
	}

	// Generate JWT token
	token, err := h.authService.GenerateToken(c.Request.Context(), user.Spec.Email, user.Spec.Roles)
	if err != nil {
		h.logger.Error("Failed to generate token", zap.String("email", req.Email), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to generate authentication token"})
		return
	}

	// Update last login time
	if err := h.userService.UpdateUserLastLogin(c.Request.Context(), user.Name); err != nil {
		h.logger.Warn("Failed to update last login time", zap.String("email", req.Email), zap.Error(err))
		// Don't fail the login if we can't update last login time
	}

	// Prepare response
	response := AuthResponse{
		Token: token,
		User: UserInfo{
			Email:       user.Spec.Email,
			DisplayName: user.Spec.DisplayName,
			IsActive:    user.Spec.Active != nil && *user.Spec.Active,
		},
		Roles: user.Spec.Roles,
	}

	h.logger.Info("Login successful", zap.String("email", req.Email))
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

	// Generate JWT token
	token, err := h.authService.GenerateToken(c.Request.Context(), user.Spec.Email, user.Spec.Roles)
	if err != nil {
		h.logger.Error("Failed to generate token", zap.String("email", req.Email), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to generate authentication token"})
		return
	}

	// Update last login time
	if err := h.userService.UpdateUserLastLogin(c.Request.Context(), user.Name); err != nil {
		h.logger.Warn("Failed to update last login time", zap.String("email", req.Email), zap.Error(err))
		// Don't fail the login if we can't update last login time
	}

	// Prepare response
	response := AuthResponse{
		Token: token,
		User: UserInfo{
			Email:       user.Spec.Email,
			DisplayName: user.Spec.DisplayName,
			IsActive:    user.Spec.Active != nil && *user.Spec.Active,
		},
		Roles: user.Spec.Roles,
	}

	h.logger.Info("Token generation successful", zap.String("email", req.Email))
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
