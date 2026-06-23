package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"github.com/kloudlite/kloudlite/api/config"
	"github.com/kloudlite/kloudlite/api/k8s"
	"github.com/kloudlite/kloudlite/api/resources/events"
	"github.com/kloudlite/kloudlite/api/resources/registry"
	"github.com/kloudlite/kloudlite/api/resources/service"
	"github.com/kloudlite/kloudlite/api/resources/store"
	"github.com/kloudlite/kloudlite/api/resources/watch"
	"github.com/kloudlite/kloudlite/controllers"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const platformAPIServiceName = "api-server"

type platformRuntime struct {
	httpsServer         *http.Server
	k8sClient           *k8s.Client
	controllerManager   *controllers.Manager
	watchManager        *watch.Manager
	webhookCABundle     []byte
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
	webhookTLS, err := ensureWebhookTLSSecret(ctx, k8sClient.RuntimeClient, webhookTLSOptions{Namespace: platformAPINamespace(cfg), ServiceName: platformAPIServiceName})
	if err != nil {
		return nil, fmt.Errorf("ensure webhook TLS secret: %w", err)
	}
	logger.Info("webhook TLS secret ready",
		zap.String("secret", webhookTLSSecretName),
		zap.String("namespace", platformAPINamespace(cfg)),
		zap.String("source", webhookTLS.Source))

	controllerManager, err := controllers.NewPlatformAPIManager(k8sClient.Config, &cfg.Installation, &cfg.Auth, logger)
	if err != nil {
		return nil, fmt.Errorf("create platform-api controller manager: %w", err)
	}

	return newPlatformRuntimeComponents(cfg, logger, k8sClient, controllerManager, webhookTLS), nil
}

func newPlatformRuntimeComponents(cfg *config.Config, logger *zap.Logger, k8sClient *k8s.Client, controllerManager *controllers.Manager, webhookTLS *webhookTLSBundle) *platformRuntime {
	resourceRegistry := registry.Default()
	resourceStore := store.New()
	resourceBroker := events.NewBroker()

	// Create a non-cached client for direct reads (no informer cache staleness)
	var kubeDirect client.Client = k8sClient.RuntimeClient
	if k8sClient.Config != nil {
		var err error
		kubeDirect, err = client.New(k8sClient.Config, client.Options{Scheme: k8sClient.Scheme})
		if err != nil {
			panic(fmt.Sprintf("create direct k8s client: %v", err))
		}
	}
	resourceService := service.NewWithEventsAndReadClient(resourceRegistry, resourceStore, k8sClient.RuntimeClient, kubeDirect, resourceBroker)
	watchManager := watch.NewManagerWithEvents(resourceRegistry, resourceStore, k8sClient.RuntimeClient, logger, resourceBroker)
	router := setupWebhookRouter(cfg, logger, k8sClient, resourceService, resourceRegistry, resourceStore, watchManager, resourceBroker)
	controllerCtx, controllerCtxCancel := context.WithCancel(context.Background())

	return &platformRuntime{
		httpsServer: &http.Server{
			Addr:      ":9443",
			Handler:   router,
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{webhookTLS.Certificate}, MinVersion: tls.VersionTLS12},
		},
		k8sClient:           k8sClient,
		controllerManager:   controllerManager,
		watchManager:        watchManager,
		webhookCABundle:     webhookTLS.CertPEM,
		resourceRegistry:    resourceRegistry,
		resourceStore:       resourceStore,
		controllerCtx:       controllerCtx,
		controllerCtxCancel: controllerCtxCancel,
	}
}

func platformAPINamespace(cfg *config.Config) string {
	if namespace := os.Getenv("POD_NAMESPACE"); namespace != "" {
		return namespace
	}
	if cfg.Kubernetes.DefaultNamespace != "" {
		return cfg.Kubernetes.DefaultNamespace
	}
	return "default"
}
