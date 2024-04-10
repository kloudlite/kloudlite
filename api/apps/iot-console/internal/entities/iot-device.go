package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
)

type IOTDevice struct {
	repos.BaseEntity        `json:",inline" graphql:"noinput"`
	common.ResourceMetadata `json:",inline"`

	Name           string `json:"name"`
	AccountName    string `json:"accountName" graphql:"noinput"`
	ProjectName    string `json:"projectName" graphql:"noinput"`
	PublicKey      string `json:"publicKey"`
	ServiceCIDR    string `json:"serviceCIDR"`
	PodCIDR        string `json:"podCIDR"`
	IP             string `json:"ip"`
	DeploymentName string `json:"deploymentName" graphql:"noinput"`
	Version        string `json:"version"`
}

var IOTDeviceIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fields.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{
				Key:   "name",
				Value: repos.IndexAsc,
			},
		},
		Unique: true,
	},
}
