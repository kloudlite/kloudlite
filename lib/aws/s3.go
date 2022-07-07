package aws

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"operators.kloudlite.io/lib/errors"
	"operators.kloudlite.io/lib/logging"
	"strings"
)

type s3Obj struct {
	cli    *s3.S3
	logger logging.Logger
}

type PolicyStatement map[string]any

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

	l, err := logging.New(&logging.Options{Name: "aws/s3", Dev: false})
	if err != nil {
		return nil, errors.NewEf(err, "initializing logging")
	}
	return &s3Obj{cli: svc, logger: l}, nil
}

func (s *s3Obj) CreateBucket(bucketName string) error {
	_, err := s.cli.CreateBucket(
		&s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
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

func (s *s3Obj) AddOwnerPolicy(bucketName string, user *User) ([]PolicyStatement, error) {
	// [source](https://aws.amazon.com/blogs/security/writing-iam-policies-how-to-grant-access-to-an-amazon-s3-bucket/)
	if user.ARN == "" {
		return nil, errors.Newf("user should have a valid ARN")
	}
	return []PolicyStatement{
		{
			// bucket-level permissions
			"Effect": "Allow",
			"Action": []string{"s3:ListBucket"},
			"Principal": map[string]string{
				"AWS": user.ARN,
			},
			"Resource": fmt.Sprintf("arn:aws:s3:::%s", bucketName),
		},
		{
			// object-level permissions
			"Effect": "Allow",
			"Principal": map[string]string{
				"AWS": user.ARN,
			},
			"Action": []string{
				"s3:PutObject",
				"s3:GetObject",
				"s3:DeleteObject",
			},
			"Resource": fmt.Sprintf("arn:aws:s3:::%s/*", bucketName),
		},
	}, nil
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

func (s *s3Obj) EmptyBucket(bucketName string) error {
	return s.cli.ListObjectsV2Pages(
		&s3.ListObjectsV2Input{
			Bucket: &bucketName,
		}, func(output *s3.ListObjectsV2Output, b bool) bool {
			for _, object := range output.Contents {
				_, err := s.cli.DeleteObject(
					&s3.DeleteObjectInput{
						Bucket: &bucketName,
						Key:    object.Key,
					},
				)
				if err != nil {
					return false
				}
			}
			return *output.IsTruncated
		},
	)
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

func (s *s3Obj) ensureDirPath(pathName string) string {
	if strings.HasSuffix(pathName, "/") {
		return pathName
	}
	return fmt.Sprintf("%s/", pathName)
}

func (s *s3Obj) MakeObjectsDirsPublic(bucketName string, dirs ...string) ([]PolicyStatement, error) {
	for _, dir := range dirs {
		if err := s.ensureObjectExists(bucketName, s.ensureDirPath(dir)); err != nil {
			return nil, err
		}
	}

	var stmts []PolicyStatement
	for _, dir := range dirs {
		stmts = append(
			stmts, PolicyStatement{
				"Effect":    "Allow",
				"Principal": "*",
				"Action":    "s3:GetObject",
				"Resource":  fmt.Sprintf("arn:aws:s3:::%s/%s*", bucketName, s.ensureDirPath(dir)),
			},
		)
	}
	return stmts, nil
}

func (s *s3Obj) ApplyPolicies(bucketName string, stmts ...PolicyStatement) error {
	m := map[string]any{
		"Version":   "2012-10-17",
		"Statement": stmts,
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

func (s *s3Obj) GetBucket(bucketName string) {

}
