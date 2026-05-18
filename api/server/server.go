package server

import (
	"context"
	"net/http"
	"time"

	"github.com/kloudlite/kloudlite/api/config"
	"github.com/kloudlite/kloudlite/api/k8s"
	"github.com/kloudlite/kloudlite/api/resources/watch"
	"github.com/kloudlite/kloudlite/api/services"
	"github.com/kloudlite/kloudlite/controllers"
	"go.uber.org/zap"
)

const (
	platformAPIStartMessage = "starting platform-api (controllers + webhooks + resources)"
	platformAPIStartedMode  = "platform-api:controllers+webhooks+resources"
)

type Server struct {
	httpsServer         *http.Server
	logger              *zap.Logger
	config              *config.Config
	k8sClient           *k8s.Client
	controllerManager   *controllers.Manager
	watchManager        *watch.Manager
	webhookCABundle     []byte
	controllerCtx       context.Context
	controllerCtxCancel context.CancelFunc
}

func New(cfg *config.Config, logger *zap.Logger) *Server {
	runtime, err := newPlatformRuntime(context.Background(), cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create platform-api runtime", zap.Error(err))
	}

	return &Server{
		httpsServer:         runtime.httpsServer,
		logger:              logger,
		config:              cfg,
		k8sClient:           runtime.k8sClient,
		controllerManager:   runtime.controllerManager,
		watchManager:        runtime.watchManager,
		webhookCABundle:     runtime.webhookCABundle,
		controllerCtx:       runtime.controllerCtx,
		controllerCtxCancel: runtime.controllerCtxCancel,
	}
}

func (s *Server) Start() error {
	s.logger.Info(platformAPIStartMessage)

	// Start controller manager first
	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Controller manager panicked",
					zap.Any("panic", r),
					zap.Stack("stack"))
			}
		}()

		s.logger.Info("starting platform-api controller manager")
		if err := s.controllerManager.Start(s.controllerCtx); err != nil {
			if s.controllerCtx.Err() == nil {
				s.logger.Error("Controller manager stopped with error", zap.Error(err))
			}
		}
	}()

	go func() {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("Resource watch manager panicked",
					zap.Any("panic", r),
					zap.Stack("stack"))
			}
		}()

		s.logger.Info("starting platform-api resource watch manager")
		s.watchManager.StartClusterScoped(s.controllerCtx)
	}()

	// Start HTTPS webhook server
	httpsErrCh := s.startHTTPSServer()

	// Give webhook server a moment to start listening
	if err := waitForHTTPSStartup(httpsErrCh, 2*time.Second); err != nil {
		s.controllerCtxCancel()
		return err
	}

	// Ensure webhook configurations now that the server is ready
	s.logger.Info("ensuring platform-api admission webhook configurations")
	webhookInstaller := services.NewWebhookInstaller(s.k8sClient.RuntimeClient, s.logger, s.webhookCABundle)
	if err := webhookInstaller.InstallWebhooks(context.Background()); err != nil {
		s.logger.Error("Failed to install webhook configurations", zap.Error(err))
		s.logger.Warn("Continuing without webhook configurations")
	}

	s.logger.Info("platform-api started successfully",
		zap.String("mode", platformAPIStartedMode),
		zap.String("webhook_addr", s.httpsServer.Addr))

	// Keep the main goroutine alive
	select {}
}

func (s *Server) startHTTPSServer() <-chan error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("starting platform-api HTTPS server", zap.String("addr", s.httpsServer.Addr))
		if err := s.httpsServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			s.logger.Error("HTTPS API server stopped with error", zap.Error(err))
			errCh <- err
		}
	}()
	return errCh
}

func waitForHTTPSStartup(errCh <-chan error, delay time.Duration) error {
	select {
	case err := <-errCh:
		return err
	case <-time.After(delay):
		return nil
	}
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
