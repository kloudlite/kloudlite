package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Port        string `envconfig:"PORT" default:"8080"`
	Environment string `envconfig:"ENVIRONMENT" default:"development"`
	LogLevel    string `envconfig:"LOG_LEVEL" default:"info"`

	// JWT Authentication configuration
	Auth AuthConfig `envconfig:"AUTH"`

	// Kubernetes configuration
	Kubernetes KubernetesConfig `envconfig:"KUBERNETES"`

	// Connection Token configuration
	ConnectionToken ConnectionTokenConfig `envconfig:"CONNECTION_TOKEN"`

	// Installation configuration
	Installation InstallationConfig `envconfig:"INSTALLATION"`
}

type AuthConfig struct {
	// JWT secret for token verification (REQUIRED - must be set via environment variable)
	JWTSecret string `envconfig:"JWT_SECRET" required:"true"`

	// Token expiry duration in hours
	TokenExpiryHours int `envconfig:"TOKEN_EXPIRY_HOURS" default:"24"`

	// Skip authentication for development/testing
	SkipAuthentication bool `envconfig:"SKIP_AUTHENTICATION" default:"false"`
}

type KubernetesConfig struct {
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

type ConnectionTokenConfig struct {
	// SSH Jump Host for workspace connections
	SSHJumpHost string `envconfig:"SSH_JUMP_HOST" default:"localhost"`

	// SSH Port for jump host
	SSHPort int `envconfig:"SSH_PORT" default:"2222"`

	// API URL for Kloudlite API
	APIURL string `envconfig:"API_URL" default:"http://localhost:8080"`
}

type InstallationConfig struct {
	// InstallationKey is the unique key for this Kloudlite installation
	InstallationKey string `envconfig:"KEY"`

	// InstallationSecret is the secret key for authentication
	InstallationSecret string `envconfig:"SECRET"`

	// ConsoleURL is the URL of the console web application
	ConsoleURL string `envconfig:"CONSOLE_URL" default:"https://console.kloudlite.io"`

	// PublicIP is the public IP address for the installation (from INSTALLATION_AWS_PUBLIC_IP env var)
	PublicIP string `envconfig:"INSTALLATION_AWS_PUBLIC_IP"`

	// PollingIntervalSeconds is the interval to poll for subdomain configuration
	PollingIntervalSeconds int `envconfig:"POLLING_INTERVAL_SECONDS" default:"30"`
}

func Load() (*Config, error) {
	// Load .env file if exists (ignore error if not found)
	_ = godotenv.Load()

	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	return &cfg, nil
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// ToK8sClientOptions converts the config to k8s client options
func (c *Config) ToK8sClientOptions() *K8sClientOptions {
	return &K8sClientOptions{
		KubeconfigPath: c.Kubernetes.KubeconfigPath,
		Context:        c.Kubernetes.Context,
		MasterURL:      c.Kubernetes.MasterURL,
	}
}

// K8sClientOptions represents options for creating a Kubernetes client
// This is a simplified version that matches what the k8s package expects
type K8sClientOptions struct {
	KubeconfigPath string
	Context        string
	MasterURL      string
}
