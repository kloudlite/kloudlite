package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
)

type Config struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	corev1.ConfigMap `json:",inline"`

	AccountName     string `json:"accountName" graphql:"noinput"`
	EnvironmentName string `json:"environmentName" graphql:"noinput"`

	common.ResourceMetadata `json:",inline"`
	SyncStatus              t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (c *Config) GetDisplayName() string {
	return c.ResourceMetadata.DisplayName
}

func (c *Config) GetStatus() operator.Status {
	return operator.Status{}
}

func (c *Config) GetResourceType() ResourceType {
	return ResourceTypeConfig
}

var ConfigIndexes = []repos.IndexField{
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
			{Key: fields.EnvironmentName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
