package aws

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func LoadAWSConfig(ctx context.Context, region, profile string) (aws.Config, error) {
	var opts []func(*config.LoadOptions) error

	// Get region from env var or config file if not specified
	if region == "" {
		region = GetDefaultRegion(profile)
	}

	if region != "" {
		opts = append(opts, config.WithRegion(region))
	}
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	return config.LoadDefaultConfig(ctx, opts...)
}

// GetDefaultRegion returns region from environment variables, ~/.aws/config, or EC2 instance metadata
func GetDefaultRegion(profile string) string {
	// Try environment variables first (in order of precedence)
	envVars := []string{"AWS_REGION", "AWS_DEFAULT_REGION"}
	for _, envVar := range envVars {
		if region := os.Getenv(envVar); region != "" {
			return region
		}
	}

	// Try reading from ~/.aws/config
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".aws", "config")
		region := readAWSConfigValue(configPath, profile, "region")
		if region != "" {
			return region
		}
	}

	// Try EC2 instance metadata service (IMDSv2)
	if region := getRegionFromEC2Metadata(); region != "" {
		return region
	}

	return ""
}

// getRegionFromEC2Metadata gets the region from EC2 instance metadata service (IMDSv2)
func getRegionFromEC2Metadata() string {
	client := &http.Client{Timeout: 1 * time.Second}

	// First, get IMDSv2 token
	tokenReq, err := http.NewRequest("PUT", "http://169.254.169.254/latest/api/token", nil)
	if err != nil {
		return ""
	}
	tokenReq.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "21600")

	tokenResp, err := client.Do(tokenReq)
	if err != nil {
		return ""
	}
	defer tokenResp.Body.Close()

	token, err := io.ReadAll(tokenResp.Body)
	if err != nil {
		return ""
	}

	// Get availability zone
	azReq, err := http.NewRequest("GET", "http://169.254.169.254/latest/meta-data/placement/availability-zone", nil)
	if err != nil {
		return ""
	}
	azReq.Header.Set("X-aws-ec2-metadata-token", string(token))

	azResp, err := client.Do(azReq)
	if err != nil {
		return ""
	}
	defer azResp.Body.Close()

	az, err := io.ReadAll(azResp.Body)
	if err != nil || len(az) == 0 {
		return ""
	}

	// Extract region from availability zone (e.g., us-east-1a -> us-east-1)
	azStr := strings.TrimSpace(string(az))
	if len(azStr) > 0 {
		// Remove the last character (availability zone letter)
		return azStr[:len(azStr)-1]
	}
	return ""
}

// readAWSConfigValue reads a value from AWS config file for a given profile
func readAWSConfigValue(configPath, profile, key string) string {
	file, err := os.Open(configPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	if profile == "" {
		profile = "default"
	}

	// AWS config uses [profile xxx] for non-default profiles, [default] for default
	targetSection := "[default]"
	if profile != "default" {
		targetSection = "[profile " + profile + "]"
	}

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
			if len(parts) == 2 && strings.TrimSpace(parts[0]) == key {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}
