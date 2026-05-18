package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kloudlite/kloudlite/api/config"
	"go.uber.org/zap"
)

func JWTAuth(auth config.AuthConfig, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if auth.SkipAuthentication {
			c.Next()
			return
		}

		tokenString, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok || auth.JWTSecret == "" {
			logger.Debug("missing authentication", zap.String("path", c.Request.URL.Path))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authentication"})
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(auth.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			logger.Debug("invalid authentication", zap.String("path", c.Request.URL.Path), zap.Error(err))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Next()
	}
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	return token, token != ""
}
