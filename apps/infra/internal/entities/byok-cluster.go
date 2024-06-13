package entities

import (
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterVisibilityMode string

const (
	ClusterVisibilityModePublic  ClusterVisibilityMode = "public"
	ClusterVisibilityModePrivate ClusterVisibilityMode = "private"
)

type ClusterVisbility struct {
	Mode           ClusterVisibilityMode `json:"mode"`
	PublicEndpoint *string               `json:"publicEndpoint" graphql:"noinput"`
}

type BYOKCluster struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	metav1.ObjectMeta `json:"metadata"`

	GlobalVPN      string `json:"globalVPN" graphql:"noinput"`
	ClusterSvcCIDR string `json:"clusterSvcCIDR" graphql:"noinput"`

	Visibility ClusterVisbility `json:"visibility"`

	ClusterToken            string `json:"clusterToken" graphql:"noinput"`
	MessageQueueTopicName   string `json:"messageQueueTopicName" graphql:"noinput"`
	common.ResourceMetadata `json:",inline"`

	SyncStatus  t.SyncStatus `json:"syncStatus" graphql:"noinput"`
	AccountName string       `json:"accountName" graphql:"noinput"`

	// to be set post sync
	Kubeconfig t.EncodedString `json:"kubeconfig" graphql:"ignore"`
}

func (c *BYOKCluster) GetDisplayName() string {
	return c.DisplayName
}

func (c *BYOKCluster) GetStatus() operator.Status {
	return operator.Status{
		IsReady: true,
		Checks:  map[string]operator.Check{},
	}
}

var BYOKClusterIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: fc.Id, Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: fc.MetadataName, Value: repos.IndexAsc},
			{Key: fc.AccountName, Value: repos.IndexAsc},
		},
		Unique: true,
	},
}

func UniqueBYOKClusterFilter(accountName string, clusterName string) repos.Filter {
	return repos.Filter{
		fc.AccountName:  accountName,
		fc.MetadataName: clusterName,
	}
}

func ListBYOKClusterFilter(accountName string) repos.Filter {
	return repos.Filter{
		fc.AccountName: accountName,
	}
}
