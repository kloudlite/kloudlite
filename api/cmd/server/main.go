package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kloudlite/kloudlite/api/internal/cloud"
	"github.com/kloudlite/kloudlite/api/internal/config"
	"github.com/kloudlite/kloudlite/api/internal/server"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Fetch public IP from cloud provider metadata service
	metadataProvider := cloud.NewAWSMetadataProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	publicIP, err := metadataProvider.GetPublicIP(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch public IP from cloud metadata service: %v", err)
	}
	log.Printf("Detected public IP from cloud metadata service: %s", publicIP)
	cfg.Installation.PublicIP = publicIP

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
