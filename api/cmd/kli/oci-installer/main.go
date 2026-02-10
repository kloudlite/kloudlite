package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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

// acquireLock tries to acquire a job lock via the console API.
// Returns true if lock acquired, false if another job is already running.
func acquireLock(cfg *Config) (bool, error) {
	return callJobLock(cfg, "lock")
}

// releaseLock releases the job lock via the console API.
func releaseLock(cfg *Config, failed bool) {
	status := "succeeded"
	if failed {
		status = "failed"
	}
	ok, err := callJobLock(cfg, "unlock", status)
	if err != nil {
		log.Printf("Warning: failed to release lock: %v", err)
	} else if !ok {
		log.Printf("Warning: lock release returned unexpected result")
	}
}

func callJobLock(cfg *Config, action string, status ...string) (bool, error) {
	url := fmt.Sprintf("%s/api/installations/job-lock", cfg.ConsoleBaseURL)

	payload := map[string]string{
		"installationKey": cfg.InstallationKey,
		"action":          action,
	}
	if len(status) > 0 {
		payload["status"] = status[0]
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("failed to call job-lock API: %w", err)
	}
	defer resp.Body.Close()

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("failed to decode job-lock response: %w", err)
	}

	if action == "lock" {
		acquired, _ := result["acquired"].(bool)
		return acquired, nil
	}

	released, _ := result["released"].(bool)
	return released, nil
}

// reportProgress sends step progress to the console API (fire-and-forget).
func reportProgress(cfg *Config, operation string, currentStep, totalSteps int, stepDescription string) {
	sendProgress(cfg, operation, currentStep, totalSteps, stepDescription, false)
}

// reportProgressComplete marks the job as completed in the console API.
func reportProgressComplete(cfg *Config, operation string, totalSteps int, stepDescription string) {
	sendProgress(cfg, operation, totalSteps, totalSteps, stepDescription, true)
}

func sendProgress(cfg *Config, operation string, currentStep, totalSteps int, stepDescription string, completed bool) {
	url := fmt.Sprintf("%s/api/installations/job-progress", cfg.ConsoleBaseURL)

	payload := map[string]any{
		"installationKey": cfg.InstallationKey,
		"operation":       operation,
		"currentStep":     currentStep,
		"totalSteps":      totalSteps,
		"stepDescription": stepDescription,
		"completed":       completed,
	}
	body, _ := json.Marshal(payload)

	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("Warning: failed to report progress: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Warning: progress report returned status %d", resp.StatusCode)
	}
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.LUTC)
	os.Exit(run())
}

func run() int {
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

	// Try to acquire lock — log warning but continue if lock service is unavailable
	acquired, lockErr := acquireLock(cfg)
	if lockErr != nil {
		log.Printf("Warning: lock service unavailable: %v (continuing anyway)", lockErr)
	} else if !acquired {
		log.Println("Warning: lock not acquired (another job may be running). Continuing anyway.")
	} else {
		log.Println("Lock acquired")
	}
	failed := false
	defer func() {
		if acquired {
			releaseLock(cfg, failed)
		}
	}()

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
			failed = true
			return 1
		}
		reportProgressComplete(cfg, "install", 9, "Installation complete")
	case "uninstall":
		if err := runUninstall(ctx, cfg); err != nil {
			log.Printf("FAILED: Uninstallation failed after %s: %v", time.Since(startTime).Truncate(time.Second), err)
			failed = true
			return 1
		}
		reportProgressComplete(cfg, "uninstall", 4, "Uninstallation complete")
	}

	log.Printf("=== Completed successfully in %s ===", time.Since(startTime).Truncate(time.Second))
	return 0
}
