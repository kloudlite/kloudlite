package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/pkg/operator"
)

type GlobalVPN struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	// wgv1.ClusterConnection `json:",inline"`
	wgv1.GlobalVPN `json:",inline"`

	common.ResourceMetadata `json:",inline"`

	AccountName           string `json:"accountName" graphql:"noinput"`
	ClusterName           string `json:"clusterName" graphql:"noinput"`
	ClusterPublicEndpoint string `json:"clusterPublicEndpoint" graphql:"noinput"`

	CIDR                  string `json:"cidr" graphql:"noinput"`
	AllocatableCIDRSuffix int    `json:"allocatableCIDRSuffix" graphql:"noinput"`
	ClusterOffset         int    `json:"clusterOffset" graphql:"noinput"`

	GatewayIPAddr string `json:"gatewayIPAddr" graphql:"ignore"`

	ParsedWgParams *wgv1.WgParams `json:"parsedWgParams" graphql:"ignore"`

	SyncStatus t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (c *GlobalVPN) GetDisplayName() string {
	return c.ResourceMetadata.DisplayName
}

func (c *GlobalVPN) GetStatus() operator.Status {
	return c.GlobalVPN.Status
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
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
		},
	},
}
