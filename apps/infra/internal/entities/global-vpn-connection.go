package entities

import (
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/pkg/operator"
)

type GlobalVPNConnDeviceRef struct {
	Name   string `json:"name"`
	IPAddr string `json:"ipAddr"`
}

type WgParams struct {
	WgPrivateKey string `json:"wg_private_key"`
	WgPublicKey  string `json:"wg_public_key"`

	IP string `json:"ip"`

	DNSServer *string `json:"dnsServer"`

	PublicGatewayHosts *string `json:"publicGatewayHosts,omitempty"`
	PublicGatewayPort  *string `json:"publicGatewayPort,omitempty"`

	VirtualCidr string `json:"virtualCidr"`
}

type GlobalVPNConnection struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	// wgv1.GlobalVPN       `json:",inline"`
	networkingv1.Gateway `json:",inline"`

	GlobalVPNName string `json:"globalVPNName"`

	common.ResourceMetadata `json:",inline"`

	AccountName string `json:"accountName" graphql:"noinput"`
	ClusterName string `json:"clusterName" graphql:"noinput"`
	ClusterCIDR string `json:"clusterSvcCIDR" graphql:"noinput"`

	Visibility ClusterVisbility `json:"visibility" graphql:"noinput"`

	// ClusterPublicEndpoint string                 `json:"clusterPublicEndpoint" graphql:"noinput"`
	DeviceRef GlobalVPNConnDeviceRef `json:"deviceRef" graphql:"noinput"`

	// ParsedWgParams *wgv1.WgParams `json:"parsedWgParams" graphql:"ignore"`
	ParsedWgParams *networkingv1.WireguardKeys `json:"parsedWgParams" graphql:"ignore"`
	SyncStatus     t.SyncStatus                `json:"syncStatus" graphql:"noinput"`
}

func (c *GlobalVPNConnection) GetDisplayName() string {
	return c.DisplayName
}

func (c *GlobalVPNConnection) GetStatus() operator.Status {
	return c.Status
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
