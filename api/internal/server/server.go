package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kloudlite/kloudlite/api/internal/config"
	"github.com/kloudlite/kloudlite/api/internal/controllers"
	"github.com/kloudlite/kloudlite/api/internal/k8s"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"github.com/kloudlite/kloudlite/api/internal/services"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
)

type Server struct {
	httpServer          *http.Server
	httpsServer         *http.Server
	logger              *zap.Logger
	config              *config.Config
	k8sClient           *k8s.Client
	repositoryManager   *repository.Manager
	servicesManager     *services.Manager
	controllerManager   *controllers.Manager
	subdomainPoller     *services.SubdomainPoller
	controllerCtx       context.Context
	controllerCtxCancel context.CancelFunc
	pollerCtx           context.Context
	pollerCtxCancel     context.CancelFunc
}

func New(cfg *config.Config, logger *zap.Logger) *Server {
	ctx := context.Background()

	// Initialize Kubernetes client
	k8sClientOptions := &k8s.ClientOptions{
		KubeconfigPath: cfg.Kubernetes.KubeconfigPath,
		Context:        cfg.Kubernetes.Context,
		MasterURL:      cfg.Kubernetes.MasterURL,
	}

	k8sClient, err := k8s.NewClient(ctx, k8sClientOptions)
	if err != nil {
		logger.Fatal("Failed to create Kubernetes client", zap.Error(err))
	}

	// Initialize repository manager
	clientset, _ := k8sClient.Clientset.(*kubernetes.Clientset)
	repoManager, err := repository.NewManager(ctx, &repository.ManagerOptions{
		K8sClient: k8sClient.RuntimeClient,
		Clientset: clientset,
	})
	if err != nil {
		logger.Fatal("Failed to create repository manager", zap.Error(err))
	}

	// Initialize services manager
	servicesManager, err := services.NewManager(ctx, &services.ManagerOptions{
		RepositoryManager: repoManager,
		Config:            cfg,
		Logger:            logger,
	})
	if err != nil {
		logger.Fatal("Failed to create services manager", zap.Error(err))
	}

	// Initialize controller manager
	controllerManager, err := controllers.NewManager(k8sClient.Config, &cfg.Installation, &cfg.Auth, logger)
	if err != nil {
		logger.Fatal("Failed to create controller manager", zap.Error(err))
	}

	caInitializer := services.NewCAInitializer(k8sClient.RuntimeClient, logger)

	// Initialize subdomain poller
	subdomainPoller := services.NewSubdomainPoller(&cfg.Installation, k8sClient.RuntimeClient, caInitializer, logger)

	// Setup router with dependencies
	router := setupRouter(cfg, logger, servicesManager)

	// Create cancellable context for controller manager
	controllerCtx, controllerCtxCancel := context.WithCancel(context.Background())

	// Create cancellable context for subdomain poller
	pollerCtx, pollerCtxCancel := context.WithCancel(context.Background())

	return &Server{
		httpServer: &http.Server{
			Addr:    ":8080",
			Handler: router,
		},
		httpsServer: &http.Server{
			Addr:    ":8443",
			Handler: router,
		},
		logger:              logger,
		config:              cfg,
		k8sClient:           k8sClient,
		repositoryManager:   repoManager,
		servicesManager:     servicesManager,
		controllerManager:   controllerManager,
		subdomainPoller:     subdomainPoller,
		controllerCtx:       controllerCtx,
		controllerCtxCancel: controllerCtxCancel,
		pollerCtx:           pollerCtx,
		pollerCtxCancel:     pollerCtxCancel,
	}
}

func (s *Server) Start() error {
	// Start HTTPS webhook server first (needed for webhooks during subdomain setup)
	go func() {
		s.logger.Info("Starting HTTPS webhook server", zap.String("addr", s.httpsServer.Addr))
		if err := s.httpsServer.ListenAndServeTLS(s.config.TLS.CertFile, s.config.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTPS server stopped with error", zap.Error(err))
		}
	}()

	// Start HTTP server in background
	go func() {
		s.logger.Info("Starting HTTP server", zap.String("addr", s.httpServer.Addr))
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTP server stopped with error", zap.Error(err))
		}
	}()

	// Give servers a moment to start listening
	time.Sleep(2 * time.Second)

	// Install webhook configurations now that the server is ready
	s.logger.Info("Installing webhook configurations...")
	caBundle, err := os.ReadFile(s.config.TLS.CertFile)
	if err != nil {
		s.logger.Error("Failed to read webhook CA certificate", zap.Error(err))
		return fmt.Errorf("failed to read webhook CA certificate: %w", err)
	}

	webhookInstaller := services.NewWebhookInstaller(s.k8sClient.RuntimeClient, s.logger, caBundle)
	if err := webhookInstaller.InstallWebhooks(context.Background()); err != nil {
		s.logger.Error("Failed to install webhook configurations", zap.Error(err))
		// Don't fail startup, just log the error
		s.logger.Warn("Continuing without webhook configurations")
	}

	// Start controller manager first so webhook handlers are registered
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Controller manager panicked",
					zap.Any("panic", r),
					zap.Stack("stack"))
			}
		}()

		s.logger.Info("Starting controller manager")
		if err := s.controllerManager.Start(s.controllerCtx); err != nil {
			if s.controllerCtx.Err() == nil {
				s.logger.Error("Controller manager stopped with error", zap.Error(err))
			}
		}
	}()

	// Give controller manager a moment to register webhook handlers
	time.Sleep(2 * time.Second)

	// Start subdomain poller after controller manager is ready
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Subdomain poller panicked",
					zap.Any("panic", r),
					zap.Stack("stack"))
			}
		}()

		s.subdomainPoller.Start(s.pollerCtx)
	}()

	// Wait for subdomain to be ready
	waitCtx, waitCancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer waitCancel()

	s.logger.Info("Waiting for subdomain...")
	if err := s.subdomainPoller.WaitUntilReady(waitCtx); err != nil {
		s.logger.Warn("Subdomain not ready, continuing anyway", zap.Error(err))
	}

	// Keep the main goroutine alive
	select {}
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")

	s.controllerCtxCancel()
	s.pollerCtxCancel()
	s.subdomainPoller.Stop()

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Failed to shutdown HTTP server", zap.Error(err))
	}

	if err := s.httpsServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Failed to shutdown HTTPS server", zap.Error(err))
	}

	if err := s.servicesManager.Close(); err != nil {
		s.logger.Error("Failed to close services manager", zap.Error(err))
	}

	if err := s.repositoryManager.Close(); err != nil {
		s.logger.Error("Failed to close repository manager", zap.Error(err))
	}

	s.logger.Info("Server shutdown complete")
	return nil
}
