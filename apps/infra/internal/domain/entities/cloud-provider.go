package entities

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"

	infraV1 "github.com/kloudlite/cluster-operator/apis/infra/v1"

	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

type Secret struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	crdsv1.Secret    `json:",inline" graphql:"uri=k8s://secrets.crds.kloudlite.io"`
	AccountName      string       `json:"accountName"`
	ClusterName      string       `json:"clusterName"`
	SyncStatus       t.SyncStatus `json:"syncStatus" graphql:"noinput"`
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
	repos.BaseEntity      `json:",inline" graphql:"noinput"`
	infraV1.CloudProvider `json:",inline" graphql:"uri=k8s://cloudproviders.infra.kloudlite.io"`
	AccountName           string       `json:"accountName"`
	ClusterName           string       `json:"clusterName"`
	SyncStatus            t.SyncStatus `json:"syncStatus" graphql:"noinput"`
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
