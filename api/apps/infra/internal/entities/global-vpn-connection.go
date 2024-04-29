package entities

import (
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/pkg/operator"
)

type GlobalVPNConnection struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	wgv1.GlobalVPN `json:",inline"`

	GlobalVPNName string `json:"globalVPNName"`

	common.ResourceMetadata `json:",inline"`

	AccountName           string `json:"accountName" graphql:"noinput"`
	ClusterName           string `json:"clusterName" graphql:"noinput"`
	ClusterSvcCIDR        string `json:"clusterSvcCIDR" graphql:"noinput"`
	ClusterPublicEndpoint string `json:"clusterPublicEndpoint" graphql:"noinput"`

	GatewayIPAddr string `json:"gatewayIPAddr" graphql:"ignore"`

	ParsedWgParams *wgv1.WgParams `json:"parsedWgParams" graphql:"ignore"`

	SyncStatus t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (c *GlobalVPNConnection) GetDisplayName() string {
	return c.ResourceMetadata.DisplayName
}

func (c *GlobalVPNConnection) GetStatus() operator.Status {
	return c.GlobalVPN.Status
}

var GlobalVPNConnectionIndices = []repos.IndexField{
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

type FreeClusterSvcCIDR struct {
	repos.BaseEntity `json:",inline"`

	AccountName   string `json:"accountName"`
	GlobalVPNName string `json:"globalVPNName"`

	ClusterSvcCIDR string `json:"clusterSvcCIDR"`
}

var FreeClusterSvcCIDRIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fc.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.FreeClusterSvcCIDRClusterSvcCIDR, Value: repos.IndexAsc},
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fc.FreeClusterSvcCIDRGlobalVPNName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

type ClaimClusterSvcCIDR struct {
	repos.BaseEntity `json:",inline"`

	AccountName   string `json:"accountName"`
	GlobalVPNName string `json:"globalVPNName"`

	ClusterSvcCIDR   string `json:"clusterSvcCIDR"`
	ClaimedByCluster string `json:"claimedByCluster"`
}

var ClaimClusterSvcCIDRIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fc.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.ClaimClusterSvcCIDRClusterSvcCIDR, Value: repos.IndexAsc},
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fc.ClaimClusterSvcCIDRGlobalVPNName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.ClaimClusterSvcCIDRClaimedByCluster, Value: repos.IndexAsc},
			{Key: fields.AccountName, Value: repos.IndexAsc},
			{Key: fc.ClaimClusterSvcCIDRGlobalVPNName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
