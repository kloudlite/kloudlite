package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"operators.kloudlite.io/pkg/errors"
)

type iamObj struct {
	cli *iam.IAM
}

type User struct {
	Name string
	ARN  string
}

// [source](https://docs.aws.amazon.com/AmazonS3/latest/userguide/example-walkthroughs-managing-access-example1.html)

func (i *iamObj) CreateUser(username string) (*User, error) {
	userRef, err := i.getUser(username)
	if err != nil {
		if err2, ok := err.(awserr.Error); ok && err2.Code() == iam.ErrCodeNoSuchEntityException {
			nUserRef, err3 := i.cli.CreateUser(
				&iam.CreateUserInput{
					Path:                nil,
					PermissionsBoundary: nil,
					Tags:                nil,
					UserName:            &username,
				},
			)
			if err3 != nil {
				return nil, err3
			}

			if err := i.cli.WaitUntilUserExists(&iam.GetUserInput{UserName: &username}); err != nil {
				return nil, err
			}

			return &User{Name: *nUserRef.User.UserName, ARN: *nUserRef.User.Arn}, nil
		}
	}

	if userRef == nil {
		return nil, errors.Newf("something wrong with getUser()")
	}

	return &User{Name: *userRef.User.UserName, ARN: *userRef.User.Arn}, nil
}

func (i *iamObj) GetUser(username string) (*User, error) {
	userRef, err := i.getUser(username)
	if err != nil {
		return nil, err
	}
	return &User{
		Name: *userRef.User.UserName,
		ARN:  *userRef.User.Arn,
	}, nil
}

func (i *iamObj) DeleteUser(username string) error {
	if err := i.deleteAccessKey(username); err != nil {
		return err
	}
	_, err := i.cli.DeleteUser(
		&iam.DeleteUserInput{
			UserName: &username,
		},
	)
	if err != nil {
		if aErr, ok := err.(awserr.Error); ok && aErr.Code() == iam.ErrCodeNoSuchEntityException {
			return nil
		}
		return err
	}
	return nil
}

func (i *iamObj) CreateAccessKey(username string) (string, string, error) {
	out, err := i.cli.CreateAccessKey(
		&iam.CreateAccessKeyInput{
			UserName: aws.String(username),
		},
	)
	if err != nil {
		return "", "", err
	}
	return *out.AccessKey.AccessKeyId, *out.AccessKey.SecretAccessKey, nil
}

func (i *iamObj) DeleteAccessKey(username string) error {
	return i.deleteAccessKey(username)
}

func (i *iamObj) deleteAccessKey(username string) error {
	keysRef, err := i.cli.ListAccessKeys(
		&iam.ListAccessKeysInput{
			UserName: &username,
		},
	)
	if err != nil {
		return nil
	}

	for _, keyMeta := range keysRef.AccessKeyMetadata {
		if _, err := i.cli.DeleteAccessKey(
			&iam.DeleteAccessKeyInput{
				AccessKeyId: keyMeta.AccessKeyId,
				UserName:    &username,
			},
		); err != nil {
			return err
		}
	}
	return nil
}

func (i *iamObj) getUser(username string) (*iam.GetUserOutput, error) {
	return i.cli.GetUser(&iam.GetUserInput{UserName: &username})
}

func NewIAMClient() (*iamObj, error) {
	s, err := newSession()
	if err != nil {
		return nil, err
	}
	iamCli := iam.New(s)
	return &iamObj{cli: iamCli}, nil
}
