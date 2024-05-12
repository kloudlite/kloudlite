package entities

import (
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Peer struct {
	Id         int      `json:"id" graphql:"noinput"`
	PubKey     string   `json:"pubKey" graphql:"noinput"`
	AllowedIps []string `json:"allowedIps" graphql:"noinput"`
}

type GlobalVPN struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	common.ResourceMetadata `json:",inline"`
	metav1.ObjectMeta       `json:"metadata"`

	// like 10.0.0.0/8
	CIDR string `json:"CIDR"`
	// to allocate 8K IPs for each GlobalVPNConnection
	// i.e. pow(2, 13) Ips, it means 13 Host bits,
	// which leaves us with (32 - 13) 19 Network Bits. It is our AllocatableCIDRSuffix
	AllocatableCIDRSuffix int    `json:"allocatableCIDRSuffix"`
	WgInterface           string `json:"wgInterface"`

	NumReservedIPsForNonClusterUse int `json:"numReservedIPsForNonClusterUse"`

	// Running Count of allocated Cluster CIDRs for clusters, under this GlobalVPN
	NumAllocatedClusterCIDRs int `json:"numAllocatedClusterCIDRs"`

	// Running Count for allocated Devices under this GlobalVPN
	// It will always be <= NumReservedIPsForNonClusterUse
	NumAllocatedDevices int `json:"numAllocatedDevices"`

	// Peers []Peer `json:"peers" graphql:"noinput"`
	AccountName string `json:"accountName" graphql:"noinput"`

	KloudliteDevice struct {
		Name   string `json:"name"`
		IPAddr string `json:"ipAddr"`
	} `json:"kloudliteDevice"`
}

func (c *GlobalVPN) GetDisplayName() string {
	return c.ResourceMetadata.DisplayName
}

var GlobalVPNIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fc.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.MetadataName, Value: repos.IndexAsc},
			{Key: fc.AccountName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
