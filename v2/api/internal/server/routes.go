package server

import (
	"github.com/gin-gonic/gin"
	"github.com/kloudlite/api/v2/internal/config"
	"github.com/kloudlite/api/v2/internal/handlers"
	"github.com/kloudlite/api/v2/internal/middleware"
	"go.uber.org/zap"
)

func setupRouter(cfg *config.Config, logger *zap.Logger) *gin.Engine {
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())

	// Health check endpoints
	router.GET("/health", handlers.HealthCheck)
	router.GET("/ready", handlers.ReadinessCheck)

	// API handlers
	apiHandlers := handlers.NewAPIHandlers(cfg)

	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/info", apiHandlers.GetInfo)
	}

	return router
}