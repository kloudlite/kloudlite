package entities

import (
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

type GitSource struct {
	Repository string  `json:"repository"`
	PullSecret *string `json:"pullSecret"`
	Branch     string  `json:"branch"`
}

type Build struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	CreatedBy     common.CreatedOrUpdatedBy `json:"createdBy" graphql:"noinput"`
	LastUpdatedBy common.CreatedOrUpdatedBy `json:"lastUpdatedBy" graphql:"noinput"`

	AccountName string `json:"accountName" graphql:"noinput"`
	Repository  string `json:"repository"`
	Name        string `json:"name"`

	Source GitSource `json:"source"`
	Tag    string    `json:"tag"`
}

var BuildIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
