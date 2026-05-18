package gcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// EnsureStorageBucket creates a Cloud Storage bucket if it doesn't exist
func EnsureStorageBucket(ctx context.Context, cfg *GCPConfig, bucketName, installationKey string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)

	// Check if bucket exists
	_, err = bucket.Attrs(ctx)
	if err == nil {
		// Bucket exists
		return nil
	}

	// Bucket doesn't exist, create it
	// Note: GCS bucket names must be globally unique
	bucketAttrs := &storage.BucketAttrs{
		Location:          cfg.Region,
		VersioningEnabled: true,
		Labels: map[string]string{
			"managed-by":      "kloudlite",
			"project":         "kloudlite",
			"purpose":         "k3s-backups",
			"installation-id": installationKey,
		},
		Lifecycle: storage.Lifecycle{
			Rules: []storage.LifecycleRule{
				{
					Action: storage.LifecycleAction{
						Type: storage.DeleteAction,
					},
					Condition: storage.LifecycleCondition{
						AgeInDays: 90, // Delete objects older than 90 days
					},
				},
			},
		},
	}

	if err := bucket.Create(ctx, cfg.Project, bucketAttrs); err != nil {
		// Check if bucket already exists (race condition)
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "409") {
			return nil
		}
		return fmt.Errorf("failed to create bucket '%s': %w", bucketName, err)
	}

	// Wait a moment for bucket creation to propagate
	time.Sleep(2 * time.Second)

	return nil
}

// DeleteStorageBucket removes a Cloud Storage bucket and all its objects
func DeleteStorageBucket(ctx context.Context, cfg *GCPConfig, bucketName string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)

	// Check if bucket exists
	_, err = bucket.Attrs(ctx)
	if err != nil {
		// Bucket doesn't exist
		if strings.Contains(err.Error(), "not exist") || strings.Contains(err.Error(), "404") {
			return nil
		}
		return fmt.Errorf("failed to check bucket: %w", err)
	}

	// Delete all objects in the bucket (including versions)
	it := bucket.Objects(ctx, &storage.Query{
		Versions: true, // Include all versions
	})
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		obj := bucket.Object(attrs.Name)
		if attrs.Generation > 0 {
			obj = obj.Generation(attrs.Generation)
		}

		if err := obj.Delete(ctx); err != nil {
			// Ignore not found errors (object may have been deleted)
			if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "404") {
				return fmt.Errorf("failed to delete object %s: %w", attrs.Name, err)
			}
		}
	}

	// Delete the bucket
	if err := bucket.Delete(ctx); err != nil {
		// Ignore not found errors
		if strings.Contains(err.Error(), "not exist") || strings.Contains(err.Error(), "404") {
			return nil
		}
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	return nil
}

// GetBucketName returns the bucket name for an installation
func GetBucketName(installationKey string) string {
	return fmt.Sprintf("kl-%s-backups", installationKey)
}
