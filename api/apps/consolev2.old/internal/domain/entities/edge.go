package entities

import (
	"encoding/json"
	"io"

	infraV1 "github.com/kloudlite/cluster-operator/apis/infra/v1"
	"kloudlite.io/pkg/repos"
)

type EdgeStatus string

const (
	EdgeStateSyncing = EdgeStatus("sync-in-progress")
	EdgeStateLive    = EdgeStatus("live")
	EdgeStateError   = EdgeStatus("error")
	EdgeStateDown    = EdgeStatus("down")
)

type NodePool struct {
	Name   string   `json:"name"`
	Config string   `json:"config"`
	Min    int      `json:"min"`
	Max    int      `json:"max"`
	Nodes  []string `bson:"nodes"`
}

type EdgeRegion struct {
	repos.BaseEntity `bson:",inline"`
	infraV1.Edge     `json:",inline" bson:",inline"`

	// IsDeleting       bool               `json:"is_deleting" bson:"is_deleting"`
	// Name             string             `bson:"name"`
	// ProviderId       repos.ID           `bson:"provider_id"`
	// Region           string             `bson:"region"`
	// Pools            []NodePool         `bson:"pools"`
	// Status           EdgeStatus         `json:"status" bson:"status"`
	// Conditions       []metav1.Condition `json:"conditions" bson:"conditions"`
}

func (er *EdgeRegion) UnmarshalGQL(v interface{}) error {
	switch res := v.(type) {
	case map[string]any:
		b, err := json.Marshal(res)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, er); err != nil {
			return err
		}
	case string:
		if err := json.Unmarshal([]byte(v.(string)), er); err != nil {
			return err
		}
	}
	return nil
}

func (er EdgeRegion) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(er)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

var EdgeRegionIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "region", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "provider", Value: repos.IndexAsc},
		},
	},
}
