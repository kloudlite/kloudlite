package entities

import (
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	"github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BYOKCluster struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	*clustersv1.ClusterSpec `json:",inline" graphql:"noinput"`

	metav1.ObjectMeta `json:"metadata"`

	common.ResourceMetadata `json:",inline"`

	SyncStatus  t.SyncStatus `json:"syncStatus" graphql:"noinput"`
	AccountName string       `json:"accountName" graphql:"noinput"`
}

func (c *BYOKCluster) GetDisplayName() string {
	return c.ResourceMetadata.DisplayName
}

func (c *BYOKCluster) GetStatus() operator.Status {
	return operator.Status{
		IsReady: true,
		// Resources:           []operator.ResourceRef{},
		// Message:             &raw_json.RawJson{},
		// CheckList:           []operator.CheckMeta{},
		Checks: map[string]operator.Check{},
		// LastReadyGeneration: 0,
		// LastReconcileTime:   &metav1.Time{},
	}
}

var BYOKClusterIndices = []repos.IndexField{
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
}
