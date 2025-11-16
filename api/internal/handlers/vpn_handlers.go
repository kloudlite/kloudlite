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

	// Parse and validate the backend JWT token (standard auth token)
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
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

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		h.logger.Warn("VPN connect: Invalid token claims")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid token claims"})
		return
	}

	// Extract email from subject
	userEmail := claims.Subject
	if userEmail == "" {
		h.logger.Warn("VPN connect: Missing user email in token")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid token - missing user email"})
		return
	}

	h.logger.Info("VPN connect: Token validated", zap.String("user", userEmail))

	// Use VPN service to get configuration (no tokenID needed anymore)
	vpnConfig, err := h.vpnService.GetVPNConfig(ctx, "", userEmail)
	if err != nil {
		h.logger.Error("VPN connect: Failed to get VPN config", zap.Error(err), zap.String("user", userEmail))

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
		zap.String("user", userEmail),
		zap.Int("hostCount", len(vpnConfig.Hosts)))

	c.JSON(http.StatusOK, vpnConfig)
}
