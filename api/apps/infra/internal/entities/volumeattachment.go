package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/toolkit/reconciler"
	storagev1 "k8s.io/api/storage/v1"
)

type VolumeAttachment struct {
	repos.BaseEntity           `json:",inline" graphql:"noinput"`
	storagev1.VolumeAttachment `json:",inline"`

	AccountName string `json:"accountName" graphql:"noinput"`
	ClusterName string `json:"clusterName" graphql:"noinput"`

	common.ResourceMetadata `json:",inline" graphql:"noinput"`
	SyncStatus              types.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (v *VolumeAttachment) GetDisplayName() string {
	return v.ResourceMetadata.DisplayName
}

func (v *VolumeAttachment) GetStatus() reconciler.Status {
	return reconciler.Status{}
}

var VolumeAttachmentIndices = []repos.IndexField{
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
