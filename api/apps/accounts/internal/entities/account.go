package entities

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"kloudlite.io/pkg/repos"
)

type Account struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	crdsv1.Account   `json:",inline" graphql:"uri=k8s://accounts.crds.kloudlite.io"`

	DisplayName  string `json:"displayName"`
	ContactEmail string `json:"contactEmail"`
	IsActive     *bool  `json:"isActive,omitempty"`
	IsDeleted    *bool  `json:"isDeleted,omitempty"`
}

var AccountIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
