package entities

import (
	fc "github.com/kloudlite/api/apps/message-office/internal/entities/field-constants"
	"github.com/kloudlite/api/pkg/repos"
)

type PlatformEdgeCluster struct {
	repos.BaseEntity `json:",inline"`
	OwnedByAccount   string `json:"owned_by_account"`
	Name             string `json:"name"`
	Region           string `json:"region"`
	CloudProvider    string `json:"cloud_provider"`
}

var PlatformEdgeClusterIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fc.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.PlatformEdgeClusterName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.PlatformEdgeClusterOwnedByAccount, Value: repos.IndexAsc},
		},
	},
}
