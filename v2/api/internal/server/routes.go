package server

import (
	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/v2/api/internal/config"
	"github.com/kloudlite/kloudlite/v2/api/internal/handlers"
	"github.com/kloudlite/kloudlite/v2/api/internal/middleware"
	"github.com/kloudlite/kloudlite/v2/api/internal/services"
	"github.com/kloudlite/kloudlite/v2/api/internal/webhooks"
	pkglogger "github.com/kloudlite/kloudlite/v2/api/pkg/logger"
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
	environmentHandlers := handlers.NewEnvironmentHandlers(
		servicesManager.RepositoryManager.Environments,
		servicesManager.RepositoryManager.Users,
		servicesManager.RepositoryManager.K8sClient,
		logger,
	)

	// Webhook handlers
	appLogger := pkglogger.NewZapLogger(logger)
	userWebhook := webhooks.NewUserWebhook(appLogger)
	environmentWebhook := webhooks.NewEnvironmentWebhook(appLogger, servicesManager.RepositoryManager.K8sClient, nil)

	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/info", apiHandlers.GetInfo)

		// User routes
		users := v1.Group("/users")
		{
			users.POST("", userHandlers.CreateUser)
			users.GET("/by-email", userHandlers.GetUserByEmail)
			users.GET("/:name", userHandlers.GetUser)
			users.PUT("/:name", userHandlers.UpdateUser)
			users.DELETE("/:name", userHandlers.DeleteUser)
			users.GET("", userHandlers.ListUsers)
		}

		// Environment routes
		environments := v1.Group("/environments")
		{
			environments.POST("", environmentHandlers.CreateEnvironment)
			environments.GET("/:name", environmentHandlers.GetEnvironment)
			environments.PUT("/:name", environmentHandlers.UpdateEnvironment)
			environments.PATCH("/:name", environmentHandlers.PatchEnvironment)
			environments.DELETE("/:name", environmentHandlers.DeleteEnvironment)
			environments.GET("", environmentHandlers.ListEnvironments)
			environments.POST("/:name/activate", environmentHandlers.ActivateEnvironment)
			environments.POST("/:name/deactivate", environmentHandlers.DeactivateEnvironment)
			environments.GET("/:name/status", environmentHandlers.GetEnvironmentStatus)
		}
	}

	// Webhook endpoints (for Kubernetes admission controllers)
	webhooksGroup := router.Group("/webhooks")
	{
		webhooksGroup.POST("/validate/users", userWebhook.ValidateUser)
		webhooksGroup.POST("/mutate/users", userWebhook.MutateUser)
		webhooksGroup.POST("/validate/environments", environmentWebhook.ValidateEnvironment)
		webhooksGroup.POST("/mutate/environments", environmentWebhook.MutateEnvironment)
	}

	return router
}