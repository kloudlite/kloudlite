package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Team struct {
	repos.BaseEntity  `json:",inline" graphql:"noinput"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	
	common.ResourceMetadata `json:",inline"`

	Slug         string  `json:"slug"`
	Description  string  `json:"description,omitempty"`
	Logo         *string `json:"logo"`
	IsActive     *bool   `json:"isActive,omitempty"`
	ContactEmail string  `json:"contactEmail,omitempty"`
	Region       string  `json:"region"`
	OwnerId      repos.ID `json:"ownerId"`
}

var TeamIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "slug", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "ownerId", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}