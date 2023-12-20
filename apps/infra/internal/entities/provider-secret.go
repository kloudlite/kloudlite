package entities

import (
	"fmt"
	"github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
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

type AWSSecretCredentials struct {
	AccessKey *string `json:"accessKey,omitempty"`
	SecretKey *string `json:"secretKey,omitempty"`

	AWSAccountId               *string `json:"awsAccountId,omitempty"`
	CfParamStackName           string  `json:"cfParamStackName,omitempty" graphql:"noinput"`
	CfParamRoleName            string  `json:"cfParamRoleName,omitempty" graphql:"noinput"`
	CfParamInstanceProfileName string  `json:"cfParamInstanceProfileName,omitempty" graphql:"noinput"`
	CfParamTrustedARN          string  `json:"cfParamTrustedARN,omitempty" graphql:"noinput"`
	CfParamExternalID          string  `json:"cfParamExternalID,omitempty" graphql:"noinput"`
}

func (asc *AWSSecretCredentials) GetAssumeRoleRoleARN() string {
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", *asc.AWSAccountId, asc.CfParamRoleName)
}

func (asc *AWSSecretCredentials) IsAssumeRoleConfiguration() bool {
	return asc.AccessKey == nil || asc.SecretKey == nil
}

func (asc *AWSSecretCredentials) Validate() error {
	if asc == nil {
		return errors.Newf("aws secret credentials, is nil")
	}

	if asc.AccessKey != nil || asc.SecretKey != nil {
		return nil
	}

	if asc.AWSAccountId == nil {
		return errors.Newf("awsAccountId, must be provided")
	}

	if asc.CfParamStackName == "" {
		return errors.Newf("cfParamStackName, must be provided")
	}
	if asc.CfParamExternalID == "" {
		return errors.Newf("cfParamExternalID, must be provided")
	}
	if asc.CfParamRoleName == "" {
		return errors.Newf("cfParamRoleName, must be provided")
	}
	if asc.CfParamTrustedARN == "" {
		return errors.Newf("cfParamTrustedARN, must be provided")
	}
	if asc.CfParamInstanceProfileName == "" {
		return errors.Newf("cfParamInstanceProfileName, must be provided")
	}

	return nil
}

type CloudProviderSecret struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	// corev1.Secret     `json:",inline" graphql:"uri=k8s://secrets.crds.kloudlite.io"`
	metav1.ObjectMeta `json:"metadata"`
	CloudProviderName ct.CloudProvider `json:"cloudProviderName"`

	common.ResourceMetadata `json:",inline"`
	AWS                     *AWSSecretCredentials `json:"aws,omitempty"`

	AccountName string `json:"accountName" graphql:"noinput"`
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
				return errors.Newf(".aws is nil, must be provided when cloudproviderName is set to aws")
			}
			if cps.AWS.AWSAccountId == nil && (cps.AWS.AccessKey == nil || cps.AWS.SecretKey == nil) {
				return errors.Newf("neither .aws.%s nor (.aws.%s and .aws.%s) is provided", AWSAccountId, AccessKey, SecretKey)
			}
		}
	default:
		{
			// if cps.StringData[AccessKey] == "" || cps.StringData[SecretKey] == "" {
			// 	return false, errors.Newf(".stringData.accessKey or .stringData.accessSecret is empty")
			// }
		}
	}

	return nil
}
