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
// DISABLED: Configuration persistence has been disabled for security reasons
func Save(cfg *Config) error {
	return fmt.Errorf("configuration persistence is disabled - credentials are not saved to disk")
}

// Update updates specific fields in the config
// DISABLED: Configuration persistence has been disabled for security reasons
func Update(token, server string) error {
	return fmt.Errorf("configuration persistence is disabled - credentials are not saved to disk")
}
