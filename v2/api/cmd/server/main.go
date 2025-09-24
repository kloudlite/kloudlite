package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kloudlite/api/v2/internal/config"
	"github.com/kloudlite/api/v2/internal/server"
	"github.com/kloudlite/api/v2/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	appLogger, err := logger.New(cfg.LogLevel, cfg.Environment)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	// Create and start server
	srv := server.New(cfg, appLogger)

	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil {
			appLogger.Fatal("Server failed to start: " + err.Error())
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown server
	if err := srv.Shutdown(context.Background()); err != nil {
		appLogger.Fatal("Server forced to shutdown: " + err.Error())
	}
}