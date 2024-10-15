package entities

import (
	fc "github.com/kloudlite/api/apps/message-office/internal/entities/field-constants"
	"github.com/kloudlite/api/pkg/repos"
)

type ClusterAllocationClusterRef struct {
	Name           string `json:"name"`
	Region         string `json:"region"`
	OwnedByAccount string `json:"owned_by_account"`
	PublicDNSHost  string `json:"public_dns_host"`
}

type ClusterAllocation struct {
	repos.BaseEntity `json:",inline"`
	To               string                      `json:"to"`
	Cluster          ClusterAllocationClusterRef `json:"cluster"`
}

var ClusterAllocationIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fc.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.ClusterAllocationTo, Value: repos.IndexAsc},
			{Key: fc.ClusterAllocationClusterName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.ClusterAllocationClusterRegion, Value: repos.IndexAsc},
		},
	},
}
