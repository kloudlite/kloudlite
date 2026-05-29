package gcp

import (
	"context"
	"fmt"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

// Ubuntu2404ImageFamily is the image family URL for Ubuntu 24.04 LTS
// Using the family URL allows GCP to automatically select the latest image
const Ubuntu2404ImageFamily = "projects/ubuntu-os-cloud/global/images/family/ubuntu-2404-lts-amd64"

// FindUbuntuImage finds the latest Ubuntu 24.04 LTS image
// Returns the full image URL for use in instance creation
func FindUbuntuImage(ctx context.Context, cfg *GCPConfig) (string, error) {
	// Try to get specific image via API for better logging
	imagesClient, err := compute.NewImagesRESTClient(ctx)
	if err != nil {
		// Fall back to image family URL (GCP will resolve to latest)
		return Ubuntu2404ImageFamily, nil
	}
	defer imagesClient.Close()

	// Ubuntu images are in the ubuntu-os-cloud project
	// Ubuntu 24.04 LTS = Noble Numbat
	req := &computepb.ListImagesRequest{
		Project: "ubuntu-os-cloud",
		Filter:  ptrString("name:ubuntu-2404-noble-amd64-*"),
	}

	var latestImage *computepb.Image
	it := imagesClient.List(ctx, req)
	for {
		image, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// Fall back to image family URL
			return Ubuntu2404ImageFamily, nil
		}

		// Skip deprecated images
		if image.Deprecated != nil && image.Deprecated.State != nil && *image.Deprecated.State != "" {
			continue
		}

		// Find the most recent image
		if latestImage == nil || (image.CreationTimestamp != nil && latestImage.CreationTimestamp != nil &&
			*image.CreationTimestamp > *latestImage.CreationTimestamp) {
			latestImage = image
		}
	}

	if latestImage == nil {
		// Fall back to image family URL
		return Ubuntu2404ImageFamily, nil
	}

	// Return the full image URL
	return *latestImage.SelfLink, nil
}

// GetImageFamily returns the image from a specific family
// This is an alternative way to get the latest image
func GetImageFamily(ctx context.Context, project, family string) (string, error) {
	imagesClient, err := compute.NewImagesRESTClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create images client: %w", err)
	}
	defer imagesClient.Close()

	req := &computepb.GetFromFamilyImageRequest{
		Project: project,
		Family:  family,
	}

	image, err := imagesClient.GetFromFamily(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to get image from family %s: %w", family, err)
	}

	return *image.SelfLink, nil
}

// GetUbuntu2404Image returns the latest Ubuntu 24.04 LTS image
// Uses the image family URL which GCP will resolve to the latest image
func GetUbuntu2404Image(ctx context.Context) (string, error) {
	// Return the image family URL - GCP will automatically resolve to latest
	// This works without requiring API credentials
	return Ubuntu2404ImageFamily, nil
}

// GetImageName extracts the image name from the full URL
func GetImageName(imageURL string) string {
	parts := strings.Split(imageURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return imageURL
}

func ptrString(s string) *string {
	return &s
}
