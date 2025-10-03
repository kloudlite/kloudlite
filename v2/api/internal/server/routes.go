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
		WorkspaceRepository:   servicesManager.RepositoryManager.Workspaces,
		UserWebhook:          webhooks.NewUserWebhook(pkglogger.NewZapLogger(logger), servicesManager.RepositoryManager.K8sClient),
		EnvironmentWebhook:   webhooks.NewEnvironmentWebhook(pkglogger.NewZapLogger(logger), servicesManager.RepositoryManager.K8sClient, nil),
		MachineTypeWebhook:   webhooks.NewMachineTypeWebhook(servicesManager.RepositoryManager.K8sClient),
		WorkMachineWebhook:   webhooks.NewWorkMachineWebhook(servicesManager.RepositoryManager.K8sClient),
		// WorkspaceWebhook:     webhooks.NewWorkspaceWebhook(pkglogger.NewZapLogger(logger), servicesManager.RepositoryManager.K8sClient, nil), // TODO: fix webhook implementation
	}

	// API handlers with services
	apiHandlers := handlers.NewAPIHandlers(cfg)
	userHandlers := handlers.NewUserHandlers(servicesManager.Users, logger)
	authHandlers := handlers.NewAuthHandlers(servicesManager.Auth, servicesManager.Users, logger)
	environmentHandlers := handlers.NewEnvironmentHandlers(
		servicesManager.RepositoryManager.Environments,
		servicesManager.RepositoryManager.Users,
		servicesManager.RepositoryManager.K8sClient,
		logger,
	)
	machineTypeHandlers := handlers.NewMachineTypeHandlers(manager)
	workMachineHandlers := handlers.NewWorkMachineHandlers(manager)
	workspaceHandlers := handlers.NewWorkspaceHandlers(
		servicesManager.RepositoryManager.Workspaces,
		servicesManager.RepositoryManager.Users,
		servicesManager.RepositoryManager.WorkMachines,
		servicesManager.RepositoryManager.K8sClient,
		logger,
	)
	compositionHandlers := handlers.NewCompositionHandlers(
		servicesManager.RepositoryManager.Compositions,
		servicesManager.RepositoryManager.K8sClient,
		logger,
	)
	serviceHandlers := handlers.NewServiceHandlers(
		servicesManager.RepositoryManager.K8sClient,
		logger,
	)

	// Webhook handlers
	appLogger := pkglogger.NewZapLogger(logger)
	userWebhook := webhooks.NewUserWebhook(appLogger, servicesManager.RepositoryManager.K8sClient)
	environmentWebhook := webhooks.NewEnvironmentWebhook(appLogger, servicesManager.RepositoryManager.K8sClient, nil)
	// workspaceWebhook := webhooks.NewWorkspaceWebhook(appLogger, servicesManager.RepositoryManager.K8sClient, nil) // TODO: fix webhook implementation

	// JWT middleware
	jwtMiddleware := middleware.JWTMiddleware(servicesManager.Auth, logger, cfg.Auth.SkipAuthentication)

	// API routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/info", apiHandlers.GetInfo)

		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandlers.Login)
			auth.POST("/token", authHandlers.GenerateToken)
			auth.POST("/validate", authHandlers.ValidateToken)
		}

		// Public OAuth providers endpoint (for signin page)
		namespace := "default"
		k8sClient := servicesManager.RepositoryManager.K8sClient
		oauthHandlers := handlers.NewOAuthHandlers(k8sClient, namespace)
		v1.GET("/providers", oauthHandlers.GetPublicOAuthProviders)

		// Protected routes require JWT authentication
		protected := v1.Group("/")
		protected.Use(jwtMiddleware)
		{
			// User routes
			users := protected.Group("/users")
			{
				users.POST("", userHandlers.CreateUser)
				users.GET("/by-email", userHandlers.GetUserByEmail)
				users.GET("/:name", userHandlers.GetUser)
				users.PUT("/:name", userHandlers.UpdateUser)
				users.DELETE("/:name", userHandlers.DeleteUser)
				users.POST("/:name/reset-password", userHandlers.ResetUserPassword)
				users.POST("/:name/update-last-login", userHandlers.UpdateUserLastLogin)
				users.POST("/:name/activate", userHandlers.ActivateUser)
				users.POST("/:name/deactivate", userHandlers.DeactivateUser)
				users.GET("", userHandlers.ListUsers)
			}

			// Environment routes
			environments := protected.Group("/environments")
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
			machineTypes := protected.Group("/machine-types")
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
			workMachines := protected.Group("/work-machines")
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

			// Workspace routes (namespaced)
			workspaces := protected.Group("/namespaces/:namespace/workspaces")
			{
				workspaces.POST("", workspaceHandlers.CreateWorkspace)
				workspaces.GET("/:name", workspaceHandlers.GetWorkspace)
				workspaces.PUT("/:name", workspaceHandlers.UpdateWorkspace)
				workspaces.DELETE("/:name", workspaceHandlers.DeleteWorkspace)
				workspaces.GET("", workspaceHandlers.ListWorkspaces)

				// Workspace actions
				workspaces.POST("/:name/suspend", workspaceHandlers.SuspendWorkspace)
				workspaces.POST("/:name/activate", workspaceHandlers.ActivateWorkspace)
				workspaces.POST("/:name/archive", workspaceHandlers.ArchiveWorkspace)
			}

			// Composition routes (namespaced)
			compositions := protected.Group("/namespaces/:namespace/compositions")
			{
				compositions.POST("", compositionHandlers.CreateComposition)
				compositions.GET("/:name", compositionHandlers.GetComposition)
				compositions.PUT("/:name", compositionHandlers.UpdateComposition)
				compositions.DELETE("/:name", compositionHandlers.DeleteComposition)
				compositions.GET("", compositionHandlers.ListCompositions)
				compositions.GET("/:name/status", compositionHandlers.GetCompositionStatus)
			}

			// Service routes (namespaced, read-only)
			services := protected.Group("/namespaces/:namespace/services")
			{
				services.GET("", serviceHandlers.ListServices)
			}

			// OAuth Provider routes (protected - for admin management)
			oauthProviders := protected.Group("/oauth-providers")
			{
				oauthProviders.GET("", oauthHandlers.GetOAuthProviders)
				oauthProviders.PUT("/:type", oauthHandlers.UpdateOAuthProvider)
			}
		}
	}

	// Webhook endpoints (for Kubernetes admission controllers)
	webhooksGroup := router.Group("/webhooks")
	{
		webhooksGroup.POST("/validate/users", userWebhook.ValidateUser)
		webhooksGroup.POST("/mutate/users", userWebhook.MutateUser)
		webhooksGroup.POST("/validate/environments", environmentWebhook.ValidateEnvironment)
		webhooksGroup.POST("/mutate/environments", environmentWebhook.MutateEnvironment)
		// webhooksGroup.POST("/validate/workspaces", workspaceWebhook.ValidateWorkspace) // TODO: fix webhook implementation
		// webhooksGroup.POST("/mutate/workspaces", workspaceWebhook.MutateWorkspace) // TODO: fix webhook implementation
	}

	return router
}