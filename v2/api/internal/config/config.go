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

	// Kubernetes configuration
	Kubernetes KubernetesConfig `envconfig:"KUBERNETES"`
}

type KubernetesConfig struct {
	// Kubeconfig file path (optional, will auto-detect if not provided)
	KubeconfigPath string `envconfig:"KUBECONFIG_PATH" default:""`

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