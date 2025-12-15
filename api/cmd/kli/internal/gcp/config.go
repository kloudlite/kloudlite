package gcp

import (
	"context"
	"fmt"
	"os"

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

	// Region is required
	if cfg.Region == "" {
		return nil, fmt.Errorf("region is required")
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

// GetDefaultProject returns project from GOOGLE_CLOUD_PROJECT or similar env vars
func GetDefaultProject(ctx context.Context) (string, error) {
	// Try environment variables (in order of precedence)
	envVars := []string{"GOOGLE_CLOUD_PROJECT", "GCLOUD_PROJECT", "GCP_PROJECT", "CLOUDSDK_CORE_PROJECT"}
	for _, envVar := range envVars {
		if project := os.Getenv(envVar); project != "" {
			return project, nil
		}
	}

	return "", fmt.Errorf("please set GOOGLE_CLOUD_PROJECT environment variable with your project ID")
}

// SelectZoneFromRegion selects an available zone in the region
func SelectZoneFromRegion(ctx context.Context, project, region string) (string, error) {
	zonesClient, err := compute.NewZonesRESTClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create zones client: %w", err)
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
			return "", fmt.Errorf("failed to list zones: %w", err)
		}

		// Check if zone is in the requested region and is up
		if zone.Region != nil && getResourceName(*zone.Region) == region {
			if zone.Status != nil && *zone.Status == "UP" {
				return *zone.Name, nil
			}
		}
	}

	return "", fmt.Errorf("no available zones found in region %s", region)
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
