package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should set all CORS headers on GET request", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With", w.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "POST, OPTIONS, GET, PUT, DELETE, PATCH", w.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("should set all CORS headers on POST request", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.POST("/test", func(c *gin.Context) {
			c.JSON(201, gin.H{"message": "created"})
		})

		req, _ := http.NewRequest("POST", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("should handle OPTIONS preflight request with 204", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "should not reach here"})
		})

		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Empty(t, w.Body.String(), "OPTIONS request should have empty body")
	})

	t.Run("should not call next handler for OPTIONS request", func(t *testing.T) {
		handlerCalled := false

		router := gin.New()
		router.Use(CORS())
		router.OPTIONS("/test", func(c *gin.Context) {
			handlerCalled = true
			c.JSON(200, gin.H{"message": "handler called"})
		})

		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.False(t, handlerCalled, "OPTIONS request should not reach the handler")
	})

	t.Run("should call next handler for GET request", func(t *testing.T) {
		handlerCalled := false

		router := gin.New()
		router.Use(CORS())
		router.GET("/test", func(c *gin.Context) {
			handlerCalled = true
			c.JSON(200, gin.H{"message": "handler called"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.True(t, handlerCalled, "GET request should reach the handler")
		assert.Contains(t, w.Body.String(), "handler called")
	})

	t.Run("should call next handler for POST request", func(t *testing.T) {
		handlerCalled := false

		router := gin.New()
		router.Use(CORS())
		router.POST("/test", func(c *gin.Context) {
			handlerCalled = true
			c.JSON(201, gin.H{"message": "created"})
		})

		req, _ := http.NewRequest("POST", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.True(t, handlerCalled, "POST request should reach the handler")
	})

	t.Run("should set CORS headers on PUT request", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.PUT("/test/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "updated"})
		})

		req, _ := http.NewRequest("PUT", "/test/123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("should set CORS headers on DELETE request", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.DELETE("/test/:id", func(c *gin.Context) {
			c.JSON(204, gin.H{})
		})

		req, _ := http.NewRequest("DELETE", "/test/456", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("should set CORS headers on PATCH request", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.PATCH("/test/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "patched"})
		})

		req, _ := http.NewRequest("PATCH", "/test/789", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("should set CORS headers on 404 responses", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.GET("/exists", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req, _ := http.NewRequest("GET", "/not-found", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("should set CORS headers on 500 responses", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.GET("/error", func(c *gin.Context) {
			c.JSON(500, gin.H{"error": "internal error"})
		})

		req, _ := http.NewRequest("GET", "/error", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("should handle OPTIONS with custom origin header", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req, _ := http.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		// Should still allow all origins
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("should set all headers for preflight request", func(t *testing.T) {
		router := gin.New()
		router.Use(CORS())
		router.POST("/api/data", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req, _ := http.NewRequest("OPTIONS", "/api/data", nil)
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Authorization")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	})
}
