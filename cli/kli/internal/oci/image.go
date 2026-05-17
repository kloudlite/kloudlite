package oci

import (
	"context"
	"fmt"
	"strings"

	"github.com/oracle/oci-go-sdk/v65/core"
)

// FindUbuntuImage finds the latest Ubuntu 24.04 image for the given compartment
func FindUbuntuImage(ctx context.Context, cfg *OCIConfig) (string, string, error) {
	computeClient, err := core.NewComputeClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", "", fmt.Errorf("failed to create compute client: %w", err)
	}

	// List platform images filtered by operating system
	shape := "VM.Standard.E4.Flex"
	resp, err := computeClient.ListImages(ctx, core.ListImagesRequest{
		CompartmentId:          &cfg.CompartmentOCID,
		OperatingSystem:        strPtr("Canonical Ubuntu"),
		OperatingSystemVersion: strPtr("24.04"),
		Shape:                  &shape,
		SortBy:                 core.ListImagesSortByTimecreated,
		SortOrder:              core.ListImagesSortOrderDesc,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to list images: %w", err)
	}

	// Find the latest amd64 image
	for _, img := range resp.Items {
		if img.Id == nil || img.DisplayName == nil {
			continue
		}
		// Skip images that are not for amd64
		displayName := *img.DisplayName
		if strings.Contains(strings.ToLower(displayName), "aarch64") {
			continue
		}
		return *img.Id, displayName, nil
	}

	// Fallback: try without shape filter
	resp2, err := computeClient.ListImages(ctx, core.ListImagesRequest{
		CompartmentId:          &cfg.CompartmentOCID,
		OperatingSystem:        strPtr("Canonical Ubuntu"),
		OperatingSystemVersion: strPtr("24.04"),
		SortBy:                 core.ListImagesSortByTimecreated,
		SortOrder:              core.ListImagesSortOrderDesc,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to list images (fallback): %w", err)
	}

	for _, img := range resp2.Items {
		if img.Id == nil || img.DisplayName == nil {
			continue
		}
		displayName := *img.DisplayName
		if strings.Contains(strings.ToLower(displayName), "aarch64") {
			continue
		}
		return *img.Id, displayName, nil
	}

	return "", "", fmt.Errorf("no Ubuntu 24.04 image found in compartment %s", cfg.CompartmentOCID)
}

// GetImageName extracts a display name from an image OCID
func GetImageName(imageID string) string {
	// Image OCIDs are like ocid1.image.oc1.iad.aaaa...
	// Just return the ID for logging
	return imageID
}

func strPtr(s string) *string {
	return &s
}
