package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/pkg/operator"
)

type ClusterConnection struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	wgv1.ClusterConnection `json:",inline"`

	common.ResourceMetadata `json:",inline"`

	AccountName      string `json:"accountName" graphql:"noinput"`
	ClusterName      string `json:"clusterName" graphql:"noinput"`
	ClusterGroupName string `json:"clusterGroupName" graphql:"noinput"`

	CIDR     string `json:"cidr" graphql:"noinput"`
	Endpoint string `json:"endpoint" graphql:"noinput"`

	SyncStatus t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (c *ClusterConnection) GetDisplayName() string {
	return c.ResourceMetadata.DisplayName
}

func (c *ClusterConnection) GetStatus() operator.Status {
	return c.ClusterConnection.Status
}

var ClusterConnIndices = []repos.IndexField{
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
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
			{Key: "spec.id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
		},
	},
}
