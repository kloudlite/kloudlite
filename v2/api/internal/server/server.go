package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kloudlite/api/v2/internal/config"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
	logger     *zap.Logger
	config     *config.Config
}

func New(cfg *config.Config, logger *zap.Logger) *Server {
	router := setupRouter(cfg, logger)

	return &Server{
		httpServer: &http.Server{
			Addr:    ":" + cfg.Port,
			Handler: router,
		},
		logger: logger,
		config: cfg,
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

	s.logger.Info("Server exited")
	return nil
}