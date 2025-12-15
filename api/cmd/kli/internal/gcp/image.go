package gcp

import (
	"context"
	"fmt"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

// FindUbuntuImage finds the latest Ubuntu 24.04 LTS image
// Returns the full image URL for use in instance creation
func FindUbuntuImage(ctx context.Context, cfg *GCPConfig) (string, error) {
	imagesClient, err := compute.NewImagesRESTClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create images client: %w", err)
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
			return "", fmt.Errorf("failed to list images: %w", err)
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
		return "", fmt.Errorf("no Ubuntu 24.04 LTS image found")
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

// GetUbuntu2404Image returns the latest Ubuntu 24.04 LTS image using the family method
// This is the preferred method as it always returns the latest patched image
func GetUbuntu2404Image(ctx context.Context) (string, error) {
	// ubuntu-2404-lts is the family name for Ubuntu 24.04 Noble Numbat
	return GetImageFamily(ctx, "ubuntu-os-cloud", "ubuntu-2404-lts-amd64")
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
