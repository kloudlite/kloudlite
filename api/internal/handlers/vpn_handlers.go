package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kloudlite/kloudlite/api/internal/dto"
	"github.com/kloudlite/kloudlite/api/internal/services"
	"go.uber.org/zap"
)

// VPNHandlers handles HTTP requests for VPN connections
type VPNHandlers struct {
	vpnService services.VPNService
	logger     *zap.Logger
	jwtSecret  string
}

// NewVPNHandlers creates a new VPNHandlers
func NewVPNHandlers(vpnService services.VPNService, logger *zap.Logger, jwtSecret string) *VPNHandlers {
	return &VPNHandlers{
		vpnService: vpnService,
		logger:     logger,
		jwtSecret:  jwtSecret,
	}
}

// GetVPNConnect handles GET /api/vpn/connect
// Fetches VPN configuration via VPN service
// Expects backend JWT token (forwarded from Next.js after validating permanent VPN token)
func (h *VPNHandlers) GetVPNConnect(c *gin.Context) {
	ctx := c.Request.Context()

	// Extract and validate Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		h.logger.Warn("VPN connect: Missing authorization header")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Authorization header required"})
		return
	}

	// Extract token from Bearer header
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		h.logger.Warn("VPN connect: Invalid authorization header format")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid authorization header format"})
		return
	}

	tokenString := authHeader[len(bearerPrefix):]

	// UserClaims matches the custom claims in auth_service.go
	type UserClaims struct {
		Username string   `json:"username"`
		Email    string   `json:"email"`
		Roles    []string `json:"roles"`
		jwt.RegisteredClaims
	}

	// Parse and validate the backend JWT token (standard auth token)
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (any, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil {
		h.logger.Warn("VPN connect: Token validation failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid token"})
		return
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		h.logger.Warn("VPN connect: Invalid token claims")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid token claims"})
		return
	}

	// Extract username for WorkMachine lookup
	username := claims.Username
	if username == "" {
		h.logger.Warn("VPN connect: Missing username in token")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid token - missing username"})
		return
	}

	h.logger.Info("VPN connect: Token validated", zap.String("username", username), zap.String("email", claims.Email))

	// Use VPN service to get configuration (using username to match WorkMachine.Spec.OwnedBy)
	vpnConfig, err := h.vpnService.GetVPNConfig(ctx, "", username)
	if err != nil {
		h.logger.Error("VPN connect: Failed to get VPN config", zap.Error(err), zap.String("username", username))

		// Determine appropriate HTTP status code based on error message
		statusCode := http.StatusInternalServerError
		errorMessage := "Failed to retrieve VPN configuration"

		errMsg := err.Error()
		if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "no work machine") {
			statusCode = http.StatusNotFound
			errorMessage = errMsg
		} else if strings.Contains(errMsg, "not configured") {
			statusCode = http.StatusNotFound
			errorMessage = errMsg
		}

		c.JSON(statusCode, dto.ErrorResponse{Error: errorMessage})
		return
	}

	// Build response
	h.logger.Info("VPN connect: Successfully returned configuration",
		zap.String("username", username),
		zap.Int("hostCount", len(vpnConfig.Hosts)))

	c.JSON(http.StatusOK, vpnConfig)
}

// GetWireGuardConfig handles GET /api/vpn/wireguard-config?device_id=UUID
func (h *VPNHandlers) GetWireGuardConfig(c *gin.Context) {
	ctx := c.Request.Context()

	// Extract device ID from query parameters
	deviceID := c.Query("device_id")
	if deviceID == "" {
		h.logger.Warn("VPN WireGuard: Missing device_id parameter")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "device_id query parameter required"})
		return
	}

	// Validate and extract username from JWT
	username, err := h.validateTokenAndGetUsername(c)
	if err != nil {
		return // Error response already sent by helper
	}

	h.logger.Info("VPN WireGuard config requested",
		zap.String("username", username),
		zap.String("deviceID", deviceID))

	// Get WireGuard configuration
	wgConfig, err := h.vpnService.GetWireGuardConfig(ctx, deviceID, username)
	if err != nil {
		h.logger.Error("VPN WireGuard: Failed to get config", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"wg_config": wgConfig})
}

// GetCACert handles GET /api/vpn/ca-cert
func (h *VPNHandlers) GetCACert(c *gin.Context) {
	ctx := c.Request.Context()

	// Validate and extract username from JWT
	username, err := h.validateTokenAndGetUsername(c)
	if err != nil {
		return // Error response already sent by helper
	}

	h.logger.Info("VPN CA cert requested", zap.String("username", username))

	// Get CA certificate
	caCert, err := h.vpnService.GetCACert(ctx, username)
	if err != nil {
		h.logger.Error("VPN CA cert: Failed to get cert", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"ca_cert": caCert})
}

// GetHosts handles GET /api/vpn/hosts
func (h *VPNHandlers) GetHosts(c *gin.Context) {
	ctx := c.Request.Context()

	// Validate and extract username from JWT
	username, err := h.validateTokenAndGetUsername(c)
	if err != nil {
		return // Error response already sent by helper
	}

	h.logger.Info("VPN hosts requested", zap.String("username", username))

	// Get hosts list
	hosts, err := h.vpnService.GetHosts(ctx, username)
	if err != nil {
		h.logger.Error("VPN hosts: Failed to get hosts", zap.Error(err))
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"hosts": hosts})
}

// validateTokenAndGetUsername validates JWT and returns username
// Returns empty string and error if validation fails (error already sent to client)
func (h *VPNHandlers) validateTokenAndGetUsername(c *gin.Context) (string, error) {
	// Extract and validate Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		h.logger.Warn("VPN: Missing authorization header")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Authorization header required"})
		return "", fmt.Errorf("missing authorization header")
	}

	// Extract token from Bearer header
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		h.logger.Warn("VPN: Invalid authorization header format")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid authorization header format"})
		return "", fmt.Errorf("invalid auth header format")
	}

	tokenString := authHeader[len(bearerPrefix):]

	// UserClaims matches the custom claims in auth_service.go
	type UserClaims struct {
		Username string   `json:"username"`
		Email    string   `json:"email"`
		Roles    []string `json:"roles"`
		jwt.RegisteredClaims
	}

	// Parse and validate JWT
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.jwtSecret), nil
	})
	if err != nil {
		h.logger.Warn("VPN: Token validation failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid token"})
		return "", err
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		h.logger.Warn("VPN: Invalid token claims")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid token claims"})
		return "", fmt.Errorf("invalid claims")
	}

	username := claims.Username
	if username == "" {
		h.logger.Warn("VPN: Missing username in token")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid token - missing username"})
		return "", fmt.Errorf("missing username")
	}

	return username, nil
}

// handleServiceError converts service errors to appropriate HTTP responses
func (h *VPNHandlers) handleServiceError(c *gin.Context, err error) {
	statusCode := http.StatusInternalServerError
	errorMessage := "Operation failed"

	errMsg := err.Error()
	if strings.Contains(errMsg, "not found") || strings.Contains(errMsg, "no work machine") {
		statusCode = http.StatusNotFound
		errorMessage = errMsg
	} else if strings.Contains(errMsg, "not configured") {
		statusCode = http.StatusNotFound
		errorMessage = errMsg
	}

	c.JSON(statusCode, dto.ErrorResponse{Error: errorMessage})
}
