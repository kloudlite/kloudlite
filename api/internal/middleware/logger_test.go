package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestLogger(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should log successful request with all fields", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		router := gin.New()
		router.Use(Logger(logger))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "success"})
		})

		req, _ := http.NewRequest("GET", "/test?param=value", nil)
		req.Header.Set("User-Agent", "test-agent")
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, logs.Len())

		logEntry := logs.All()[0]
		assert.Equal(t, "HTTP Request", logEntry.Message)
		assert.Equal(t, zapcore.InfoLevel, logEntry.Level)

		// Verify all expected fields are present
		fields := logEntry.ContextMap()
		assert.Equal(t, int64(200), fields["status"])
		assert.Equal(t, "GET", fields["method"])
		assert.Equal(t, "/test", fields["path"])
		assert.Equal(t, "param=value", fields["query"])
		assert.Equal(t, "test-agent", fields["user-agent"])
		assert.NotNil(t, fields["latency"])
		assert.NotNil(t, fields["ip"])
	})

	t.Run("should log POST request", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		router := gin.New()
		router.Use(Logger(logger))
		router.POST("/create", func(c *gin.Context) {
			c.JSON(201, gin.H{"message": "created"})
		})

		req, _ := http.NewRequest("POST", "/create", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, 1, logs.Len())

		logEntry := logs.All()[0]
		fields := logEntry.ContextMap()
		assert.Equal(t, int64(201), fields["status"])
		assert.Equal(t, "POST", fields["method"])
		assert.Equal(t, "/create", fields["path"])
	})

	t.Run("should log PUT request", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		router := gin.New()
		router.Use(Logger(logger))
		router.PUT("/update/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "updated"})
		})

		req, _ := http.NewRequest("PUT", "/update/123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, logs.Len())

		logEntry := logs.All()[0]
		fields := logEntry.ContextMap()
		assert.Equal(t, "PUT", fields["method"])
		assert.Equal(t, "/update/123", fields["path"])
	})

	t.Run("should log DELETE request", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		router := gin.New()
		router.Use(Logger(logger))
		router.DELETE("/delete/:id", func(c *gin.Context) {
			c.JSON(204, gin.H{})
		})

		req, _ := http.NewRequest("DELETE", "/delete/456", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, 1, logs.Len())

		logEntry := logs.All()[0]
		fields := logEntry.ContextMap()
		assert.Equal(t, int64(204), fields["status"])
		assert.Equal(t, "DELETE", fields["method"])
	})

	t.Run("should log 404 not found", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		router := gin.New()
		router.Use(Logger(logger))
		router.GET("/exists", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req, _ := http.NewRequest("GET", "/not-found", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		assert.Equal(t, 1, logs.Len())

		logEntry := logs.All()[0]
		fields := logEntry.ContextMap()
		assert.Equal(t, int64(404), fields["status"])
		assert.Equal(t, "/not-found", fields["path"])
	})

	t.Run("should log 500 internal server error", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		router := gin.New()
		router.Use(Logger(logger))
		router.GET("/error", func(c *gin.Context) {
			c.JSON(500, gin.H{"error": "internal error"})
		})

		req, _ := http.NewRequest("GET", "/error", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Equal(t, 1, logs.Len())

		logEntry := logs.All()[0]
		fields := logEntry.ContextMap()
		assert.Equal(t, int64(500), fields["status"])
	})

	t.Run("should log request without query parameters", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		router := gin.New()
		router.Use(Logger(logger))
		router.GET("/no-query", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req, _ := http.NewRequest("GET", "/no-query", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, logs.Len())

		logEntry := logs.All()[0]
		fields := logEntry.ContextMap()
		assert.Equal(t, "/no-query", fields["path"])
		assert.Equal(t, "", fields["query"])
	})

	t.Run("should log request with multiple query parameters", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		router := gin.New()
		router.Use(Logger(logger))
		router.GET("/search", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req, _ := http.NewRequest("GET", "/search?q=test&limit=10&offset=20", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, logs.Len())

		logEntry := logs.All()[0]
		fields := logEntry.ContextMap()
		assert.Equal(t, "q=test&limit=10&offset=20", fields["query"])
	})

	t.Run("should measure latency", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		router := gin.New()
		router.Use(Logger(logger))
		router.GET("/slow", func(c *gin.Context) {
			time.Sleep(10 * time.Millisecond)
			c.JSON(200, gin.H{"message": "ok"})
		})

		req, _ := http.NewRequest("GET", "/slow", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, logs.Len())

		logEntry := logs.All()[0]
		fields := logEntry.ContextMap()
		latency, ok := fields["latency"].(time.Duration)
		assert.True(t, ok, "latency should be a time.Duration")
		assert.GreaterOrEqual(t, latency, 10*time.Millisecond, "latency should be at least 10ms")
	})

	t.Run("should log client IP", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		router := gin.New()
		router.Use(Logger(logger))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "10.0.0.1:8080"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, logs.Len())

		logEntry := logs.All()[0]
		fields := logEntry.ContextMap()
		assert.NotEmpty(t, fields["ip"])
	})

	t.Run("should log user agent", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		router := gin.New()
		router.Use(Logger(logger))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		req.Header.Set("User-Agent", "Mozilla/5.0 (Test)")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, logs.Len())

		logEntry := logs.All()[0]
		fields := logEntry.ContextMap()
		assert.Equal(t, "Mozilla/5.0 (Test)", fields["user-agent"])
	})

	t.Run("should log request with empty user agent", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		router := gin.New()
		router.Use(Logger(logger))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		// Don't set User-Agent header
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, logs.Len())

		logEntry := logs.All()[0]
		fields := logEntry.ContextMap()
		assert.Equal(t, "", fields["user-agent"])
	})

	t.Run("should log multiple requests", func(t *testing.T) {
		core, logs := observer.New(zapcore.InfoLevel)
		logger := zap.New(core)

		router := gin.New()
		router.Use(Logger(logger))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "ok"})
		})

		// Make 3 requests
		for i := 0; i < 3; i++ {
			req, _ := http.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
		}

		assert.Equal(t, 3, logs.Len())
		for _, logEntry := range logs.All() {
			assert.Equal(t, "HTTP Request", logEntry.Message)
		}
	})
}
