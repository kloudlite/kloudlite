package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	connectiontokenv1 "github.com/kloudlite/kloudlite/api/internal/controllers/connectiontoken/v1"
	"github.com/kloudlite/kloudlite/api/internal/dto"
	"github.com/kloudlite/kloudlite/api/internal/middleware"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConnectionTokenHandlers handles HTTP requests for ConnectionToken resources
type ConnectionTokenHandlers struct {
	k8sClient   client.Client
	logger      *zap.Logger
	jwtSecret   string
	sshJumpHost string
	sshPort     int
	apiURL      string
}

// NewConnectionTokenHandlers creates a new ConnectionTokenHandlers
func NewConnectionTokenHandlers(k8sClient client.Client, logger *zap.Logger, jwtSecret, sshJumpHost, apiURL string, sshPort int) *ConnectionTokenHandlers {
	return &ConnectionTokenHandlers{
		k8sClient:   k8sClient,
		logger:      logger,
		jwtSecret:   jwtSecret,
		sshJumpHost: sshJumpHost,
		sshPort:     sshPort,
		apiURL:      apiURL,
	}
}

// CreateConnectionTokenRequest represents the request to create a connection token
type CreateConnectionTokenRequest struct {
	DisplayName string `json:"displayName" binding:"required,max=100"`
	WebURL      string `json:"webUrl,omitempty"`
}

// ConnectionTokenResponse represents the response with token data
type ConnectionTokenResponse struct {
	Token *connectiontokenv1.ConnectionToken `json:"token"`
	JWT   string                             `json:"jwt"`
}

// ConnectionTokenClaims represents the JWT claims for a connection token
type ConnectionTokenClaims struct {
	Email       string `json:"email"`
	TokenID     string `json:"tokenId"`
	SSHJumpHost string `json:"sshJumpHost"`
	SSHPort     int    `json:"sshPort"`
	APIURL      string `json:"apiUrl"`
	jwt.RegisteredClaims
}

// generateTokenName creates a unique name for the connection token
func generateTokenName(userEmail string) (string, error) {
	// Generate random bytes
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	// Convert to hex string
	randomStr := hex.EncodeToString(randomBytes)

	// Create name with ct- prefix (connection-token)
	// Format: ct-{timestamp}-{random}
	timestamp := time.Now().Unix()
	name := fmt.Sprintf("ct-%d-%s", timestamp, randomStr)

	return name, nil
}

// CreateConnectionToken handles POST /connection-tokens
func (h *ConnectionTokenHandlers) CreateConnectionToken(c *gin.Context) {
	var req CreateConnectionTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid connection token request", zap.Error(err))
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: fmt.Sprintf("Invalid request payload: %v", err),
		})
		return
	}

	// Get current user from JWT middleware context
	_, userEmail, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Warn("No user found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Authentication required"})
		return
	}

	h.logger.Info("Creating connection token", zap.String("user", userEmail), zap.String("displayName", req.DisplayName))

	// Use webUrl from request if provided, otherwise fall back to apiURL
	apiURL := h.apiURL
	if req.WebURL != "" {
		apiURL = req.WebURL
		h.logger.Info("Using webUrl from request", zap.String("webUrl", req.WebURL))
	}

	// Generate unique name for the token
	tokenName, err := generateTokenName(userEmail)
	if err != nil {
		h.logger.Error("Failed to generate token name", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to generate token"})
		return
	}

	// Create ConnectionToken CR (don't use email in labels as it contains invalid characters)
	connectionToken := &connectiontokenv1.ConnectionToken{
		ObjectMeta: metav1.ObjectMeta{
			Name: tokenName,
			Labels: map[string]string{
				"kloudlite.io/type": "connection-token",
			},
		},
		Spec: connectiontokenv1.ConnectionTokenSpec{
			DisplayName: req.DisplayName,
			UserID:      userEmail,
			SSHJumpHost: h.sshJumpHost,
			SSHPort:     h.sshPort,
			APIURL:      apiURL,
		},
	}

	// Create the ConnectionToken in Kubernetes
	if err := h.k8sClient.Create(c.Request.Context(), connectionToken); err != nil {
		h.logger.Error("Failed to create connection token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to create connection token"})
		return
	}

	// Generate JWT token with connection information
	now := time.Now()
	// Connection tokens have longer expiry (1 year by default)
	expirationTime := now.Add(365 * 24 * time.Hour)

	claims := &ConnectionTokenClaims{
		Email:       userEmail,
		TokenID:     tokenName,
		SSHJumpHost: h.sshJumpHost,
		SSHPort:     h.sshPort,
		APIURL:      apiURL,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "kloudlite-api",
			Subject:   userEmail,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		h.logger.Error("Failed to sign JWT token", zap.Error(err))

		// Clean up the created ConnectionToken
		_ = h.k8sClient.Delete(c.Request.Context(), connectionToken)

		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to generate JWT token"})
		return
	}

	// Update ConnectionToken status with the JWT (ephemeral - for one-time display)
	connectionToken.Status = connectiontokenv1.ConnectionTokenStatus{
		IsReady: true,
		Message: "Token created successfully",
		Token:   jwtString, // This will be shown once and then cleared
	}

	if err := h.k8sClient.Status().Update(c.Request.Context(), connectionToken); err != nil {
		h.logger.Warn("Failed to update connection token status", zap.Error(err))
		// Don't fail the request if status update fails
	}

	h.logger.Info("Connection token created successfully", zap.String("tokenName", tokenName), zap.String("user", userEmail))

	// Return response with token and JWT
	c.JSON(http.StatusOK, ConnectionTokenResponse{
		Token: connectionToken,
		JWT:   jwtString,
	})
}

// ListConnectionTokens handles GET /connection-tokens
func (h *ConnectionTokenHandlers) ListConnectionTokens(c *gin.Context) {
	// Get current user from JWT middleware context
	_, userEmail, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Warn("No user found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Authentication required"})
		return
	}

	h.logger.Info("Listing connection tokens", zap.String("user", userEmail))

	// List all ConnectionTokens (we can't use labels with @ symbol)
	var tokenList connectiontokenv1.ConnectionTokenList
	if err := h.k8sClient.List(c.Request.Context(), &tokenList); err != nil {
		h.logger.Error("Failed to list connection tokens", zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to list connection tokens"})
		return
	}

	// Filter tokens for the current user
	var userTokens []connectiontokenv1.ConnectionToken
	for _, token := range tokenList.Items {
		if token.Spec.UserID == userEmail {
			// Clear the ephemeral JWT token from status (should only be shown on creation)
			token.Status.Token = ""
			userTokens = append(userTokens, token)
		}
	}

	h.logger.Info("Listed connection tokens", zap.String("user", userEmail), zap.Int("count", len(userTokens)))

	// Return the filtered list
	filteredList := connectiontokenv1.ConnectionTokenList{
		Items: userTokens,
	}
	c.JSON(http.StatusOK, filteredList)
}

// DeleteConnectionToken handles DELETE /connection-tokens/:name
func (h *ConnectionTokenHandlers) DeleteConnectionToken(c *gin.Context) {
	tokenName := c.Param("name")

	// Get current user from JWT middleware context
	_, userEmail, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		h.logger.Warn("No user found in context")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Authentication required"})
		return
	}

	h.logger.Info("Deleting connection token", zap.String("user", userEmail), zap.String("tokenName", tokenName))

	// Get the connection token to verify ownership
	var connectionToken connectiontokenv1.ConnectionToken
	if err := h.k8sClient.Get(c.Request.Context(), client.ObjectKey{Name: tokenName}, &connectionToken); err != nil {
		h.logger.Warn("Connection token not found", zap.String("tokenName", tokenName), zap.Error(err))
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Connection token not found"})
		return
	}

	// Verify that the token belongs to the current user
	if connectionToken.Spec.UserID != userEmail {
		h.logger.Warn("Unauthorized token deletion attempt",
			zap.String("tokenName", tokenName),
			zap.String("requester", userEmail),
			zap.String("owner", connectionToken.Spec.UserID))
		c.JSON(http.StatusForbidden, dto.ErrorResponse{Error: "You don't have permission to delete this token"})
		return
	}

	// Delete the connection token
	if err := h.k8sClient.Delete(c.Request.Context(), &connectionToken); err != nil {
		h.logger.Error("Failed to delete connection token", zap.String("tokenName", tokenName), zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "Failed to delete connection token"})
		return
	}

	h.logger.Info("Connection token deleted successfully", zap.String("tokenName", tokenName), zap.String("user", userEmail))

	c.JSON(http.StatusOK, gin.H{"message": "Connection token deleted successfully"})
}

// ValidateConnectionToken validates a connection token JWT
// This can be used by the VS Code extension to validate tokens
func (h *ConnectionTokenHandlers) ValidateConnectionToken(c *gin.Context) {
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

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &ConnectionTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.jwtSecret), nil
	})

	if err != nil {
		h.logger.Warn("Connection token validation failed", zap.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid token"})
		return
	}

	claims, ok := token.Claims.(*ConnectionTokenClaims)
	if !ok || !token.Valid {
		h.logger.Warn("Invalid connection token claims")
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Invalid token claims"})
		return
	}

	// Verify that the connection token still exists in Kubernetes
	var connectionToken connectiontokenv1.ConnectionToken
	err = h.k8sClient.Get(context.Background(), client.ObjectKey{Name: claims.TokenID}, &connectionToken)
	if err != nil {
		h.logger.Warn("Connection token not found in Kubernetes", zap.String("tokenId", claims.TokenID), zap.Error(err))
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "Token has been revoked"})
		return
	}

	// Update last used time
	now := metav1.Now()
	connectionToken.Status.LastUsed = &now
	if err := h.k8sClient.Status().Update(context.Background(), &connectionToken); err != nil {
		h.logger.Warn("Failed to update last used time", zap.String("tokenId", claims.TokenID), zap.Error(err))
		// Don't fail validation if we can't update last used time
	}

	h.logger.Info("Connection token validated successfully", zap.String("tokenId", claims.TokenID), zap.String("user", claims.Email))

	// Return token information
	c.JSON(http.StatusOK, gin.H{
		"valid":       true,
		"email":       claims.Email,
		"tokenId":     claims.TokenID,
		"sshJumpHost": claims.SSHJumpHost,
		"sshPort":     claims.SSHPort,
		"apiUrl":      claims.APIURL,
	})
}
