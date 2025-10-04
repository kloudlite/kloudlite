package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthCheck(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("should return healthy status", func(t *testing.T) {
		// Create a test router
		router := gin.New()
		router.GET("/health", HealthCheck)

		// Create a test request
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		// Perform the request
		router.ServeHTTP(w, req)

		// Assert the response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "healthy")
		assert.Contains(t, w.Body.String(), "status")
		assert.Contains(t, w.Body.String(), "time")
	})
}

func TestReadinessCheck(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	t.Run("should return ready status", func(t *testing.T) {
		// Create a test router
		router := gin.New()
		router.GET("/ready", ReadinessCheck)

		// Create a test request
		req, _ := http.NewRequest("GET", "/ready", nil)
		w := httptest.NewRecorder()

		// Perform the request
		router.ServeHTTP(w, req)

		// Assert the response
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "ready")
		assert.Contains(t, w.Body.String(), "status")
		assert.Contains(t, w.Body.String(), "time")
	})
}
