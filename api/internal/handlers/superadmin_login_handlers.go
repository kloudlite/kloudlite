package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"github.com/kloudlite/kloudlite/api/internal/dto"
	"github.com/kloudlite/kloudlite/api/internal/services"
	"go.uber.org/zap"
)

// SuperAdminLoginHandlers provides HTTP handlers for superadmin login via token
type SuperAdminLoginHandlers struct {
	authService        services.AuthService
	installationSecret string
	logger             *zap.Logger
}

// NewSuperAdminLoginHandlers creates a new SuperAdminLoginHandlers instance
func NewSuperAdminLoginHandlers(authService services.AuthService, installationSecret string, logger *zap.Logger) *SuperAdminLoginHandlers {
	return &SuperAdminLoginHandlers{
		authService:        authService,
		installationSecret: installationSecret,
		logger:             logger,
	}
}

// SuperAdminLoginTokenPayload represents the token payload from console
type SuperAdminLoginTokenPayload struct {
	Type            string `json:"type"`
	InstallationID  string `json:"installationId"`
	InstallationKey string `json:"installationKey"`
	Timestamp       int64  `json:"timestamp"`
	Nonce           string `json:"nonce"`
	ExpiresAt       int64  `json:"expiresAt"`
}

// ValidateSuperAdminLoginRequest represents the request to validate superadmin login token
type ValidateSuperAdminLoginRequest struct {
	Token string `json:"token" binding:"required"`
}

// ValidateSuperAdminLoginResponse represents the response for valid superadmin login
type ValidateSuperAdminLoginResponse struct {
	Valid bool                        `json:"valid"`
	Token string                      `json:"token"` // JWT token for API access
	User  UserInfo                    `json:"user"`
	Roles []platformv1alpha1.RoleType `json:"roles"`
}

// ValidateSuperAdminLogin validates superadmin login token and returns JWT for super admin access
func (h *SuperAdminLoginHandlers) ValidateSuperAdminLogin(c *gin.Context) {
	var req ValidateSuperAdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid superadmin login validation request",
			zap.Error(err),
			zap.String("error_details", err.Error()))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: fmt.Sprintf("Invalid request payload: %v", err),
		})
		return
	}

	h.logger.Info("Validating superadmin login token")

	// Split token into payload and signature
	var payloadB64, signature string
	dotIndex := -1
	for i := 0; i < len(req.Token); i++ {
		if req.Token[i] == '.' {
			dotIndex = i
			break
		}
	}

	if dotIndex == -1 {
		h.logger.Warn("Invalid token format - missing separator")
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid token format",
		})
		return
	}

	payloadB64 = req.Token[:dotIndex]
	signature = req.Token[dotIndex+1:]

	// Decode payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		h.logger.Warn("Failed to decode token payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid token format",
		})
		return
	}

	var payload SuperAdminLoginTokenPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		h.logger.Warn("Failed to parse token payload", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid token format",
		})
		return
	}

	// Verify signature using installation secret
	expectedSignature := h.computeSignature(payloadBytes)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		h.logger.Warn("Invalid token signature")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "Invalid token signature",
		})
		return
	}

	// Check token type
	if payload.Type != "superadmin-login" {
		h.logger.Warn("Invalid token type", zap.String("type", payload.Type))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Invalid token type",
		})
		return
	}

	// Check expiry
	now := time.Now().UnixMilli()
	if now > payload.ExpiresAt {
		h.logger.Warn("Token expired",
			zap.Int64("now", now),
			zap.Int64("expiresAt", payload.ExpiresAt))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error: "Token has expired. Please generate a new admin login URL.",
		})
		return
	}

	// Token is valid - generate JWT for super admin access
	// Use "root" as the username - this is a virtual user not stored in user repo
	superAdminUsername := "root"
	superAdminEmail := "root@kloudlite.io"
	roles := []platformv1alpha1.RoleType{platformv1alpha1.RoleSuperAdmin}

	jwtToken, err := h.authService.GenerateToken(c.Request.Context(), superAdminUsername, superAdminEmail, roles)
	if err != nil {
		h.logger.Error("Failed to generate JWT token for super admin", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error: "Failed to generate authentication token",
		})
		return
	}

	h.logger.Info("Super admin login successful",
		zap.String("installation_id", payload.InstallationID),
		zap.String("installation_key", payload.InstallationKey))

	c.JSON(http.StatusOK, ValidateSuperAdminLoginResponse{
		Valid: true,
		Token: jwtToken,
		User: UserInfo{
			Email:       superAdminUsername,
			DisplayName: "root",
			IsActive:    true,
		},
		Roles: roles,
	})
}

// computeSignature computes HMAC-SHA256 signature for the payload
func (h *SuperAdminLoginHandlers) computeSignature(payload []byte) string {
	mac := hmac.New(sha256.New, []byte(h.installationSecret))
	mac.Write(payload)
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
