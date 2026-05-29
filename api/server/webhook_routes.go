package server

import (
	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/config"
	"github.com/kloudlite/kloudlite/api/handlers"
	"github.com/kloudlite/kloudlite/api/k8s"
	"github.com/kloudlite/kloudlite/api/middleware"
	"github.com/kloudlite/kloudlite/api/resources/events"
	resourcehandlers "github.com/kloudlite/kloudlite/api/resources/handlers"
	"github.com/kloudlite/kloudlite/api/resources/registry"
	resourcerpc "github.com/kloudlite/kloudlite/api/resources/rpc"
	"github.com/kloudlite/kloudlite/api/resources/service"
	"github.com/kloudlite/kloudlite/api/resources/store"
	"github.com/kloudlite/kloudlite/api/resources/watch"
	"github.com/kloudlite/kloudlite/api/webhooks"
	pkglogger "github.com/kloudlite/kloudlite/pkg/logger"
	"go.uber.org/zap"
)

type routeDependencies struct {
	cfg              *config.Config
	logger           *zap.Logger
	k8sClient        *k8s.Client
	resourceService  *service.Service
	resourceRegistry *registry.Registry
	resourceStore    *store.Store
	watchManager     *watch.Manager
	resourceBroker   *events.Broker
}

// setupWebhookRouter creates a router with webhooks, health checks, and API resource routes.
func setupWebhookRouter(cfg *config.Config, logger *zap.Logger, k8sClient *k8s.Client, resourceService *service.Service, resourceRegistry *registry.Registry, resourceStore *store.Store, watchManager *watch.Manager, resourceBroker *events.Broker) *gin.Engine {
	return newPlatformRouter(routeDependencies{
		cfg:              cfg,
		logger:           logger,
		k8sClient:        k8sClient,
		resourceService:  resourceService,
		resourceRegistry: resourceRegistry,
		resourceStore:    resourceStore,
		watchManager:     watchManager,
		resourceBroker:   resourceBroker,
	})
}

func newPlatformRouter(deps routeDependencies) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(deps.logger))
	router.Use(middleware.CORS())

	// Health check endpoints
	router.GET("/health", handlers.HealthCheck)
	router.GET("/ready", handlers.ReadinessCheck)

	// Webhook handlers
	appLogger := pkglogger.NewZapLogger(deps.logger)
	userWebhook := webhooks.NewUserWebhook(appLogger, deps.k8sClient.RuntimeClient)
	environmentWebhook := webhooks.NewEnvironmentWebhook(appLogger, deps.k8sClient.RuntimeClient, nil)
	machineTypeWebhook := webhooks.NewMachineTypeGinWebhook(appLogger, deps.k8sClient.RuntimeClient)
	workMachineWebhook := webhooks.NewWorkMachineWebhook(appLogger, deps.k8sClient.RuntimeClient, deps.cfg)
	workspaceWebhook := webhooks.NewWorkspaceWebhook(appLogger, deps.k8sClient.RuntimeClient)
	envVarWebhook := webhooks.NewEnvVarWebhook(appLogger, deps.k8sClient.RuntimeClient)
	serviceMutationWebhook := webhooks.NewServiceMutationWebhook(appLogger, deps.k8sClient.RuntimeClient)
	podMutationWebhook := webhooks.NewPodMutationWebhook(appLogger, deps.k8sClient.RuntimeClient)
	checkpointWebhook := webhooks.NewCheckpointWebhook(appLogger, deps.k8sClient.RuntimeClient)

	// Webhook endpoints (for Kubernetes admission controllers)
	webhooksGroup := router.Group("/webhooks")
	{
		webhooksGroup.POST("/validate/users", userWebhook.ValidateUser)
		webhooksGroup.POST("/mutate/users", userWebhook.MutateUser)
		webhooksGroup.POST("/validate/environments", environmentWebhook.ValidateEnvironment)
		webhooksGroup.POST("/mutate/environments", environmentWebhook.MutateEnvironment)
		webhooksGroup.POST("/validate/machinetypes", machineTypeWebhook.ValidateMachineType)
		webhooksGroup.POST("/mutate/machinetypes", machineTypeWebhook.MutateMachineType)
		webhooksGroup.POST("/validate/workmachines", workMachineWebhook.ValidateWorkMachine)
		webhooksGroup.POST("/mutate/workmachines", workMachineWebhook.MutateWorkMachine)
		webhooksGroup.POST("/validate/workspaces", workspaceWebhook.ValidateWorkspace)
		webhooksGroup.POST("/mutate/workspaces", workspaceWebhook.MutateWorkspace)
		webhooksGroup.POST("/validate/configmaps", envVarWebhook.ValidateConfigMap)
		webhooksGroup.POST("/validate/secrets", envVarWebhook.ValidateSecret)
		webhooksGroup.POST("/mutate/services", serviceMutationWebhook.MutateService)
		webhooksGroup.POST("/mutate/pods", podMutationWebhook.MutatePod)
		webhooksGroup.POST("/validate/checkpoints", checkpointWebhook.ValidateCheckpoint)
		webhooksGroup.POST("/validate/checkpointrestores", checkpointWebhook.ValidateCheckpointRestore)
		webhooksGroup.POST("/validate/environmentcheckpoints", checkpointWebhook.ValidateEnvironmentCheckpoint)
		webhooksGroup.POST("/validate/environmentcheckpointrestores", checkpointWebhook.ValidateEnvironmentCheckpointRestore)
	}

	// VPN connection endpoints (used by kltun CLI)
	// Note: VPN handlers require Auth and VPN services - keeping for backward compatibility
	// TODO: Consider moving VPN to a separate service or removing if not needed
	v1 := router.Group("/api/v1")
	{
		resourceRoutes := v1.Group("")
		resourceRoutes.Use(middleware.JWTAuth(deps.cfg.Auth, deps.logger))
		resourcehandlers.New(deps.resourceService, deps.resourceRegistry, deps.resourceStore, deps.watchManager).RegisterRoutes(resourceRoutes)
		registerResourceRPC(resourceRoutes, resourcerpc.NewResourceServer(deps.resourceService, deps.resourceRegistry, deps.resourceStore, deps.resourceBroker, deps.watchManager))

		// VPN endpoints are currently disabled - uncomment if VPN service is re-enabled
		// vpnHandlers := handlers.NewVPNHandlers(vpnService, logger, cfg.Auth.JWTSecret)
		// vpn := v1.Group("/vpn")
		// {
		// 	vpn.GET("/ca-cert", vpnHandlers.GetCACert)
		// 	vpn.GET("/hosts", vpnHandlers.GetHosts)
		// 	vpn.GET("/tunnel-endpoint", vpnHandlers.GetTunnelEndpoint)
		// }

		// Placeholder info endpoint
		v1.GET("/info", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"service": "kloudlite-api",
				"mode":    "controllers+webhooks",
				"message": "resource operations are served by kloudlite-api",
			})
		})
	}

	deps.logger.Info("Webhook router initialized (webhooks + health checks + resource routes)")
	return router
}

func registerResourceRPC(group *gin.RouterGroup, server *resourcerpc.ResourceServer) {
	for _, procedure := range []string{
		resourcerpc.ProcedureGet,
		resourcerpc.ProcedureList,
		resourcerpc.ProcedureCreate,
		resourcerpc.ProcedurePatch,
		resourcerpc.ProcedureDelete,
		resourcerpc.ProcedureWatchObject,
		resourcerpc.ProcedureWatchList,
	} {
		procedure := procedure
		handler := resourcerpc.Handler(procedure, server)
		group.Any(procedure, gin.WrapH(handler))
	}
}
