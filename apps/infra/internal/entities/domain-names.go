package entities

import (
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

type DomainEntry struct {
	repos.BaseEntity        `json:",inline" graphql:"noinput"`
	common.ResourceMetadata `json:",inline"`

	DomainName string `json:"domainName"`

	AccountName string `json:"accountName" graphql:"noinput"`
	ClusterName string `json:"clusterName"`
}

var DomainEntryIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},

	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
		},
	},

	{
		Field: []repos.IndexKey{
			{Key: "domainName", Value: repos.IndexAsc},
			{Key: "clusterName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
