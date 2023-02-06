package entities

import (
	"encoding/json"
	"io"

	clusterV1 "github.com/kloudlite/cluster-operator/api/v1"
	"kloudlite.io/pkg/repos"
)

type Cluster struct {
	repos.BaseEntity  `bson:",inline" json:",inline"`
	clusterV1.Cluster `json:",inline"`
	// Name             string   `json:"name,omitempty"`
	// AccountId        repos.ID `json:"accountId,omitempty"`
	// SubDomain        string   `json:"subDomain,omitempty"`
	// KubeConfig       string   `json:"kubeConfig,omitempty"`
}

func (c *Cluster) UnmarshalGQL(v interface{}) error {
	switch t := v.(type) {
	case map[string]any:
		b, err := json.Marshal(t)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(b, c); err != nil {
			return err
		}

	case string:
		if err := json.Unmarshal([]byte(t), c); err != nil {
			return err
		}
	}

	return nil
}

func (c Cluster) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(c)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

var ClusterIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "accountId", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
