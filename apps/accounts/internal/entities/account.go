package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
)

type Account struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	crdsv1.Account   `json:",inline" graphql:"uri=k8s://accounts.crds.kloudlite.io"`

	common.ResourceMetadata `json:",inline"`

	ContactEmail string  `json:"contactEmail"`
	Logo         *string `json:"logo"`
	Description  *string `json:"description"`
	IsActive     *bool   `json:"isActive,omitempty"`
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
