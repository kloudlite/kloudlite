package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
)

type ManagedResource struct {
	repos.BaseEntity       `json:",inline" graphql:"noinput"`
	crdsv1.ManagedResource `json:",inline"`

	AccountName        string `json:"accountName" graphql:"noinput"`
	EnvironmentName    string `json:"environmentName" graphql:"noinput"`
	ManagedServiceName string `json:"managedServiceName" graphql:"noinput"`
	ClusterName        string `json:"clusterName" graphql:"noinput"`

	SyncedOutputSecretRef *corev1.Secret `json:"syncedOutputSecretRef" graphql:"noinput"`

	common.ResourceMetadata `json:",inline"`
	SyncStatus              t.SyncStatus `json:"syncStatus" graphql:"noinput"`

	IsImported bool   `json:"isImported" graphql:"noinput"`
	MresRef    string `json:"mresRef" graphql:"noinput"`
}

func (m *ManagedResource) GetDisplayName() string {
	return m.ResourceMetadata.DisplayName
}

func (m *ManagedResource) GetStatus() reconciler.Status {
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
