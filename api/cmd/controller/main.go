package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/kloudlite/kloudlite/api/internal/controllers"
	"github.com/kloudlite/kloudlite/api/internal/k8s"
	"github.com/kloudlite/kloudlite/api/pkg/logger"
	"go.uber.org/zap"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	// Kubeconfig file path (optional, will auto-detect if not provided)
	KubeconfigPath string `envconfig:"KUBECONFIG" default:""`

	// Kubernetes context to use (optional, uses current context if not provided)
	Context string `envconfig:"CONTEXT" default:""`

	// Master URL (optional, for in-cluster config override)
	MasterURL string `envconfig:"MASTER_URL" default:""`

	// Default namespace for operations
	DefaultNamespace string `envconfig:"DEFAULT_NAMESPACE" default:"default"`

	// Enable in-cluster configuration
	InCluster bool `envconfig:"IN_CLUSTER" default:"false"`

	// Connection timeout in seconds
	TimeoutSeconds int `envconfig:"TIMEOUT_SECONDS" default:"30"`
}

func main() {
	// Initialize controller manager
	logger, err := logger.New("debug", "development")
	if err != nil {
		panic(fmt.Errorf("Failed to initialize logger: %v", err))
	}

	slog.Info("HERE")

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		panic(fmt.Errorf("failed to parse configuration: %w", err))
	}

	// Initialize Kubernetes client
	k8sClientOptions := &k8s.ClientOptions{
		KubeconfigPath: cfg.KubeconfigPath,
		Context:        cfg.Context,
		MasterURL:      cfg.MasterURL,
	}

	ctx, cf := signal.NotifyContext(context.TODO(), syscall.SIGINT, syscall.SIGTERM)
	defer cf()

	k8sClient, err := k8s.NewClient(ctx, k8sClientOptions)
	if err != nil {
		logger.Fatal("Failed to create Kubernetes client", zap.Error(err))
	}

	controllerManager, err := controllers.NewManager(k8sClient.Config, logger)
	if err != nil {
		logger.Fatal("Failed to create controller manager", zap.Error(err))
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error("Controller manager panicked",
				zap.Any("panic", r),
				zap.Stack("stack"))
			// Optionally: could restart controller manager here
		}
	}()

	logger.Info("Starting controller manager")
	if err := controllerManager.Start(ctx); err != nil {
		// Only log error if it's not due to context cancellation
		if ctx.Err() == nil {
			logger.Error("Controller manager stopped with error", zap.Error(err))
		} else {
			logger.Info("Controller manager stopped gracefully")
		}
	}
}
