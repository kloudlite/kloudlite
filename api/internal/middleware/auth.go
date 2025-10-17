package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/services"
	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	"go.uber.org/zap"
)

// JWTMiddleware provides JWT authentication middleware
func JWTMiddleware(authService services.AuthService, logger *zap.Logger, skipAuth bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authentication if configured (for development/testing)
		if skipAuth {
			logger.Debug("Skipping JWT authentication as configured")
			c.Next()
			return
		}

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// Log only safe, non-sensitive headers for debugging
			safeHeaders := make(map[string]string)
			sensitiveHeaderKeys := []string{"authorization", "cookie", "set-cookie", "x-api-key", "x-auth-token"}

			for key, values := range c.Request.Header {
				lowerKey := strings.ToLower(key)
				isSensitive := false
				for _, sensitive := range sensitiveHeaderKeys {
					if lowerKey == sensitive {
						isSensitive = true
						break
					}
				}
				if !isSensitive && len(values) > 0 {
					safeHeaders[key] = values[0]
				}
			}

			logger.Warn("Missing Authorization header",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.Any("safe_headers", safeHeaders),
			)
			SendErrorResponse(c, http.StatusUnauthorized, "Authorization header required")
			c.Abort()
			return
		}

		// Check for Bearer token format
		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			logger.Warn("Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := authHeader[len(bearerPrefix):]
		if tokenString == "" {
			logger.Warn("Empty JWT token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Empty token"})
			c.Abort()
			return
		}

		// Verify and parse JWT token
		userClaims, err := authService.ValidateToken(c.Request.Context(), tokenString)
		if err != nil {
			logger.Warn("JWT token validation failed", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("user_email", userClaims.Email)
		c.Set("user_roles", userClaims.Roles)

		logger.Debug("User authenticated via JWT",
			zap.String("email", userClaims.Email),
			zap.Any("roles", userClaims.Roles),
		)

		c.Next()
	}
}

// GetUserFromContext extracts user information from gin context
func GetUserFromContext(c *gin.Context) (email string, roles []platformv1alpha1.RoleType, exists bool) {
	emailValue, emailExists := c.Get("user_email")
	rolesValue, rolesExists := c.Get("user_roles")

	if !emailExists || !rolesExists {
		return "", nil, false
	}

	email, emailOk := emailValue.(string)
	if !emailOk {
		return "", nil, false
	}

	roles, rolesOk := rolesValue.([]platformv1alpha1.RoleType)
	if !rolesOk {
		return "", nil, false
	}

	return email, roles, true
}

// RequireRoles is a middleware that requires specific roles
func RequireRoles(requiredRoles ...platformv1alpha1.RoleType) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, userRoles, exists := GetUserFromContext(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Check if user has any of the required roles
		hasRequiredRole := false
		for _, userRole := range userRoles {
			for _, requiredRole := range requiredRoles {
				if userRole == requiredRole {
					hasRequiredRole = true
					break
				}
			}
			if hasRequiredRole {
				break
			}
		}

		if !hasRequiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSuperAdmin is a convenience middleware for super admin only routes
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRoles(platformv1alpha1.RoleSuperAdmin)
}

// RequireAdminOrSuperAdmin is a convenience middleware for admin+ routes
func RequireAdminOrSuperAdmin() gin.HandlerFunc {
	return RequireRoles(platformv1alpha1.RoleAdmin, platformv1alpha1.RoleSuperAdmin)
}
