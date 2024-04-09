package entities

import (
	"fmt"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/pkg/operator"
)

const (
	wgIpIndex       = 16
	clusterPodIndex = 13
)

func GetCidrRanges(index int) (*string, error) {
	switch index {
	case wgIpIndex, clusterPodIndex:
		return nil, fmt.Errorf("it can't be %d or %d", wgIpIndex, clusterPodIndex)
	}

	if index < 0 || index > 255 {
		return nil, fmt.Errorf("ip range can only be between 0 and 255")
	}

	return functions.New(fmt.Sprintf("10.%d.0.0/16", index)), nil
}

type Peer struct {
	Id         int      `json:"id" graphql:"noinput"`
	PubKey     string   `json:"pubKey" graphql:"noinput"`
	AllowedIps []string `json:"allowedIps" graphql:"noinput"`
}

type ClusterGroup struct {
	repos.BaseEntity `json:",inline" graphql:"noinput"`

	common.ResourceMetadata `json:",inline"`

	// Peers []Peer `json:"peers" graphql:"noinput"`

	AccountName string `json:"accountName" graphql:"noinput"`
	ClusterName string `json:"clusterName" graphql:"noinput"`

	SyncStatus t.SyncStatus `json:"syncStatus" graphql:"noinput"`
}

func (c *ClusterGroup) GetDisplayName() string {
	return c.ResourceMetadata.DisplayName
}

func (c *ClusterGroup) GetStatus() operator.Status {
	return operator.Status{
		IsReady: true,
	}
}

var ClusterGroupIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
			{Key: "accountName", Value: repos.IndexAsc},
			{Key: "clusterName", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
			{Key: "spec.id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountName", Value: repos.IndexAsc},
		},
	},
}
