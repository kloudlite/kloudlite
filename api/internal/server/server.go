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
	"github.com/kloudlite/kloudlite/api/internal/services"
	"go.uber.org/zap"
)

type Server struct {
	httpsServer         *http.Server
	logger              *zap.Logger
	config              *config.Config
	k8sClient           *k8s.Client
	controllerManager   *controllers.Manager
	controllerCtx       context.Context
	controllerCtxCancel context.CancelFunc
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

	// Initialize controller manager
	controllerManager, err := controllers.NewManager(k8sClient.Config, &cfg.Installation, &cfg.Auth, logger)
	if err != nil {
		logger.Fatal("Failed to create controller manager", zap.Error(err))
	}

	// Setup minimal router for webhooks and VPN only
	router := setupWebhookRouter(cfg, logger, k8sClient)

	// Create cancellable context for controller manager
	controllerCtx, controllerCtxCancel := context.WithCancel(context.Background())

	return &Server{
		httpsServer: &http.Server{
			Addr:    ":8443",
			Handler: router,
		},
		logger:              logger,
		config:              cfg,
		k8sClient:           k8sClient,
		controllerManager:   controllerManager,
		controllerCtx:       controllerCtx,
		controllerCtxCancel: controllerCtxCancel,
	}
}

func (s *Server) Start() error {
	s.logger.Info("Starting Kloudlite API server (controllers + webhooks only)")

	// Start controller manager first
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

	// Start HTTPS webhook server
	go func() {
		s.logger.Info("Starting HTTPS webhook server", zap.String("addr", s.httpsServer.Addr))
		if err := s.httpsServer.ListenAndServeTLS(s.config.TLS.CertFile, s.config.TLS.KeyFile); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTPS server stopped with error", zap.Error(err))
		}
	}()

	// Give webhook server a moment to start listening
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
		s.logger.Warn("Continuing without webhook configurations")
	}

	s.logger.Info("API server started successfully",
		zap.String("mode", "controllers+webhooks"),
		zap.String("webhook_addr", s.httpsServer.Addr))

	// Keep the main goroutine alive
	select {}
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down server...")

	s.controllerCtxCancel()

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.httpsServer.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Failed to shutdown HTTPS server", zap.Error(err))
	}

	s.logger.Info("Server shutdown complete")
	return nil
}
