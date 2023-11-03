package entities

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

type CloudProviderName string

const (
	CloudProviderNameDo        CloudProviderName = "do"
	CloudProviderNameAws       CloudProviderName = "aws"
	CloudProviderNameAzure     CloudProviderName = "azure"
	CloudProviderNameGcp       CloudProviderName = "gcp"
	CloudProviderNameOci       CloudProviderName = "oci"
	CloudProviderNameOpenstack CloudProviderName = "openstack"
	CloudProviderNameVmware    CloudProviderName = "vmware"
)

const (
	AccessKey    string = "accessKey"
	AccessSecret string = "accessSecret"
)

type CloudProviderSecret struct {
	repos.BaseEntity  `json:",inline" graphql:"noinput"`
	corev1.Secret     `json:",inline" graphql:"uri=k8s://secrets.crds.kloudlite.io"`
	CloudProviderName CloudProviderName `json:"cloudProviderName"`

	common.ResourceMetadata `json:",inline"`

	AccountName string `json:"accountName" graphql:"noinput"`
}

var SecretIndices = []repos.IndexField{
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

func (cps *CloudProviderSecret) Validate() (bool, error) {
	if cps == nil {
		return false, fmt.Errorf("cloud provider secret is nil")
	}

	if cps.StringData[AccessKey] == "" || cps.StringData[AccessSecret] == "" {
		return false, fmt.Errorf(".stringData.accessKey or .stringData.accessSecret is empty")
	}

	return true, nil
}
