package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	ServerURL    string `json:"server_url"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserID       string `json:"user_id"`
	ServerAddr   string `json:"server_addr"` // gRPC server address
}

var (
	defaultConfigDir  = filepath.Join(os.Getenv("HOME"), ".kloudlite")
	defaultConfigFile = "config.json"
)

func DefaultConfigPath() string {
	return filepath.Join(defaultConfigDir, defaultConfigFile)
}

func EnsureConfigDir() error {
	return os.MkdirAll(defaultConfigDir, 0700)
}

func Load() (*Config, error) {
	if err := EnsureConfigDir(); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := DefaultConfigPath()
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				ServerURL: "https://api.kloudlite.io",
			}, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	if err := EnsureConfigDir(); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath := DefaultConfigPath()
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) IsAuthenticated() bool {
	return c.AccessToken != ""
}

func (c *Config) ClearAuth() {
	c.AccessToken = ""
	c.RefreshToken = ""
}

func (c *Config) SetAuth(accessToken, refreshToken string) {
	c.AccessToken = accessToken
	c.RefreshToken = refreshToken
}

// Save is a static function to save config
func Save(cfg *Config) error {
	return cfg.Save()
}

// Clear removes the config file
func Clear() error {
	configPath := DefaultConfigPath()
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove config file: %w", err)
	}
	return nil
}

// IsLoggedIn checks if user is logged in
func IsLoggedIn() (bool, error) {
	cfg, err := Load()
	if err != nil {
		return false, err
	}
	return cfg.IsAuthenticated(), nil
}