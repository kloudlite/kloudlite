package server

import (
	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/config"
	"github.com/kloudlite/kloudlite/api/internal/handlers"
	"github.com/kloudlite/kloudlite/api/internal/k8s"
	"github.com/kloudlite/kloudlite/api/internal/middleware"
	"github.com/kloudlite/kloudlite/api/internal/webhooks"
	pkglogger "github.com/kloudlite/kloudlite/api/pkg/logger"
	"go.uber.org/zap"
)

// setupWebhookRouter creates a minimal router with only webhooks and VPN endpoints
// All CRUD operations are now handled by Next.js Server Actions
func setupWebhookRouter(cfg *config.Config, logger *zap.Logger, k8sClient *k8s.Client) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())

	// Health check endpoints
	router.GET("/health", handlers.HealthCheck)
	router.GET("/ready", handlers.ReadinessCheck)

	// Webhook handlers
	appLogger := pkglogger.NewZapLogger(logger)
	userWebhook := webhooks.NewUserWebhook(appLogger, k8sClient.RuntimeClient)
	environmentWebhook := webhooks.NewEnvironmentWebhook(appLogger, k8sClient.RuntimeClient, nil)
	machineTypeWebhook := webhooks.NewMachineTypeGinWebhook(appLogger, k8sClient.RuntimeClient)
	workMachineWebhook := webhooks.NewWorkMachineWebhook(appLogger, k8sClient.RuntimeClient, cfg)
	workspaceWebhook := webhooks.NewWorkspaceWebhook(appLogger, k8sClient.RuntimeClient)
	envVarWebhook := webhooks.NewEnvVarWebhook(appLogger, k8sClient.RuntimeClient)
	serviceMutationWebhook := webhooks.NewServiceMutationWebhook(appLogger, k8sClient.RuntimeClient)
	podMutationWebhook := webhooks.NewPodMutationWebhook(appLogger, k8sClient.RuntimeClient)
	snapshotWebhook := webhooks.NewSnapshotWebhook(appLogger, k8sClient.RuntimeClient)

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
		webhooksGroup.POST("/validate/snapshotrequests", snapshotWebhook.ValidateSnapshotRequest)
		webhooksGroup.POST("/validate/snapshotrestores", snapshotWebhook.ValidateSnapshotRestore)
		webhooksGroup.POST("/validate/environmentsnapshotrequests", snapshotWebhook.ValidateEnvironmentSnapshotRequest)
		webhooksGroup.POST("/validate/environmentsnapshotrestores", snapshotWebhook.ValidateEnvironmentSnapshotRestore)
		webhooksGroup.POST("/validate/snapshots", snapshotWebhook.ValidateSnapshot)
	}

	// VPN connection endpoints (used by kltun CLI)
	// Note: VPN handlers require Auth and VPN services - keeping for backward compatibility
	// TODO: Consider moving VPN to a separate service or removing if not needed
	v1 := router.Group("/api/v1")
	{
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
				"message": "CRUD operations are handled by Next.js Server Actions",
			})
		})
	}

	logger.Info("Webhook router initialized (webhooks + health checks only)")
	return router
}
