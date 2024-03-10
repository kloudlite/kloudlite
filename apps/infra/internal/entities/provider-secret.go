package entities

import (
	"fmt"

	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/operator/pkg/operator"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	ct "github.com/kloudlite/operator/apis/common-types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	AccessKey string = "accessKey"
	SecretKey string = "secretKey"

	AWSAccountId            string = "awsAccountId"
	AWSAssumeRoleExternalId string = "awsAssumeRoleExternalId"
	AWAssumeRoleRoleARN     string = "awsAssumeRoleRoleARN"
	AWSInstanceProfileName  string = "awsInstanceProfileName"
)

type AWSAssumeRoleParams struct {
	AWSAccountID                   string `json:"awsAccountId"`
	CfParamTrustedARN              string `json:"cfParamTrustedARN" graphql:"noinput"`
	clustersv1.AwsAssumeRoleParams `json:",inline" graphql:"noinput"`
}

type AWSAuthSecretKeys struct {
	CfParamUserName              string `json:"cfParamUserName" graphql:"noinput"`
	clustersv1.AwsAuthSecretKeys `json:",inline"`
}

type AWSSecretCredentials struct {
	CfParamStackName           string `json:"cfParamStackName,omitempty" graphql:"noinput"`
	CfParamRoleName            string `json:"cfParamRoleName,omitempty" graphql:"noinput"`
	CfParamInstanceProfileName string `json:"cfParamInstanceProfileName,omitempty" graphql:"noinput"`

	AuthMechanism clustersv1.AwsAuthMechanism `json:"authMechanism"`

	AuthSecretKeys   *AWSAuthSecretKeys   `json:"authSecretKeys,omitempty"`
	AssumeRoleParams *AWSAssumeRoleParams `json:"assumeRoleParams,omitempty"`
}

func (asc *AWSSecretCredentials) GetAssumeRoleRoleARN() string {
	if asc.AssumeRoleParams != nil {
		return fmt.Sprintf("arn:aws:iam::%s:role/%s", asc.AssumeRoleParams.AWSAccountID, asc.CfParamRoleName)
	}
	return ""
}

func (asc *AWSSecretCredentials) IsAssumeRoleConfiguration() bool {
	return asc.AuthMechanism == clustersv1.AwsAuthMechanismAssumeRole
}

func (asc *AWSSecretCredentials) Validate() error {
	if asc == nil {
		return errors.Newf("aws secret credentials, is nil")
	}

	switch asc.AuthMechanism {
	case clustersv1.AwsAuthMechanismSecretKeys:
		{
			if asc.AuthSecretKeys == nil {
				return fmt.Errorf("with aws auth mechanism (%s), secretKeys must be set", asc.AuthMechanism)
			}
			if asc.AuthSecretKeys.AccessKey == "" || asc.AuthSecretKeys.SecretKey == "" {
				return fmt.Errorf("with aws auth mechanism (%s), secretKeys accessKey, and secretKey must be set", asc.AuthMechanism)
			}
		}

	case clustersv1.AwsAuthMechanismAssumeRole:
		{
			if asc.AssumeRoleParams == nil {
				return errors.Newf(".spec.assumeRoleParams, must be set, when accessKey and secretKey are not set")
			}

			if asc.AssumeRoleParams.AWSAccountID == "" {
				return errors.Newf("awsAccountId, must be provided")
			}

			if asc.CfParamStackName == "" {
				return errors.Newf("cfParamStackName, must be provided")
			}
			if asc.AssumeRoleParams.ExternalID == "" {
				return errors.Newf("ExternalID, must be provided")
			}
			if asc.CfParamRoleName == "" {
				return errors.Newf("cfParamRoleName, must be provided")
			}
			if asc.AssumeRoleParams.CfParamTrustedARN == "" {
				return errors.Newf("CfParamTrustedARN, must be provided")
			}
			if asc.CfParamInstanceProfileName == "" {
				return errors.Newf("cfParamInstanceProfileName, must be provided")
			}
		}
	default:
		{
			return fmt.Errorf("unknown aws auth mechanism (%s)", asc.AuthMechanism)
		}
	}

	return nil
}

type CloudProviderSecret struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	AccountName      string `json:"accountName" graphql:"noinput"`

	metav1.ObjectMeta `json:"metadata"`

	CloudProviderName ct.CloudProvider `json:"cloudProviderName"`

	common.ResourceMetadata `json:",inline"`
	AWS                     *AWSSecretCredentials `json:"aws,omitempty"`
}

func (cps *CloudProviderSecret) GetDisplayName() string {
	return cps.ResourceMetadata.DisplayName
}

func (cps *CloudProviderSecret) GetStatus() operator.Status {
	return operator.Status{}
}

var CloudProviderSecretIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},

	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "metadata.namespace", Value: repos.IndexAsc},
		},
		Unique: true,
	},

	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "accountName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

func (cps *CloudProviderSecret) Validate() error {
	if cps == nil {
		return errors.Newf("cloud provider secret is nil")
	}

	switch cps.CloudProviderName {
	case ct.CloudProviderAWS:
		{
			if cps.AWS == nil {
				return errors.Newf(".aws is nil, it must be provided when cloudproviderName is set to aws")
			}

			return nil
			// return cps.AWS.Validate()
		}
	default:
		{
			return fmt.Errorf("not implemented for cloudprovider (%s)", cps.CloudProviderName)
		}
	}
}
