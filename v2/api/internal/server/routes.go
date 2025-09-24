package server

import (
	"github.com/gin-gonic/gin"
	"github.com/kloudlite/api/v2/internal/config"
	"github.com/kloudlite/api/v2/internal/handlers"
	"github.com/kloudlite/api/v2/internal/middleware"
	"github.com/kloudlite/api/v2/internal/services"
	"go.uber.org/zap"
)

func setupRouter(cfg *config.Config, logger *zap.Logger, servicesManager *services.Manager) *gin.Engine {
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

	// API handlers with services
	apiHandlers := handlers.NewAPIHandlers(cfg)
	userHandlers := handlers.NewUserHandlers(servicesManager.Users, logger)

	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/info", apiHandlers.GetInfo)

		// User routes
		users := v1.Group("/users")
		{
			users.POST("", userHandlers.CreateUser)
			users.GET("/:name", userHandlers.GetUser)
			users.PUT("/:name", userHandlers.UpdateUser)
			users.DELETE("/:name", userHandlers.DeleteUser)
			users.GET("", userHandlers.ListUsers)
		}
	}

	return router
}