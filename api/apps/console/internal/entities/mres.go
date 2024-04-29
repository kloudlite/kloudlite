package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
)

type ManagedResource struct {
	repos.BaseEntity       `json:",inline" graphql:"noinput"`
	crdsv1.ManagedResource `json:",inline"`

	AccountName     string `json:"accountName" graphql:"noinput"`
	EnvironmentName string `json:"environmentName" graphql:"noinput"`

	// SyncedOutputSecretRef *corev1.Secret `json:"syncedOutputSecretRef" graphql:"noinput" struct-json-path:",ignore-nesting"`
	SyncedOutputSecretRef *corev1.Secret `json:"syncedOutputSecretRef" graphql:"noinput"`

	common.ResourceMetadata `json:",inline"`
	SyncStatus              t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (m *ManagedResource) GetDisplayName() string {
	return m.ResourceMetadata.DisplayName
}

func (m *ManagedResource) GetStatus() operator.Status {
	return m.ManagedResource.Status
}

func (m *ManagedResource) GetResourceType() ResourceType {
	return ResourceTypeManagedResource
}

var MresIndexes = []repos.IndexField{
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
