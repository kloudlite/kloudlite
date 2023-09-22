package entities

import (
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

type Build struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	CreatedBy     common.CreatedOrUpdatedBy `json:"createdBy" graphql:"noinput"`
	LastUpdatedBy common.CreatedOrUpdatedBy `json:"lastUpdatedBy" graphql:"noinput"`

	AccountName string  `json:"accountName" graphql:"noinput"`
	Repository  string  `json:"repository"`
	PullSecret  *string `json:"pullSecret"`

	Name string `json:"name"`
}

var BuildIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
