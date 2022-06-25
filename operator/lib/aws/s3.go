package aws

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/logger"
	"strings"
)

type s3Obj struct {
	cli    *s3.S3
	logger logger.Logger
}

func NewS3Client(region string) (*s3Obj, error) {
	sess, err := newSession()
	if err != nil {
		return nil, err
	}
	svc := s3.New(
		sess, &aws.Config{
			Region: aws.String(region),
		},
	)

	l, err := logger.New(true)
	if err != nil {
		return nil, errors.NewEf(err, "initializing logger")
	}
	return &s3Obj{cli: svc, logger: l}, nil
}

func (s *s3Obj) CreateBucket(bucketName string) error {
	_, err := s.cli.CreateBucket(
		&s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
			// CreateBucketConfiguration: &cli.CreateBucketConfiguration{
			// 	LocationConstraint: aws.String(region),
			// },
		},
	)

	if err != nil {
		if aErr, ok := err.(awserr.Error); ok && aErr.Code() == s3.ErrCodeBucketAlreadyOwnedByYou {
			s.logger.Debugf("bucket already exists and owned by user")
			return nil
		}
		return errors.NewEf(err, "while creating bucket")
	}

	// Wait until bucket is created before finishing
	s.logger.Infof("Waiting for bucket %q to be created...\n", bucketName)

	return s.cli.WaitUntilBucketExists(
		&s3.HeadBucketInput{
			Bucket: aws.String(bucketName),
		},
	)
}

func (s *s3Obj) DeleteBucket(bucketName string) error {
	_, err := s.cli.DeleteBucket(
		&s3.DeleteBucketInput{
			Bucket: &bucketName,
		},
	)
	if err != nil {
		if aErr, ok := err.(awserr.Error); ok && aErr.Code() == s3.ErrCodeNoSuchBucket {
			return nil
		}
		return err
	}
	return nil
}

func (s *s3Obj) MakePublicReadable(bucketName string) error {
	// private | public-read | public-read-write | authenticated-read
	// See https://docs.aws.amazon.com/AmazonS3/latest/dev/acl-overview.html#CannedACL for details
	acl := "public-read"

	params := &s3.PutBucketAclInput{
		Bucket: &bucketName,
		ACL:    &acl,
	}

	// Set bucket ACL
	_, err := s.cli.PutBucketAcl(params)
	if err != nil {
		return err
	}
	return nil
}

func (s *s3Obj) ensureObjectExists(bucketName string, objectKey string) error {
	_, err := s.cli.PutObject(
		&s3.PutObjectInput{
			Bucket: &bucketName,
			Key:    &objectKey,
		},
	)
	if err != nil {
		return err
	}

	return s.cli.WaitUntilObjectExists(
		&s3.HeadObjectInput{
			Bucket: &bucketName,
			Key:    &objectKey,
		},
	)
}

func (s *s3Obj) MakeObjectsPublic(bucketName string, objectKey string) error {
	acl := "public-read"

	key := objectKey
	if !strings.HasSuffix(key, "/") {
		key = fmt.Sprintf("%s/", key)
	}

	if err := s.ensureObjectExists(bucketName, key); err != nil {
		return err
	}

	_, err := s.cli.PutObjectAcl(
		&s3.PutObjectAclInput{
			ACL:    &acl,
			Bucket: &bucketName,
			Key:    &key,
		},
	)
	if err != nil {
		return err
	}

	m := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Effect":    "Allow",
				"Principal": "*",
				"Action":    "s3:GetObject",
				"Resource":  fmt.Sprintf("arn:aws:s3:::%s/public/*", bucketName),
				// "Condition": map[string]any{
				// 	"StringEquals": map[string]any{
				// 		"s3:ExistingObjectTag/public": "yes",
				// 	},
				// },
			},
		},
	}

	policyB, err := json.Marshal(m)
	if err != nil {
		return err
	}
	policyJson := string(policyB)

	if _, err := s.cli.PutBucketPolicy(
		&s3.PutBucketPolicyInput{
			Bucket: &bucketName,
			Policy: &policyJson,
		},
	); err != nil {
		return err
	}

	return nil
}
