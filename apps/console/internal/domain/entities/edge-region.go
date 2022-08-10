package entities

import (
	"kloudlite.io/pkg/repos"
)

type EdgeRegion struct {
	repos.BaseEntity `bson:",inline"`
	Name             string   `bson:"name"`
	ProviderId       repos.ID `bson:"provider_id"`
	Region           string   `bson:"region"`
}

type CloudProvider struct {
	repos.BaseEntity `bson:",inline"`
	Name             string    `bson:"name"`
	AccountId        *repos.ID `json:"account_id,omitempty" bson:"account_id"`
	Provider         string    `json:"provider" bson:"provider"`
}

var CloudProviderIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "account_id", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "provider", Value: repos.IndexAsc},
		},
	},
}

var EdgeRegionIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "region", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "provider", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "region", Value: repos.IndexAsc},
			{Key: "provider", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
