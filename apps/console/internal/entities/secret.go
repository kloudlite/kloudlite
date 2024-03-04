package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
)

type Secret struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	corev1.Secret    `json:",inline"`

	AccountName     string `json:"accountName" graphql:"noinput"`
	ProjectName     string `json:"projectName" graphql:"noinput"`
	EnvironmentName string `json:"environmentName" graphql:"noinput"`

	common.ResourceMetadata `json:",inline"`
	SyncStatus              t.SyncStatus `json:"syncStatus" graphql:"noinput"`

	IsReadOnly bool `json:"isReadyOnly" graphql:"noinput"`
}

func (s *Secret) GetDisplayName() string {
	return s.ResourceMetadata.DisplayName
}

func (s *Secret) GetStatus() operator.Status {
	return operator.Status{}
}

func (s *Secret) GetResourceType() ResourceType {
	return ResourceTypeSecret
}

var SecretIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fields.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fields.MetadataName, Value: repos.IndexAsc},
			{Key: fields.MetadataNamespace, Value: repos.IndexAsc},
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fields.ProjectName, Value: repos.IndexAsc},
			{Key: fields.EnvironmentName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
