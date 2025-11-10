package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config represents the kltun configuration
type Config struct {
	Token  string `yaml:"token"`
	Server string `yaml:"server"`
}

var (
	configMutex sync.RWMutex
	configPath  string
)

func init() {
	// Get config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory
		home, err := os.UserHomeDir()
		if err != nil {
			configDir = "."
		} else {
			configDir = filepath.Join(home, ".config")
		}
	}

	configPath = filepath.Join(configDir, "kltun", "config.yaml")
}

// GetConfigPath returns the path to the config file
func GetConfigPath() string {
	return configPath
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	configMutex.RLock()
	defer configMutex.RUnlock()

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

// Save saves the configuration to disk
func Save(cfg *Config) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	// Create config directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to temp file first
	tmpFile := configPath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tmpFile, configPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to save config file: %w", err)
	}

	return nil
}

// Update updates specific fields in the config
func Update(token, server string) error {
	cfg, err := Load()
	if err != nil {
		return err
	}

	if token != "" {
		cfg.Token = token
	}
	if server != "" {
		cfg.Server = server
	}

	return Save(cfg)
}
