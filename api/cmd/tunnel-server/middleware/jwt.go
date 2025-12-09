package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// UserClaims represents the JWT claims structure
type UserClaims struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	jwt.RegisteredClaims
}

// ContextKey is the type for context keys
type ContextKey string

const (
	// UserContextKey is the context key for user claims
	UserContextKey ContextKey = "user"
)

// NewJWTMiddleware creates a new JWT authentication middleware
func NewJWTMiddleware(jwtSecret string, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenString string

			// Try Authorization header first
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				const bearerPrefix = "Bearer "
				if strings.HasPrefix(authHeader, bearerPrefix) {
					tokenString = strings.TrimPrefix(authHeader, bearerPrefix)
				}
			}

			// If no Authorization header, try session cookie (for browser requests)
			// NextAuth v5 uses 'authjs' prefix by default
			if tokenString == "" {
				// Try secure cookie first (production), then non-secure (development)
				cookieNames := []string{"__Secure-authjs.session-token", "authjs.session-token"}
				for _, cookieName := range cookieNames {
					if cookie, err := r.Cookie(cookieName); err == nil && cookie.Value != "" {
						tokenString = cookie.Value
						break
					}
				}
			}

			if tokenString == "" {
				logger.Debug("no valid authentication found",
					zap.String("path", r.URL.Path),
					zap.String("remote_addr", r.RemoteAddr))
				http.Error(w, `{"error": "missing authentication"}`, http.StatusUnauthorized)
				return
			}

			// Parse and validate JWT token
			claims := &UserClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				// Validate signing method is HMAC
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil {
				logger.Debug("failed to parse JWT token",
					zap.String("path", r.URL.Path),
					zap.Error(err))
				http.Error(w, `{"error": "invalid token"}`, http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				logger.Debug("invalid JWT token",
					zap.String("path", r.URL.Path))
				http.Error(w, `{"error": "invalid token"}`, http.StatusUnauthorized)
				return
			}

			// Log authenticated request
			logger.Debug("authenticated request",
				zap.String("path", r.URL.Path),
				zap.String("username", claims.Username),
				zap.String("email", claims.Email))

			// Add claims to request context
			ctx := context.WithValue(r.Context(), UserContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserFromContext extracts user claims from context
func GetUserFromContext(ctx context.Context) (*UserClaims, bool) {
	claims, ok := ctx.Value(UserContextKey).(*UserClaims)
	return claims, ok
}
