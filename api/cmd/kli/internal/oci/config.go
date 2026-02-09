package oci

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/common/auth"
)

// OCIConfig holds OCI configuration
type OCIConfig struct {
	TenancyOCID     string
	UserOCID        string
	Region          string
	CompartmentOCID string
	Fingerprint     string
	KeyFile         string
	KeyContent      string // PEM private key content (alternative to KeyFile)
	PassPhrase      string
	ConfigProvider  common.ConfigurationProvider
}

// LoadOCIConfig loads OCI configuration from flags/env/config file
func LoadOCIConfig(ctx context.Context, tenancy, user, region, compartment, fingerprint, keyFile string) (*OCIConfig, error) {
	cfg := &OCIConfig{
		TenancyOCID:     tenancy,
		UserOCID:        user,
		Region:          region,
		CompartmentOCID: compartment,
		Fingerprint:     fingerprint,
		KeyFile:         keyFile,
	}

	// Try environment variables for any missing values
	if cfg.TenancyOCID == "" {
		cfg.TenancyOCID = os.Getenv("OCI_CLI_TENANCY")
	}
	if cfg.UserOCID == "" {
		cfg.UserOCID = os.Getenv("OCI_CLI_USER")
	}
	if cfg.Region == "" {
		cfg.Region = os.Getenv("OCI_CLI_REGION")
	}
	if cfg.CompartmentOCID == "" {
		cfg.CompartmentOCID = os.Getenv("OCI_CLI_COMPARTMENT")
	}
	if cfg.Fingerprint == "" {
		cfg.Fingerprint = os.Getenv("OCI_CLI_FINGERPRINT")
	}
	if cfg.KeyContent == "" {
		cfg.KeyContent = os.Getenv("OCI_CLI_KEY_CONTENT")
	}
	if cfg.KeyFile == "" {
		cfg.KeyFile = os.Getenv("OCI_CLI_KEY_FILE")
	}

	// Try loading from OCI config file (~/.oci/config)
	if cfg.TenancyOCID == "" || cfg.UserOCID == "" || cfg.Region == "" {
		fileCfg := readOCIConfigFile("")
		if cfg.TenancyOCID == "" {
			cfg.TenancyOCID = fileCfg["tenancy"]
		}
		if cfg.UserOCID == "" {
			cfg.UserOCID = fileCfg["user"]
		}
		if cfg.Region == "" {
			cfg.Region = fileCfg["region"]
		}
		if cfg.Fingerprint == "" {
			cfg.Fingerprint = fileCfg["fingerprint"]
		}
		if cfg.KeyFile == "" {
			cfg.KeyFile = fileCfg["key_file"]
		}
		if cfg.PassPhrase == "" {
			cfg.PassPhrase = fileCfg["pass_phrase"]
		}
	}

	// Try instance metadata (IMDS) for region if still missing
	if cfg.Region == "" {
		if region := getOCIMetadata("instance/region"); region != "" {
			cfg.Region = region
		}
	}

	// Use default config provider from OCI SDK (handles full chain)
	configProvider, err := getConfigProvider(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OCI config provider: %w", err)
	}
	cfg.ConfigProvider = configProvider

	// Resolve remaining values from the config provider
	if cfg.TenancyOCID == "" {
		t, err := configProvider.TenancyOCID()
		if err == nil {
			cfg.TenancyOCID = t
		}
	}
	if cfg.Region == "" {
		r, err := configProvider.Region()
		if err == nil {
			cfg.Region = r
		}
	}

	// Validate required fields
	if cfg.TenancyOCID == "" {
		return nil, fmt.Errorf("tenancy OCID not specified: set OCI_CLI_TENANCY env var, use --tenancy flag, or configure ~/.oci/config")
	}
	if cfg.Region == "" {
		return nil, fmt.Errorf("region not specified: set OCI_CLI_REGION env var, use --region flag, or configure ~/.oci/config")
	}

	// Default compartment to tenancy if not specified
	if cfg.CompartmentOCID == "" {
		cfg.CompartmentOCID = cfg.TenancyOCID
	}

	return cfg, nil
}

// getConfigProvider returns an OCI config provider
func getConfigProvider(cfg *OCIConfig) (common.ConfigurationProvider, error) {
	// If we have all the explicit values, use them
	if cfg.TenancyOCID != "" && cfg.UserOCID != "" && cfg.Region != "" && cfg.Fingerprint != "" && (cfg.KeyContent != "" || cfg.KeyFile != "") {
		var keyPEM string

		if cfg.KeyContent != "" {
			// Use key content directly (e.g., from OCI_CLI_KEY_CONTENT env var)
			// Handle escaped newlines from env vars (e.g. "-----BEGIN...\\nMII..." → real newlines)
			keyPEM = strings.ReplaceAll(cfg.KeyContent, `\n`, "\n")
		} else {
			// Expand ~ in key file path
			keyFile := cfg.KeyFile
			if strings.HasPrefix(keyFile, "~") {
				home, err := os.UserHomeDir()
				if err == nil {
					keyFile = filepath.Join(home, keyFile[1:])
				}
			}

			// Read the private key file content
			keyBytes, err := os.ReadFile(keyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read key file %s: %w", keyFile, err)
			}
			keyPEM = string(keyBytes)
		}

		var passphrase *string
		if cfg.PassPhrase != "" {
			passphrase = &cfg.PassPhrase
		}

		return common.NewRawConfigurationProvider(
			cfg.TenancyOCID,
			cfg.UserOCID,
			cfg.Region,
			cfg.Fingerprint,
			keyPEM,
			passphrase,
		), nil
	}

	// Try the default config provider (reads ~/.oci/config)
	provider := common.DefaultConfigProvider()
	// Verify it works by trying to get tenancy
	_, err := provider.TenancyOCID()
	if err == nil {
		return provider, nil
	}

	// Try instance principal (for running on OCI instances)
	instanceProvider, ipErr := auth.InstancePrincipalConfigurationProvider()
	if ipErr == nil {
		return instanceProvider, nil
	}

	return nil, fmt.Errorf("could not create config provider: default config error: %w, instance principal error: %v", err, ipErr)
}

// getOCIMetadata gets a value from OCI Instance Metadata Service (IMDS)
func getOCIMetadata(path string) string {
	client := &http.Client{Timeout: 1 * time.Second}

	req, err := http.NewRequest("GET", "http://169.254.169.254/opc/v2/"+path, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Authorization", "Bearer Oracle")

	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(body))
}

// readOCIConfigFile reads OCI config from ~/.oci/config
// profile can be empty for DEFAULT profile
func readOCIConfigFile(profile string) map[string]string {
	result := make(map[string]string)

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return result
	}

	configPath := filepath.Join(homeDir, ".oci", "config")
	return readOCIINI(configPath, profile)
}

// readOCIINI reads key-value pairs from an OCI INI-style config file
func readOCIINI(filePath, profile string) map[string]string {
	result := make(map[string]string)

	file, err := os.Open(filePath)
	if err != nil {
		return result
	}
	defer file.Close()

	if profile == "" {
		profile = "DEFAULT"
	}
	targetSection := "[" + profile + "]"

	scanner := bufio.NewScanner(file)
	inTargetSection := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Check for section headers
		if strings.HasPrefix(line, "[") {
			inTargetSection = strings.EqualFold(line, targetSection)
			continue
		}

		// Look for key=value in target section
		if inTargetSection && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				result[key] = value
			}
		}
	}

	return result
}

// GetOCIConfigPath returns the path to the OCI config file
func GetOCIConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".oci", "config")
}

// OCIConfigExists checks if the OCI config file exists
func OCIConfigExists() bool {
	configPath := GetOCIConfigPath()
	if configPath == "" {
		return false
	}
	_, err := os.Stat(configPath)
	return err == nil
}
