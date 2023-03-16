package entities

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"

	infraV1 "github.com/kloudlite/cluster-operator/apis/infra/v1"

	"kloudlite.io/pkg/repos"
)

type Secret struct {
	repos.BaseEntity `json:",inline"`
	crdsv1.Secret    `json:",inline"`
	AccountName      string `json:"accountName"`
	ClusterName      string `json:"clusterName"`
}

var SecretIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},

	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "metadata.namespace", Value: repos.IndexAsc},
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
			{Key: "clusterName", Value: repos.IndexAsc},
		},
	},
}

type CloudProvider struct {
	repos.BaseEntity      `bson:",inline" json:",inline"`
	infraV1.CloudProvider `bson:",inline" json:",inline"`
	AccountName           string `json:"accountName"`
	ClusterName           string `json:"clusterName"`
}

var CloudProviderIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "spec.accountId", Value: repos.IndexAsc},
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
			{Key: "clusterName", Value: repos.IndexAsc},
		},
	},
}
