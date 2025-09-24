package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kloudlite/api/v2/internal/config"
	"github.com/kloudlite/api/v2/internal/k8s"
	"github.com/kloudlite/api/v2/internal/repository"
	"github.com/kloudlite/api/v2/internal/services"
	"go.uber.org/zap"
)

type Server struct {
	httpServer        *http.Server
	logger            *zap.Logger
	config            *config.Config
	k8sClient         *k8s.Client
	repositoryManager *repository.Manager
	servicesManager   *services.Manager
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
	})
	if err != nil {
		logger.Fatal("Failed to create services manager", zap.Error(err))
	}

	// Setup router with dependencies
	router := setupRouter(cfg, logger, servicesManager)

	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: router,
		},
		logger:            logger,
		config:            cfg,
		k8sClient:         k8sClient,
		repositoryManager: repoManager,
		servicesManager:   servicesManager,
	}
}

func (s *Server) Start() error {
	s.logger.Info("Starting server", zap.String("port", s.config.Port))

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
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