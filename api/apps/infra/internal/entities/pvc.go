package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
)

type PersistentVolumeClaim struct {
	repos.BaseEntity             `json:",inline" graphql:"noinput"`
	corev1.PersistentVolumeClaim `json:",inline" graphql:"noinput"`

	AccountName string `json:"accountName" graphql:"noinput"`
	ClusterName string `json:"clusterName" graphql:"noinput"`

	common.ResourceMetadata `json:",inline" graphql:"noinput"`
	SyncStatus              types.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (p *PersistentVolumeClaim) GetDisplayName() string {
	return p.ResourceMetadata.DisplayName
}

func (p *PersistentVolumeClaim) GetStatus() reconciler.Status {
	return reconciler.Status{}
}

var PersistentVolumeClaimIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "metadata.namespace", Value: repos.IndexAsc},
			{Key: "accountName", Value: repos.IndexAsc},
			{Key: "clusterName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
