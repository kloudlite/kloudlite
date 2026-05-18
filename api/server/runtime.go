package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kloudlite/kloudlite/api/config"
	"github.com/kloudlite/kloudlite/api/k8s"
	"github.com/kloudlite/kloudlite/api/resources/events"
	"github.com/kloudlite/kloudlite/api/resources/registry"
	"github.com/kloudlite/kloudlite/api/resources/service"
	"github.com/kloudlite/kloudlite/api/resources/store"
	"github.com/kloudlite/kloudlite/api/resources/watch"
	"github.com/kloudlite/kloudlite/controllers"
	"go.uber.org/zap"
)

type platformRuntime struct {
	httpsServer         *http.Server
	k8sClient           *k8s.Client
	controllerManager   *controllers.Manager
	watchManager        *watch.Manager
	resourceRegistry    *registry.Registry
	resourceStore       *store.Store
	controllerCtx       context.Context
	controllerCtxCancel context.CancelFunc
}

func newPlatformRuntime(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*platformRuntime, error) {
	k8sClientOptions := &k8s.ClientOptions{
		KubeconfigPath: cfg.Kubernetes.KubeconfigPath,
		Context:        cfg.Kubernetes.Context,
		MasterURL:      cfg.Kubernetes.MasterURL,
	}

	k8sClient, err := k8s.NewClient(ctx, k8sClientOptions)
	if err != nil {
		return nil, fmt.Errorf("create Kubernetes client: %w", err)
	}

	controllerManager, err := controllers.NewPlatformAPIManager(k8sClient.Config, &cfg.Installation, &cfg.Auth, logger)
	if err != nil {
		return nil, fmt.Errorf("create platform-api controller manager: %w", err)
	}

	return newPlatformRuntimeComponents(cfg, logger, k8sClient, controllerManager), nil
}

func newPlatformRuntimeComponents(cfg *config.Config, logger *zap.Logger, k8sClient *k8s.Client, controllerManager *controllers.Manager) *platformRuntime {
	resourceRegistry := registry.Default()
	resourceStore := store.New()
	resourceBroker := events.NewBroker()
	resourceService := service.NewWithEvents(resourceRegistry, resourceStore, k8sClient.RuntimeClient, resourceBroker)
	watchManager := watch.NewManagerWithEvents(resourceRegistry, resourceStore, k8sClient.RuntimeClient, logger, resourceBroker)
	router := setupWebhookRouter(cfg, logger, k8sClient, resourceService, resourceRegistry, resourceStore, watchManager, resourceBroker)
	controllerCtx, controllerCtxCancel := context.WithCancel(context.Background())

	return &platformRuntime{
		httpsServer: &http.Server{
			Addr:    ":9443",
			Handler: router,
		},
		k8sClient:           k8sClient,
		controllerManager:   controllerManager,
		watchManager:        watchManager,
		resourceRegistry:    resourceRegistry,
		resourceStore:       resourceStore,
		controllerCtx:       controllerCtx,
		controllerCtxCancel: controllerCtxCancel,
	}
}
