package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Account struct {
	repos.BaseEntity  `json:",inline" graphql:"noinput"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	TargetNamespace   string `json:"targetNamespace,omitempty" graphql:"noinput"`

	common.ResourceMetadata `json:",inline"`

	Logo         *string `json:"logo"`
	IsActive     *bool   `json:"isActive,omitempty"`
	ContactEmail string  `json:"contactEmail,omitempty"`

	KloudliteGatewayRegion string `json:"kloudliteGatewayRegion"`
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
	{
		Field: []repos.IndexKey{
			{Key: "targetNamespace", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
