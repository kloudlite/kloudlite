package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
)

type Secret struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	corev1.Secret    `json:",inline"`

	AccountName     string `json:"accountName" graphql:"noinput"`
	EnvironmentName string `json:"environmentName" graphql:"noinput"`

	// For is the resource type and name of the resource that this secret is being created for in format <resource-type>/<namespace>/<name>
	// It is supposed to be nil for traditional secrets, and to be used by ImagePullSecrets / Imported Managed Resources
	For *SecretCreatedFor `json:"for,omitempty" graphql:"noinput"`

	common.ResourceMetadata `json:",inline"`
	SyncStatus              t.SyncStatus `json:"syncStatus" graphql:"noinput"`

	IsReadOnly    bool    `json:"isReadyOnly" graphql:"noinput"`
	CreatedByHelm *string `json:"createdByHelm,omitempty" graphql:"noinput"`
}

type SecretCreatedFor struct {
	RefId        repos.ID     `json:"refId"`
	ResourceType ResourceType `json:"resourceType"`
	Name         string       `json:"name"`
	Namespace    string       `json:"namespace"`
}

func (s *Secret) GetDisplayName() string {
	return s.ResourceMetadata.DisplayName
}

func (s *Secret) GetStatus() reconciler.Status {
	return reconciler.Status{}
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
			{Key: fields.EnvironmentName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
