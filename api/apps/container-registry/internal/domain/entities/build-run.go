package entities

import (
	"github.com/kloudlite/api/pkg/repos"
	distributionv1 "github.com/kloudlite/operator/apis/distribution/v1"
)

type BuildRun struct {
	repos.BaseEntity        `json:",inline" graphql:"noinput"`
	distributionv1.BuildRun `json:",inline" graphql:"noinput"`
	AccountName             string `json:"accountName" graphql:"noinput"`
	ClusterName             string `json:"clusterName" graphql:"noinput"`
}

var BuildRunIndices = []repos.IndexField{
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
