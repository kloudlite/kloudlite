package entities

import (
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

type BYOCCluster struct {
	repos.BaseEntity `bson:",inline" json:",inline"`
	Name             string       `json:"clusterName"`
	DisplayName      string       `json:"displayName"`
	AccountName      string       `json:"accountName"`
	Region           string       `json:"region"`
	Provider         string       `json:"provider"`
	IsConnected      bool         `json:"isConnected"`
	SyncStatus       t.SyncStatus `json:"syncStatus"`
}

var BYOCClusterIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "name", Value: repos.IndexAsc},
			{Key: "accountName", Value: repos.IndexAsc},
			{Key: "region", Value: repos.IndexAsc},
			{Key: "provider", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "name", Value: repos.IndexAsc},
		},
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
		},
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
}
