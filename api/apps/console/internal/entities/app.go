package entities

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

type App struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	crdsv1.App `json:",inline" graphql:"uri=k8s://apps.crds.kloudlite.io"`

	common.ResourceMetadata `json:",inline"`

	AccountName   string `json:"accountName" graphql:"noinput"`
	ClusterName   string `json:"clusterName" graphql:"noinput"`
	ProjectName   string `json:"projectName" graphql:"noinput"`
	WorkspaceName string `json:"workspaceName" graphql:"noinput"`

	SyncStatus t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

var AppIndexes = []repos.IndexField{
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
			{Key: "clusterName", Value: repos.IndexAsc},
			{Key: "accountName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
