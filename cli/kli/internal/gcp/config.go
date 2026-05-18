package gcp

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

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

// GCPConfig holds GCP configuration
type GCPConfig struct {
	Project string
	Region  string
	Zone    string
}

// LoadGCPConfig loads GCP configuration from environment/flags
func LoadGCPConfig(ctx context.Context, project, region, zone string) (*GCPConfig, error) {
	cfg := &GCPConfig{
		Project: project,
		Region:  region,
		Zone:    zone,
	}

	// Get project from environment if not specified
	if cfg.Project == "" {
		var err error
		cfg.Project, err = GetDefaultProject(ctx)
		if err != nil {
			return nil, fmt.Errorf("project not specified and could not determine default: %w", err)
		}
	}

	// Get region from gcloud config if not specified
	if cfg.Region == "" {
		var err error
		cfg.Region, err = GetDefaultRegion(ctx)
		if err != nil {
			return nil, fmt.Errorf("region not specified and could not determine default: %w", err)
		}
	}

	// Auto-select zone if not specified
	if cfg.Zone == "" {
		var err error
		cfg.Zone, err = SelectZoneFromRegion(ctx, cfg.Project, cfg.Region)
		if err != nil {
			return nil, fmt.Errorf("failed to select zone from region %s: %w", cfg.Region, err)
		}
	}

	return cfg, nil
}

// GetDefaultProject returns project from environment variables, gcloud config, or GCE metadata
func GetDefaultProject(ctx context.Context) (string, error) {
	// Try environment variables first (in order of precedence)
	envVars := []string{"GOOGLE_CLOUD_PROJECT", "GCLOUD_PROJECT", "GCP_PROJECT", "CLOUDSDK_CORE_PROJECT"}
	for _, envVar := range envVars {
		if project := os.Getenv(envVar); project != "" {
			return project, nil
		}
	}

	// Try reading from gcloud config files
	project := readGCloudConfig("core", "project")
	if project != "" {
		return project, nil
	}

	// Try GCE metadata service
	if project := getGCEMetadata("project/project-id"); project != "" {
		return project, nil
	}

	return "", fmt.Errorf("please set GOOGLE_CLOUD_PROJECT environment variable")
}

// GetDefaultRegion returns region from environment variables, gcloud config, or GCE metadata
func GetDefaultRegion(ctx context.Context) (string, error) {
	// Try environment variables first
	envVars := []string{"CLOUDSDK_COMPUTE_REGION", "GOOGLE_CLOUD_REGION", "GCP_REGION"}
	for _, envVar := range envVars {
		if region := os.Getenv(envVar); region != "" {
			return region, nil
		}
	}

	// Try reading from gcloud config files
	region := readGCloudConfig("compute", "region")
	if region != "" {
		return region, nil
	}

	// Try GCE metadata service (get zone and extract region)
	if zone := getGCEMetadata("instance/zone"); zone != "" {
		// Zone format: projects/123456/zones/us-central1-a
		// Extract region (us-central1) from zone
		parts := strings.Split(zone, "/")
		if len(parts) > 0 {
			zoneName := parts[len(parts)-1]
			// Remove the last part after hyphen (e.g., us-central1-a -> us-central1)
			lastHyphen := strings.LastIndex(zoneName, "-")
			if lastHyphen > 0 {
				return zoneName[:lastHyphen], nil
			}
		}
	}

	return "", fmt.Errorf("please set CLOUDSDK_COMPUTE_REGION environment variable or use --region flag")
}

// getGCEMetadata gets a value from GCE instance metadata service
func getGCEMetadata(path string) string {
	client := &http.Client{Timeout: 1 * time.Second}

	req, err := http.NewRequest("GET", "http://metadata.google.internal/computeMetadata/v1/"+path, nil)
	if err != nil {
		return ""
	}
	req.Header.Set("Metadata-Flavor", "Google")

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

// readGCloudConfig reads a value from gcloud configuration files
// It checks ~/.config/gcloud/properties and active configuration
func readGCloudConfig(section, key string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	gcloudDir := filepath.Join(homeDir, ".config", "gcloud")

	// First, try the properties file (legacy format)
	propertiesPath := filepath.Join(gcloudDir, "properties")
	if value := readINIValue(propertiesPath, section, key); value != "" {
		return value
	}

	// Then, try the active configuration
	activeConfig := readActiveConfig(gcloudDir)
	if activeConfig != "" {
		configPath := filepath.Join(gcloudDir, "configurations", "config_"+activeConfig)
		if value := readINIValue(configPath, section, key); value != "" {
			return value
		}
	}

	// Finally, try the default configuration
	defaultConfigPath := filepath.Join(gcloudDir, "configurations", "config_default")
	return readINIValue(defaultConfigPath, section, key)
}

// readActiveConfig reads the active configuration name from gcloud
func readActiveConfig(gcloudDir string) string {
	activeConfigPath := filepath.Join(gcloudDir, "active_config")
	data, err := os.ReadFile(activeConfigPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// readINIValue reads a value from an INI-style config file
func readINIValue(filePath, section, key string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inTargetSection := false
	targetSection := "[" + section + "]"

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

// SelectZoneFromRegion selects an available zone in the region
func SelectZoneFromRegion(ctx context.Context, project, region string) (string, error) {
	// First try to get zone from gcloud config
	if zone := readGCloudConfig("compute", "zone"); zone != "" {
		// Verify zone is in the requested region
		if strings.HasPrefix(zone, region+"-") {
			return zone, nil
		}
	}

	// Try to use GCP API to find an available zone
	zonesClient, err := compute.NewZonesRESTClient(ctx)
	if err != nil {
		// Fall back to default zone pattern if API is unavailable
		return region + "-a", nil
	}
	defer zonesClient.Close()

	req := &computepb.ListZonesRequest{
		Project: project,
	}

	it := zonesClient.List(ctx, req)
	for {
		zone, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// Fall back to default zone pattern if API call fails
			return region + "-a", nil
		}

		// Check if zone is in the requested region and is up
		if zone.Region != nil && getResourceName(*zone.Region) == region {
			if zone.Status != nil && *zone.Status == "UP" {
				return *zone.Name, nil
			}
		}
	}

	// Fall back to default zone pattern if no zones found
	return region + "-a", nil
}

// getResourceName extracts the resource name from a full resource URL
// e.g., "https://www.googleapis.com/compute/v1/projects/my-project/regions/us-central1" -> "us-central1"
func getResourceName(url string) string {
	for i := len(url) - 1; i >= 0; i-- {
		if url[i] == '/' {
			return url[i+1:]
		}
	}
	return url
}

// GetZoneURL returns the full URL for a zone
func GetZoneURL(project, zone string) string {
	return fmt.Sprintf("projects/%s/zones/%s", project, zone)
}

// GetRegionURL returns the full URL for a region
func GetRegionURL(project, region string) string {
	return fmt.Sprintf("projects/%s/regions/%s", project, region)
}

// GetNetworkURL returns the full URL for a network
func GetNetworkURL(project, network string) string {
	return fmt.Sprintf("projects/%s/global/networks/%s", project, network)
}

// GetSubnetworkURL returns the full URL for a subnetwork
func GetSubnetworkURL(project, region, subnetwork string) string {
	return fmt.Sprintf("projects/%s/regions/%s/subnetworks/%s", project, region, subnetwork)
}
