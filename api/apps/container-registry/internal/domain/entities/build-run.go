package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	distributionv1 "github.com/kloudlite/operator/apis/distribution/v1"
	"github.com/kloudlite/operator/pkg/operator"
)

type BuildRun struct {
	repos.BaseEntity        `json:",inline" graphql:"noinput"`
	distributionv1.BuildRun `json:",inline" graphql:"noinput"`
	BuildId                 repos.ID     `json:"buildId" graphql:"noinput,scalar-type=ID"`
	AccountName             string       `json:"accountName" graphql:"noinput"`
	ClusterName             string       `json:"clusterName" graphql:"noinput"`
	SyncStatus              t.SyncStatus `json:"syncStatus" graphql:"noinput"`
	common.ResourceMetadata `json:",inline"`
}

func (a *BuildRun) GetDisplayName() string {
	return a.ResourceMetadata.DisplayName
}

func (a *BuildRun) GetGeneration() int64 {
	return a.ObjectMeta.Generation
}

func (a *BuildRun) GetStatus() operator.Status {
	return a.BuildRun.Status
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
