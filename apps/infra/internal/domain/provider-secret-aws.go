package domain

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/sts"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/apps/infra/internal/env"
	"github.com/kloudlite/api/pkg/errors"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
)

func createAwsSession(awscreds *entities.AWSSecretCredentials) (*session.Session, error) {
	sess, err := session.NewSession()
	sess.Config.Region = aws.String("ap-south-1")
	if err != nil {
		return nil, errors.NewE(err)
	}

	svc := sts.New(sess)

	switch awscreds.AuthMechanism {
	case clustersv1.AwsAuthMechanismSecretKeys:
		{
			if awscreds.AuthSecretKeys == nil {
				return nil, fmt.Errorf("auth secret keys not set, can't proceed with cloudformation checks")
			}
			return session.NewSession(&aws.Config{
				Region:      aws.String("ap-south-1"),
				Credentials: credentials.NewStaticCredentials(awscreds.AuthSecretKeys.AccessKey, awscreds.AuthSecretKeys.SecretKey, ""),
			})
		}
	case clustersv1.AwsAuthMechanismAssumeRole:
		{
			resp, err := svc.AssumeRole(&sts.AssumeRoleInput{
				RoleArn:         aws.String(awscreds.AssumeRoleParams.RoleARN),
				ExternalId:      aws.String(awscreds.AssumeRoleParams.ExternalID),
				RoleSessionName: aws.String("TestSession"),
			})
			if err != nil {
				return nil, errors.NewEf(err, "while asumming role identity")
			}

			if resp.AssumedRoleUser == nil || resp.AssumedRoleUser.Arn == nil {
				return nil, fmt.Errorf("AWS assume role (%s) not found", awscreds.AssumeRoleParams.RoleARN)
			}

			return session.NewSession(&aws.Config{
				Region:      aws.String("ap-south-1"),
				Credentials: credentials.NewStaticCredentials(*resp.Credentials.AccessKeyId, *resp.Credentials.SecretAccessKey, *resp.Credentials.SessionToken),
			})
		}
	default:
		{
			return nil, fmt.Errorf("unknown aws auth mechanism: %s", awscreds.AuthMechanism)
		}
	}
}

func checkAwsCloudformationCompletion(awscreds *entities.AWSSecretCredentials) error {
	sess, err := createAwsSession(awscreds)
	if err != nil {
		return errors.NewE(err)
	}

	cf := cloudformation.New(sess)
	dso, err := cf.DescribeStacks(&cloudformation.DescribeStacksInput{
		StackName: &awscreds.CfParamStackName,
	})
	if err != nil {
		return errors.NewE(err)
	}

	stackFound := false

	for i := range dso.Stacks {
		if dso.Stacks[i] != nil && *dso.Stacks[i].StackName == awscreds.CfParamStackName {
			stackFound = true
			if *dso.Stacks[i].StackStatus != cloudformation.StackStatusCreateComplete {
				return errors.Newf("cloudformation stack (%s) is not completed, yet", awscreds.CfParamStackName)
			}
		}
	}

	if !stackFound {
		return errors.Newf("waiting for cloudformation stack to be created")
	}

	return nil
}

func generateAWSCloudformationTemplateUrl(creds entities.AWSSecretCredentials, ev *env.Env) (string, error) {
	var qp []string

	switch creds.AuthMechanism {
	case clustersv1.AwsAuthMechanismSecretKeys:
		{
			qp = []string{
				"templateURL=" + ev.AWSCfStackS3URL,
				"stackName=" + creds.CfParamStackName,
				"param_RoleName=" + creds.CfParamRoleName,
				"param_InstanceProfileName=" + creds.CfParamInstanceProfileName,
				"param_UserName=" + creds.AuthSecretKeys.CfParamUserName,
			}
		}
	case clustersv1.AwsAuthMechanismAssumeRole:
		{
			if creds.AssumeRoleParams == nil {
				return "", errors.Newf("assume role params not defined")
			}
			qp = []string{
				"templateURL=" + ev.AWSCfStackS3URL,
				"stackName=" + creds.CfParamStackName,
				"param_ExternalId=" + creds.AssumeRoleParams.ExternalID,
				"param_TrustedArn=" + creds.AssumeRoleParams.CfParamTrustedARN,
				"param_RoleName=" + creds.CfParamRoleName,
				"param_InstanceProfileName=" + creds.CfParamInstanceProfileName,
			}
		}
	}

	result := bytes.NewBuffer(nil)
	_, err := fmt.Fprintf(result, "https://console.aws.amazon.com/cloudformation/home#/stacks/quickcreate?")
	if err != nil {
		return "", errors.NewE(err)
	}
	_, err = fmt.Fprint(result, strings.Join(qp, "&"))
	if err != nil {
		return "", errors.NewE(err)
	}
	return result.String(), nil
}

type AWSAccessValidationOutput struct {
	Result          bool
	InstallationURL *string
}

func (d *domain) ValidateProviderSecretAWSAccess(ctx InfraContext, name string) (*AWSAccessValidationOutput, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateCloudProviderSecret); err != nil {
		return nil, errors.NewE(err)
	}

	psecret, err := d.findProviderSecret(ctx, name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := psecret.Validate(); err != nil {
		return nil, errors.NewE(err)
	}

	if err := checkAwsCloudformationCompletion(psecret.AWS); err != nil {
		installationURL, err := generateAWSCloudformationTemplateUrl(*psecret.AWS, d.env)
		if err != nil {
			return nil, errors.NewE(err)
		}
		return &AWSAccessValidationOutput{
			Result:          false,
			InstallationURL: &installationURL,
		}, nil
	}

	return &AWSAccessValidationOutput{Result: true}, nil
}
