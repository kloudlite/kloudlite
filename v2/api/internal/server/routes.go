package server

import (
	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/v2/api/internal/config"
	"github.com/kloudlite/kloudlite/v2/api/internal/handlers"
	"github.com/kloudlite/kloudlite/v2/api/internal/managers"
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

	// Create manager that combines repositories and webhooks
	manager := &managers.Manager{
		K8sClient:             servicesManager.RepositoryManager.K8sClient,
		UserRepository:        servicesManager.RepositoryManager.Users,
		EnvironmentRepository: servicesManager.RepositoryManager.Environments,
		MachineTypeRepository: servicesManager.RepositoryManager.MachineTypes,
		WorkMachineRepository: servicesManager.RepositoryManager.WorkMachines,
		UserWebhook:          webhooks.NewUserWebhook(pkglogger.NewZapLogger(logger)),
		EnvironmentWebhook:   webhooks.NewEnvironmentWebhook(pkglogger.NewZapLogger(logger), servicesManager.RepositoryManager.K8sClient, nil),
		MachineTypeWebhook:   webhooks.NewMachineTypeWebhook(servicesManager.RepositoryManager.K8sClient),
		WorkMachineWebhook:   webhooks.NewWorkMachineWebhook(servicesManager.RepositoryManager.K8sClient),
	}

	// API handlers with services
	apiHandlers := handlers.NewAPIHandlers(cfg)
	userHandlers := handlers.NewUserHandlers(servicesManager.Users, logger)
	environmentHandlers := handlers.NewEnvironmentHandlers(
		servicesManager.RepositoryManager.Environments,
		servicesManager.RepositoryManager.Users,
		servicesManager.RepositoryManager.K8sClient,
		logger,
	)
	machineTypeHandlers := handlers.NewMachineTypeHandlers(manager)
	workMachineHandlers := handlers.NewWorkMachineHandlers(manager)

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

		// Machine Type routes
		machineTypes := v1.Group("/machine-types")
		{
			machineTypes.GET("", machineTypeHandlers.ListMachineTypes)
			machineTypes.POST("", machineTypeHandlers.CreateMachineType)
			machineTypes.GET("/:name", machineTypeHandlers.GetMachineType)
			machineTypes.PUT("/:name", machineTypeHandlers.UpdateMachineType)
			machineTypes.DELETE("/:name", machineTypeHandlers.DeleteMachineType)
			machineTypes.PUT("/:name/activate", machineTypeHandlers.ActivateMachineType)
			machineTypes.PUT("/:name/deactivate", machineTypeHandlers.DeactivateMachineType)
			machineTypes.POST("/:name/toggle-active", machineTypeHandlers.ToggleMachineTypeActive)
		}

		// Work Machine routes
		workMachines := v1.Group("/work-machines")
		{
			// User's own machine management
			workMachines.GET("/my", workMachineHandlers.GetMyWorkMachine)
			workMachines.POST("/my", workMachineHandlers.CreateMyWorkMachine)
			workMachines.PUT("/my", workMachineHandlers.UpdateMyWorkMachine)
			workMachines.DELETE("/my", workMachineHandlers.DeleteMyWorkMachine)
			workMachines.POST("/my/start", workMachineHandlers.StartMyWorkMachine)
			workMachines.POST("/my/stop", workMachineHandlers.StopMyWorkMachine)

			// Admin routes for all machines
			workMachines.GET("", workMachineHandlers.ListAllWorkMachines)
			workMachines.GET("/:name", workMachineHandlers.GetWorkMachine)
		}

		// OAuth Provider routes
		namespace := "default" // Use the appropriate namespace
		k8sClient := servicesManager.RepositoryManager.K8sClient
		oauthHandlers := handlers.NewOAuthHandlers(k8sClient, namespace)
		oauthProviders := v1.Group("/oauth-providers")
		{
			oauthProviders.GET("", oauthHandlers.GetOAuthProviders)
			oauthProviders.PUT("/:type", oauthHandlers.UpdateOAuthProvider)
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