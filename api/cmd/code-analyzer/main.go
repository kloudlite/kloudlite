package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/kloudlite/kloudlite/api/cmd/code-analyzer/analyzer"
	"github.com/kloudlite/kloudlite/api/cmd/code-analyzer/queue"
	"github.com/kloudlite/kloudlite/api/cmd/code-analyzer/server"
	"github.com/kloudlite/kloudlite/api/cmd/code-analyzer/storage"
	"github.com/kloudlite/kloudlite/api/cmd/code-analyzer/watcher"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds the application configuration
type Config struct {
	Namespace         string
	WorkmachineName   string
	ClaudeAPIURL      string
	ClaudeAPIKey      string
	WorkspacesPath    string
	ReportsPath       string
	DebounceSeconds   int
	MaxConcurrentJobs int
	HTTPPort          int
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func loadConfig() (*Config, error) {
	config := &Config{
		Namespace:         getEnv("NAMESPACE", ""),
		WorkmachineName:   getEnv("WORKMACHINE_NAME", ""),
		ClaudeAPIURL:      getEnv("CLAUDE_API_URL", "https://console.kloudlite.io/api/anthropic/v1/messages"),
		ClaudeAPIKey:      getEnv("CLAUDE_API_KEY", ""),
		WorkspacesPath:    getEnv("WORKSPACES_PATH", "/var/lib/kloudlite/home/workspaces"),
		ReportsPath:       getEnv("REPORTS_PATH", "/var/lib/kloudlite/code-analysis"),
		DebounceSeconds:   getEnvInt("DEBOUNCE_SECONDS", 45),
		MaxConcurrentJobs: getEnvInt("MAX_CONCURRENT_ANALYSES", 2),
		HTTPPort:          getEnvInt("HTTP_PORT", 8082),
	}

	// Validate required config
	if config.ClaudeAPIKey == "" {
		return nil, fmt.Errorf("CLAUDE_API_KEY is required")
	}

	return config, nil
}

func setupLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Use debug level if DEBUG env var is set
	if os.Getenv("DEBUG") != "" {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	return logger
}

func main() {
	logger := setupLogger()
	defer logger.Sync()

	logger.Info("Starting code-analyzer service")

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	logger.Info("Configuration loaded",
		zap.String("namespace", config.Namespace),
		zap.String("workmachine", config.WorkmachineName),
		zap.String("workspaces_path", config.WorkspacesPath),
		zap.String("reports_path", config.ReportsPath),
		zap.Int("debounce_seconds", config.DebounceSeconds),
		zap.Int("max_concurrent", config.MaxConcurrentJobs),
	)

	// Create storage
	reportStorage := storage.NewStorage(config.ReportsPath, logger)

	// Create Claude Code client
	claudeCode := analyzer.NewClaudeCode(
		config.ClaudeAPIURL,
		config.ClaudeAPIKey,
		config.WorkspacesPath,
		logger,
	)

	// Create analyzer
	codeAnalyzer := analyzer.NewAnalyzer(
		claudeCode,
		reportStorage,
		config.WorkspacesPath,
		logger,
	)

	// Create analysis queue
	analysisQueue := queue.NewQueue(
		config.MaxConcurrentJobs,
		func(ctx context.Context, workspaceName string) error {
			return codeAnalyzer.AnalyzeWorkspace(ctx, workspaceName)
		},
		logger,
	)

	// Create watcher manager
	debounceDuration := time.Duration(config.DebounceSeconds) * time.Second
	watcherManager := watcher.NewManager(
		config.WorkspacesPath,
		debounceDuration,
		func(workspaceName string) {
			analysisQueue.EnqueueAnalysis(workspaceName)
		},
		logger,
	)

	// Create HTTP server
	httpServer := server.NewServer(
		reportStorage,
		analysisQueue,
		watcherManager,
		logger,
	)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the analysis queue
	analysisQueue.Start()

	// Discover and watch existing workspaces
	go discoverAndWatchWorkspaces(ctx, config.WorkspacesPath, watcherManager, reportStorage, analysisQueue, logger)

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start HTTP server in background
	go func() {
		addr := fmt.Sprintf(":%d", config.HTTPPort)
		if err := httpServer.Start(addr); err != nil {
			logger.Error("HTTP server error", zap.Error(err))
		}
	}()

	logger.Info("Code analyzer service started",
		zap.Int("http_port", config.HTTPPort),
	)

	// Wait for shutdown signal
	sig := <-sigChan
	logger.Info("Received shutdown signal", zap.String("signal", sig.String()))

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop HTTP server
	if err := httpServer.Stop(shutdownCtx); err != nil {
		logger.Error("Error stopping HTTP server", zap.Error(err))
	}

	// Stop watchers
	watcherManager.StopAll()

	// Stop queue
	analysisQueue.Stop()

	logger.Info("Code analyzer service stopped")
}

// discoverAndWatchWorkspaces discovers existing workspaces and starts watching them
func discoverAndWatchWorkspaces(ctx context.Context, workspacesPath string, manager *watcher.Manager, reportStorage *storage.Storage, analysisQueue *queue.Queue, logger *zap.Logger) {
	// Wait a bit for the service to fully start
	time.Sleep(5 * time.Second)

	// List existing workspace directories
	entries, err := os.ReadDir(workspacesPath)
	if err != nil {
		logger.Warn("Failed to read workspaces directory", zap.Error(err))
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		workspaceName := entry.Name()
		workspacePath := filepath.Join(workspacesPath, workspaceName)

		// Check if it's a valid workspace (has files)
		files, err := os.ReadDir(workspacePath)
		if err != nil || len(files) == 0 {
			continue
		}

		// Start watching
		if err := manager.StartWatching(ctx, workspaceName); err != nil {
			logger.Warn("Failed to start watching workspace",
				zap.String("workspace", workspaceName),
				zap.Error(err),
			)
			continue
		}

		// Check if reports exist, if not trigger initial analysis
		securityReport, _ := reportStorage.GetLatestReport(workspaceName, storage.ReportTypeSecurity)
		qualityReport, _ := reportStorage.GetLatestReport(workspaceName, storage.ReportTypeQuality)

		if securityReport == nil || qualityReport == nil {
			logger.Info("No existing reports found, triggering initial analysis",
				zap.String("workspace", workspaceName),
			)
			analysisQueue.EnqueueAnalysis(workspaceName)
		}
	}

	// Continue monitoring for new workspaces
	go monitorForNewWorkspaces(ctx, workspacesPath, manager, reportStorage, analysisQueue, logger)
}

// monitorForNewWorkspaces periodically checks for new workspace directories
func monitorForNewWorkspaces(ctx context.Context, workspacesPath string, manager *watcher.Manager, reportStorage *storage.Storage, analysisQueue *queue.Queue, logger *zap.Logger) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			entries, err := os.ReadDir(workspacesPath)
			if err != nil {
				continue
			}

			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}

				workspaceName := entry.Name()
				if !manager.IsWatching(workspaceName) {
					workspacePath := filepath.Join(workspacesPath, workspaceName)
					files, err := os.ReadDir(workspacePath)
					if err != nil || len(files) == 0 {
						continue
					}

					if err := manager.StartWatching(ctx, workspaceName); err != nil {
						logger.Warn("Failed to start watching new workspace",
							zap.String("workspace", workspaceName),
							zap.Error(err),
						)
					} else {
						logger.Info("Started watching new workspace",
							zap.String("workspace", workspaceName),
						)

						// Trigger initial analysis for new workspace
						securityReport, _ := reportStorage.GetLatestReport(workspaceName, storage.ReportTypeSecurity)
						qualityReport, _ := reportStorage.GetLatestReport(workspaceName, storage.ReportTypeQuality)

						if securityReport == nil || qualityReport == nil {
							logger.Info("No existing reports found, triggering initial analysis",
								zap.String("workspace", workspaceName),
							)
							analysisQueue.EnqueueAnalysis(workspaceName)
						}
					}
				}
			}
		}
	}
}
