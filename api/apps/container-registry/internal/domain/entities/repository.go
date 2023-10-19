package entities

import (
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

type Repository struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	CreatedBy     common.CreatedOrUpdatedBy `json:"createdBy" graphql:"noinput"`
	LastUpdatedBy common.CreatedOrUpdatedBy `json:"lastUpdatedBy" graphql:"noinput"`

	AccountName string `json:"accountName" graphql:"noinput"`
	Name        string `json:"name"`
}

var RepositoryIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "name", Value: repos.IndexAsc},
			{Key: "accountName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
