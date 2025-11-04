package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kloudlite/kloudlite/api/internal/config"
	"github.com/kloudlite/kloudlite/api/internal/controllers"
	"github.com/kloudlite/kloudlite/api/internal/k8s"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	"github.com/kloudlite/kloudlite/api/internal/services"
	"go.uber.org/zap"
)

type Server struct {
	httpServer          *http.Server
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
	repoManager, err := repository.NewManager(ctx, &repository.ManagerOptions{
		K8sClient: k8sClient.RuntimeClient,
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
	controllerManager, err := controllers.NewManager(k8sClient.Config, &cfg.Installation, logger)
	if err != nil {
		logger.Fatal("Failed to create controller manager", zap.Error(err))
	}

	// Initialize subdomain poller
	subdomainPoller := services.NewSubdomainPoller(&cfg.Installation, k8sClient.RuntimeClient, logger)

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
	// Start controller manager in goroutine with panic recovery
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Controller manager panicked",
					zap.Any("panic", r),
					zap.Stack("stack"))
				// Optionally: could restart controller manager here
			}
		}()

		s.logger.Info("Starting controller manager")
		if err := s.controllerManager.Start(s.controllerCtx); err != nil {
			// Only log error if it's not due to context cancellation
			if s.controllerCtx.Err() == nil {
				s.logger.Error("Controller manager stopped with error", zap.Error(err))
			} else {
				s.logger.Info("Controller manager stopped gracefully")
			}
		}
	}()

	// Start subdomain poller in goroutine with panic recovery
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

	// Start HTTP server on port 8080
	// Note: Port 8443 is reserved for the controller-runtime webhook server
	s.logger.Info("Starting HTTP server", zap.String("addr", s.httpServer.Addr))
	s.logger.Info("Webhook server will be started by controller-runtime on port 8443")

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")

	// Stop controller manager first
	s.logger.Info("Stopping controller manager...")
	s.controllerCtxCancel()

	// Stop subdomain poller
	s.logger.Info("Stopping subdomain poller...")
	s.pollerCtxCancel()
	s.subdomainPoller.Stop()

	// Graceful shutdown with timeout for both servers
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Failed to shutdown HTTP server", zap.Error(err))
	}

	// Clean up managers
	if err := s.servicesManager.Close(); err != nil {
		s.logger.Error("Failed to close services manager", zap.Error(err))
	}

	if err := s.repositoryManager.Close(); err != nil {
		s.logger.Error("Failed to close repository manager", zap.Error(err))
	}

	s.logger.Info("Server exited")
	return nil
}
