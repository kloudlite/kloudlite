package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kloudlite/kloudlite/api/internal/services"
	"go.uber.org/zap"
)

// RegistryAuthHandlers handles Docker Registry v2 token authentication
type RegistryAuthHandlers struct {
	authService services.AuthService
	jwtSecret   string
	logger      *zap.Logger
}

// NewRegistryAuthHandlers creates a new RegistryAuthHandlers
func NewRegistryAuthHandlers(authService services.AuthService, jwtSecret string, logger *zap.Logger) *RegistryAuthHandlers {
	return &RegistryAuthHandlers{
		authService: authService,
		jwtSecret:   jwtSecret,
		logger:      logger,
	}
}

// DockerTokenClaims represents the JWT claims for Docker Registry token
// See: https://docs.docker.com/registry/spec/auth/jwt/
type DockerTokenClaims struct {
	Access []DockerAccessEntry `json:"access"`
	jwt.RegisteredClaims
}

// DockerAccessEntry represents a single access entry in Docker token
type DockerAccessEntry struct {
	Type    string   `json:"type"`
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
}

// RegistryTokenRequest represents the token request from Docker client
type RegistryTokenRequest struct {
	Service string `form:"service"`           // The service name (registry hostname)
	Scope   string `form:"scope"`             // The requested scope (e.g., "repository:username/image:push,pull")
	Account string `form:"account,omitempty"` // Optional account name
}

// TokenResponse represents the token response to Docker client
type TokenResponse struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token,omitempty"` // Alias for token (some clients use this)
	ExpiresIn   int    `json:"expires_in"`
	IssuedAt    string `json:"issued_at"`
}

// GetToken handles the Docker Registry v2 token authentication endpoint
// This is called by Docker when the registry returns a 401 with WWW-Authenticate header
// Docker sends Basic Auth with username and password (Kloudlite JWT token as password)
func (h *RegistryAuthHandlers) GetToken(c *gin.Context) {
	// Parse query parameters
	var req RegistryTokenRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("Invalid token request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	h.logger.Debug("Registry token request",
		zap.String("service", req.Service),
		zap.String("scope", req.Scope),
		zap.String("account", req.Account),
	)

	// Extract Basic Auth credentials
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
		h.logger.Warn("Missing or invalid Authorization header")
		c.Header("WWW-Authenticate", `Basic realm="Kloudlite Registry"`)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Decode Basic Auth
	encodedCreds := strings.TrimPrefix(authHeader, "Basic ")
	decodedCreds, err := base64.StdEncoding.DecodeString(encodedCreds)
	if err != nil {
		h.logger.Warn("Failed to decode Basic Auth", zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Split username:password (password is the Kloudlite JWT token)
	parts := strings.SplitN(string(decodedCreds), ":", 2)
	if len(parts) != 2 {
		h.logger.Warn("Invalid Basic Auth format")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials format"})
		return
	}

	username := parts[0]
	kloudliteToken := parts[1]

	// Validate the Kloudlite JWT token
	claims, err := h.authService.ValidateToken(c.Request.Context(), kloudliteToken)
	if err != nil {
		h.logger.Warn("Invalid Kloudlite token", zap.String("username", username), zap.Error(err))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Verify username matches the token's username
	if claims.Username != username {
		h.logger.Warn("Username mismatch",
			zap.String("provided", username),
			zap.String("token_username", claims.Username),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "username does not match token"})
		return
	}

	// Parse and authorize the requested scope
	accessEntries, err := h.authorizeScope(req.Scope, claims.Username)
	if err != nil {
		h.logger.Warn("Scope authorization failed",
			zap.String("username", username),
			zap.String("scope", req.Scope),
			zap.Error(err),
		)
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Generate Docker Registry token
	now := time.Now()
	expiresIn := 3600 // 1 hour

	dockerClaims := DockerTokenClaims{
		Access: accessEntries,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "kloudlite-registry-auth",
			Subject:   username,
			Audience:  jwt.ClaimStrings{req.Service},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expiresIn) * time.Second)),
			NotBefore: jwt.NewNumericDate(now.Add(-10 * time.Second)), // Allow 10s clock skew
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, dockerClaims)
	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		h.logger.Error("Failed to sign Docker token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	h.logger.Info("Registry token issued",
		zap.String("username", username),
		zap.String("scope", req.Scope),
		zap.Int("access_count", len(accessEntries)),
	)

	c.JSON(http.StatusOK, TokenResponse{
		Token:       tokenString,
		AccessToken: tokenString,
		ExpiresIn:   expiresIn,
		IssuedAt:    now.Format(time.RFC3339),
	})
}

// authorizeScope parses the requested scope and returns authorized access entries
// Scope format: "repository:namespace/image:actions" (e.g., "repository:karthik/myapp:push,pull")
// Users can only push to their own namespace (username/*)
func (h *RegistryAuthHandlers) authorizeScope(scope string, username string) ([]DockerAccessEntry, error) {
	if scope == "" {
		// No scope requested - return empty access (authentication only)
		return []DockerAccessEntry{}, nil
	}

	var entries []DockerAccessEntry

	// Multiple scopes can be requested, separated by space
	scopes := strings.Split(scope, " ")

	for _, s := range scopes {
		// Parse scope: "type:name:actions"
		parts := strings.SplitN(s, ":", 3)
		if len(parts) != 3 {
			return nil, fmt.Errorf("invalid scope format: %s", s)
		}

		scopeType := parts[0]
		scopeName := parts[1]    // e.g., "username/image" or "library/image"
		scopeActions := parts[2] // e.g., "push,pull" or "pull"

		if scopeType != "repository" {
			// We only handle repository scopes
			return nil, fmt.Errorf("unsupported scope type: %s", scopeType)
		}

		// Parse actions
		requestedActions := strings.Split(scopeActions, ",")
		authorizedActions := []string{}

		// Check if this is the user's namespace
		// Repository name format: "namespace/image" or just "image" (library namespace)
		isOwnNamespace := strings.HasPrefix(scopeName, username+"/")

		for _, action := range requestedActions {
			action = strings.TrimSpace(action)
			switch action {
			case "pull":
				// Allow pull for everyone (read access)
				authorizedActions = append(authorizedActions, "pull")
			case "push":
				// Only allow push to user's own namespace
				if isOwnNamespace {
					authorizedActions = append(authorizedActions, "push")
				} else {
					h.logger.Warn("Push denied - not user's namespace",
						zap.String("username", username),
						zap.String("repository", scopeName),
					)
					return nil, fmt.Errorf("push denied: can only push to %s/* repositories", username)
				}
			case "delete":
				// Only allow delete in user's own namespace
				if isOwnNamespace {
					authorizedActions = append(authorizedActions, "delete")
				}
			default:
				h.logger.Warn("Unknown action requested", zap.String("action", action))
			}
		}

		if len(authorizedActions) > 0 {
			entries = append(entries, DockerAccessEntry{
				Type:    scopeType,
				Name:    scopeName,
				Actions: authorizedActions,
			})
		}
	}

	return entries, nil
}
