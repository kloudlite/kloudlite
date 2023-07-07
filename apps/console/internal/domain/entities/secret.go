package entities

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

type Secret struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`
	crdsv1.Secret    `json:",inline" graphql:"uri=k8s://secrets.crds.kloudlite.io"`
	AccountName      string       `json:"accountName" graphql:"noinput"`
	ClusterName      string       `json:"clusterName" graphql:"noinput"`
	SyncStatus       t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

var SecretIndexes = []repos.IndexField{
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
			{Key: "accountName", Value: repos.IndexAsc},
			{Key: "clusterName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
