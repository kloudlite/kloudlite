package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Config holds the environment-based configuration for the OCI installer job.
type Config struct {
	Operation                string // "install" or "uninstall"
	InstallationKey          string
	ConsoleBaseURL           string
	OCITenancy               string
	OCIUser                  string
	OCIRegion                string
	OCICompartment           string
	OCIFingerprint           string
	OCIKeyContent            string
	SkipLB                   bool
	EnableDeletionProtection bool
}

func loadConfig() (*Config, error) {
	cfg := &Config{
		Operation:       os.Getenv("OPERATION"),
		InstallationKey: os.Getenv("INSTALLATION_KEY"),
		ConsoleBaseURL:  os.Getenv("CONSOLE_BASE_URL"),
		OCITenancy:      os.Getenv("OCI_CLI_TENANCY"),
		OCIUser:         os.Getenv("OCI_CLI_USER"),
		OCIRegion:       os.Getenv("OCI_CLI_REGION"),
		OCICompartment:  os.Getenv("OCI_CLI_COMPARTMENT"),
		OCIFingerprint:  os.Getenv("OCI_CLI_FINGERPRINT"),
		OCIKeyContent:   os.Getenv("OCI_CLI_KEY_CONTENT"),
	}

	if cfg.ConsoleBaseURL == "" {
		cfg.ConsoleBaseURL = "https://console.kloudlite.io"
	}

	cfg.SkipLB, _ = strconv.ParseBool(os.Getenv("SKIP_LB"))

	enableDeletion := os.Getenv("ENABLE_DELETION_PROTECTION")
	if enableDeletion == "" {
		cfg.EnableDeletionProtection = true
	} else {
		cfg.EnableDeletionProtection, _ = strconv.ParseBool(enableDeletion)
	}

	// Validate required fields
	cfg.Operation = strings.ToLower(cfg.Operation)
	if cfg.Operation != "install" && cfg.Operation != "uninstall" {
		return nil, fmt.Errorf("OPERATION must be 'install' or 'uninstall', got %q", cfg.Operation)
	}
	if cfg.InstallationKey == "" {
		return nil, fmt.Errorf("INSTALLATION_KEY is required")
	}
	if cfg.OCITenancy == "" {
		return nil, fmt.Errorf("OCI_CLI_TENANCY is required")
	}
	if cfg.OCIUser == "" {
		return nil, fmt.Errorf("OCI_CLI_USER is required")
	}
	if cfg.OCIRegion == "" {
		return nil, fmt.Errorf("OCI_CLI_REGION is required")
	}
	if cfg.OCIFingerprint == "" {
		return nil, fmt.Errorf("OCI_CLI_FINGERPRINT is required")
	}
	if cfg.OCIKeyContent == "" {
		return nil, fmt.Errorf("OCI_CLI_KEY_CONTENT is required")
	}

	return cfg, nil
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)

	log.Println("=== Kloudlite OCI Installer Job ===")
	startTime := time.Now()

	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	log.Printf("Operation:        %s", cfg.Operation)
	log.Printf("Installation Key: %s", cfg.InstallationKey)
	log.Printf("OCI Region:       %s", cfg.OCIRegion)
	log.Printf("Console URL:      %s", cfg.ConsoleBaseURL)

	// 25-minute timeout for the overall operation
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Minute)
	defer cancel()

	// Handle SIGTERM for graceful shutdown (ACA sends SIGTERM before killing)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigChan
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
		cancel()
	}()

	switch cfg.Operation {
	case "install":
		if err := runInstall(ctx, cfg); err != nil {
			log.Printf("FAILED: Installation failed after %s: %v", time.Since(startTime).Truncate(time.Second), err)
			os.Exit(1)
		}
	case "uninstall":
		if err := runUninstall(ctx, cfg); err != nil {
			log.Printf("FAILED: Uninstallation failed after %s: %v", time.Since(startTime).Truncate(time.Second), err)
			os.Exit(1)
		}
	}

	log.Printf("=== Completed successfully in %s ===", time.Since(startTime).Truncate(time.Second))
}
