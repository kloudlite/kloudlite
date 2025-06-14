package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
)

type ClusterManagedService struct {
	repos.BaseEntity             `json:",inline" graphql:"noinput"`
	crdsv1.ClusterManagedService `json:",inline"`

	AccountName string `json:"accountName" graphql:"noinput"`
	ClusterName string `json:"clusterName"`

	IsArchived *bool `json:"isArchived,omitempty" graphql:"noinput"`

	SyncedOutputSecretRef *corev1.Secret `json:"syncedOutputSecretRef" graphql:"ignore"`

	common.ResourceMetadata `json:",inline"`
	SyncStatus              t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (c *ClusterManagedService) GetDisplayName() string {
	return c.DisplayName
}

func (c *ClusterManagedService) GetStatus() reconciler.Status {
	return c.Status
}

var ClusterManagedServiceIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "accountName", Value: repos.IndexAsc},
			{Key: "clusterName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
