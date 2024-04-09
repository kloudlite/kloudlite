package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
)

type IOTDeployment struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	Name             string           `json:"name"`
	AccountName      string           `json:"accountName" graphql:"noinput"`
	ProjectName      string           `json:"projectName" graphql:"noinput"`
	CIDR             string           `json:"CIDR"`
	ExposedServices  []ExposedService `json:"exposedServices"`

	common.ResourceMetadata `json:",inline"`
}

type ExposedService struct {
	Name string `json:"name"`
	IP   string `json:"ip"`
}

var IOTDeploymentIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fields.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "CIDR", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
