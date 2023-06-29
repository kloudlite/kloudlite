package entities

import (
	artifactsv1 "github.com/kloudlite/operator/apis/artifacts/v1"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

type HarborRobotUser struct {
	repos.BaseEntity              `json:",inline" graphql:"noinput"`
	artifactsv1.HarborUserAccount `json:",inline" graphql:"uri=k8s://harboruseraccounts.artifacts.kloudlite.io"`
	SyncStatus                    t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

var HarborRobotUserIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "metadata.namespace", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "spec.harborProjectName", Value: repos.IndexAsc},
		},
	},
}
