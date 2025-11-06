package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func EnsureS3Bucket(ctx context.Context, cfg aws.Config, bucketName, installationKey string) error {
	s3Client := s3.NewFromConfig(cfg)

	// Check if bucket exists
	_, err := s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err == nil {
		// Bucket exists
		return nil
	}

	// Create bucket
	createInput := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	}

	// For regions other than us-east-1, we need to specify LocationConstraint
	if cfg.Region != "us-east-1" {
		createInput.CreateBucketConfiguration = &s3Types.CreateBucketConfiguration{
			LocationConstraint: s3Types.BucketLocationConstraint(cfg.Region),
		}
	}

	_, err = s3Client.CreateBucket(ctx, createInput)
	if err != nil {
		return fmt.Errorf("failed to create S3 bucket: %w", err)
	}

	// Enable versioning for backup safety
	_, err = s3Client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: aws.String(bucketName),
		VersioningConfiguration: &s3Types.VersioningConfiguration{
			Status: s3Types.BucketVersioningStatusEnabled,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to enable bucket versioning: %w", err)
	}

	// Add lifecycle policy to expire old backups after 30 days
	_, err = s3Client.PutBucketLifecycleConfiguration(ctx, &s3.PutBucketLifecycleConfigurationInput{
		Bucket: aws.String(bucketName),
		LifecycleConfiguration: &s3Types.BucketLifecycleConfiguration{
			Rules: []s3Types.LifecycleRule{
				{
					ID:     aws.String("expire-old-backups"),
					Status: s3Types.ExpirationStatusEnabled,
					Expiration: &s3Types.LifecycleExpiration{
						Days: aws.Int32(30),
					},
					Filter: &s3Types.LifecycleRuleFilter{
						Prefix: aws.String(""),
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to set lifecycle policy: %w", err)
	}

	// Add tags
	_, err = s3Client.PutBucketTagging(ctx, &s3.PutBucketTaggingInput{
		Bucket: aws.String(bucketName),
		Tagging: &s3Types.Tagging{
			TagSet: []s3Types.Tag{
				{Key: aws.String("Name"), Value: aws.String(bucketName)},
				{Key: aws.String("ManagedBy"), Value: aws.String("kloudlite")},
				{Key: aws.String("Project"), Value: aws.String("kloudlite")},
				{Key: aws.String("Purpose"), Value: aws.String("k3s-backups")},
				{Key: aws.String("kloudlite.io/installation-id"), Value: aws.String(installationKey)},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to tag bucket: %w", err)
	}

	return nil
}

func DeleteS3Bucket(ctx context.Context, cfg aws.Config, bucketName string) error {
	s3Client := s3.NewFromConfig(cfg)

	// List and delete all objects (including versions)
	listInput := &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
	}

	for {
		listOutput, err := s3Client.ListObjectVersions(ctx, listInput)
		if err != nil {
			return fmt.Errorf("failed to list objects: %w", err)
		}

		// Delete versions
		if len(listOutput.Versions) > 0 {
			var objects []s3Types.ObjectIdentifier
			for _, version := range listOutput.Versions {
				objects = append(objects, s3Types.ObjectIdentifier{
					Key:       version.Key,
					VersionId: version.VersionId,
				})
			}

			_, err = s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(bucketName),
				Delete: &s3Types.Delete{
					Objects: objects,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to delete object versions: %w", err)
			}
		}

		// Delete delete markers
		if len(listOutput.DeleteMarkers) > 0 {
			var objects []s3Types.ObjectIdentifier
			for _, marker := range listOutput.DeleteMarkers {
				objects = append(objects, s3Types.ObjectIdentifier{
					Key:       marker.Key,
					VersionId: marker.VersionId,
				})
			}

			_, err = s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(bucketName),
				Delete: &s3Types.Delete{
					Objects: objects,
				},
			})
			if err != nil {
				return fmt.Errorf("failed to delete markers: %w", err)
			}
		}

		// Check if there are more objects
		if !aws.ToBool(listOutput.IsTruncated) {
			break
		}
		listInput.KeyMarker = listOutput.NextKeyMarker
		listInput.VersionIdMarker = listOutput.NextVersionIdMarker
	}

	// Delete the bucket
	_, err := s3Client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	return nil
}
