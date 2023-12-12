package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
)

type BuildCacheKey struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	AccountName string `json:"accountName" graphql:"noinput"`

	CreatedBy     common.CreatedOrUpdatedBy `json:"createdBy" graphql:"noinput"`
	LastUpdatedBy common.CreatedOrUpdatedBy `json:"lastUpdatedBy" graphql:"noinput"`

	VolumeSize float32 `json:"volumeSizeInGB"`
}

var BuildCacheKeyIndexes = []repos.IndexField{
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
