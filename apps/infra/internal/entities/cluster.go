package entities

import (
	"time"

	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/toolkit/reconciler"
)

type Cluster struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	clustersv1.Cluster `json:",inline"`

	// if not specified, a default will be used, each cluster must be part of one global VPN
	GlobalVPN *string `json:"globalVPN"`

	common.ResourceMetadata `json:",inline"`

	AccountName string       `json:"accountName" graphql:"noinput"`
	SyncStatus  t.SyncStatus `json:"syncStatus" graphql:"noinput"`

	LastOnlineAt *time.Time `json:"lastOnlineAt,omitempty" graphql:"noinput"`

	OwnedBy *string `json:"ownedBy,omitempty" graphql:"noinput"`
}

func (c *Cluster) GetDisplayName() string {
	return c.ResourceMetadata.DisplayName
}

func (c *Cluster) GetStatus() reconciler.Status {
	return reconciler.Status{}
	// return c.Cluster.Status
}

var ClusterIndices = []repos.IndexField{
	{
		Field:  []repos.IndexKey{{Key: fc.Id, Value: repos.IndexAsc}},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.MetadataName, Value: repos.IndexAsc},
			{Key: fc.MetadataNamespace, Value: repos.IndexAsc},
			{Key: fc.AccountName, Value: repos.IndexAsc},
			{Key: fc.ClusterGlobalVPN, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.ClusterSpecClusterServiceCIDR, Value: repos.IndexAsc},
			{Key: fc.ClusterGlobalVPN, Value: repos.IndexAsc},
			{Key: fc.AccountName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.ClusterOwnedBy, Value: repos.IndexAsc},
		},
		Unique: false,
	},
}
