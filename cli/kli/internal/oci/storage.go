package oci

import (
	"context"
	"fmt"
	"strings"

	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

// GetBucketName returns the bucket name for an installation
func GetBucketName(installationKey string) string {
	return fmt.Sprintf("kl-%s-backups", installationKey)
}

// EnsureStorageBucket creates an Object Storage bucket if it doesn't exist
func EnsureStorageBucket(ctx context.Context, cfg *OCIConfig, bucketName, installationKey string) error {
	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return fmt.Errorf("failed to create object storage client: %w", err)
	}

	// Get the namespace (required for OCI Object Storage)
	nsResp, err := client.GetNamespace(ctx, objectstorage.GetNamespaceRequest{
		CompartmentId: &cfg.CompartmentOCID,
	})
	if err != nil {
		return fmt.Errorf("failed to get object storage namespace: %w", err)
	}
	namespace := *nsResp.Value

	// Check if bucket exists
	_, err = client.GetBucket(ctx, objectstorage.GetBucketRequest{
		NamespaceName: &namespace,
		BucketName:    &bucketName,
	})
	if err == nil {
		return nil // Bucket exists
	}

	// Create bucket
	tags := freeformTags(installationKey)
	tags["purpose"] = "k3s-backups"

	_, err = client.CreateBucket(ctx, objectstorage.CreateBucketRequest{
		NamespaceName: &namespace,
		CreateBucketDetails: objectstorage.CreateBucketDetails{
			Name:             &bucketName,
			CompartmentId:    &cfg.CompartmentOCID,
			PublicAccessType: objectstorage.CreateBucketDetailsPublicAccessTypeNopublicaccess,
			FreeformTags:     tags,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "BucketAlreadyExists") || strings.Contains(err.Error(), "409") {
			return nil
		}
		return fmt.Errorf("failed to create bucket '%s': %w", bucketName, err)
	}

	return nil
}

// DeleteStorageBucket removes an Object Storage bucket and all its objects
func DeleteStorageBucket(ctx context.Context, cfg *OCIConfig, bucketName string) error {
	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return fmt.Errorf("failed to create object storage client: %w", err)
	}

	// Get the namespace
	nsResp, err := client.GetNamespace(ctx, objectstorage.GetNamespaceRequest{
		CompartmentId: &cfg.CompartmentOCID,
	})
	if err != nil {
		return fmt.Errorf("failed to get object storage namespace: %w", err)
	}
	namespace := *nsResp.Value

	// Check if bucket exists
	_, err = client.GetBucket(ctx, objectstorage.GetBucketRequest{
		NamespaceName: &namespace,
		BucketName:    &bucketName,
	})
	if err != nil {
		if strings.Contains(err.Error(), "BucketNotFound") || strings.Contains(err.Error(), "404") {
			return nil
		}
		return fmt.Errorf("failed to check bucket: %w", err)
	}

	// Delete all objects in the bucket
	var nextStart *string
	for {
		listResp, err := client.ListObjects(ctx, objectstorage.ListObjectsRequest{
			NamespaceName: &namespace,
			BucketName:    &bucketName,
			Start:         nextStart,
		})
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range listResp.Objects {
			_, delErr := client.DeleteObject(ctx, objectstorage.DeleteObjectRequest{
				NamespaceName: &namespace,
				BucketName:    &bucketName,
				ObjectName:    obj.Name,
			})
			if delErr != nil {
				if !strings.Contains(delErr.Error(), "404") {
					return fmt.Errorf("failed to delete object %s: %w", *obj.Name, delErr)
				}
			}
		}

		if listResp.NextStartWith == nil {
			break
		}
		nextStart = listResp.NextStartWith
	}

	// Delete the bucket
	_, err = client.DeleteBucket(ctx, objectstorage.DeleteBucketRequest{
		NamespaceName: &namespace,
		BucketName:    &bucketName,
	})
	if err != nil {
		if strings.Contains(err.Error(), "BucketNotFound") || strings.Contains(err.Error(), "404") {
			return nil
		}
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	return nil
}

// GetObjectStorageNamespace returns the Object Storage namespace for the tenancy
func GetObjectStorageNamespace(ctx context.Context, cfg *OCIConfig) (string, error) {
	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(cfg.ConfigProvider)
	if err != nil {
		return "", fmt.Errorf("failed to create object storage client: %w", err)
	}

	nsResp, err := client.GetNamespace(ctx, objectstorage.GetNamespaceRequest{
		CompartmentId: &cfg.CompartmentOCID,
	})
	if err != nil {
		return "", fmt.Errorf("failed to get object storage namespace: %w", err)
	}

	return *nsResp.Value, nil
}
