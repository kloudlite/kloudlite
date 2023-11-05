package entities

import (
	"fmt"

	ct "github.com/kloudlite/operator/apis/common-types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

const (
	AccessKey string = "accessKey"
	SecretKey string = "secretKey"

	AWSAccountId            string = "awsAccountId"
	AWSAssumeRoleExternalId string = "awsAssumeRoleExternalId"
	AWAssumeRoleRoleARN     string = "awsAssumeRoleRoleARN"
)

type AWSSecretCredentials struct {
	AWSAccountId            *string `json:"awsAccountId,omitempty"`
	AWSAssumeRoleExternalId string  `json:"awsAssumeRoleExternalId,omitempty" graphql:"noinput"`
	AWAssumeRoleRoleARN     string  `json:"awsAssumeRoleRoleARN,omitempty" graphql:"noinput"`

	AccessKey *string `json:"accessKey,omitempty"`
	SecretKey *string `json:"secretKey,omitempty"`
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
		return fmt.Errorf("cloud provider secret is nil")
	}

	switch cps.CloudProviderName {
	case ct.CloudProviderAWS:
		{
			if cps.AWS == nil {
				return fmt.Errorf(".aws is nil, must be provided when cloudproviderName is set to aws")
			}
			if cps.AWS.AWSAccountId == nil && (cps.AWS.AccessKey == nil || cps.AWS.SecretKey == nil) {
				return fmt.Errorf("neither .aws.%s nor (.aws.%s and .aws.%s) is provided", AWSAccountId, AccessKey, SecretKey)
			}
		}
	default:
		{
			// if cps.StringData[AccessKey] == "" || cps.StringData[SecretKey] == "" {
			// 	return false, fmt.Errorf(".stringData.accessKey or .stringData.accessSecret is empty")
			// }
		}
	}

	return nil
}
