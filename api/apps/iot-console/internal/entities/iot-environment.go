package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
)

type IOTEnvironment struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	Name        string `json:"name"`
	AccountName string `json:"accountName" graphql:"noinput"`
	ProjectName string `json:"projectName" graphql:"noinput"`

	common.ResourceMetadata `json:",inline"`
}

var IOTEnvironmentIndexes = []repos.IndexField{
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
