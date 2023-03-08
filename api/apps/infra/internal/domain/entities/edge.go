package entities

import (
	"encoding/json"
	"io"

	infraV1 "github.com/kloudlite/cluster-operator/apis/infra/v1"
	"kloudlite.io/pkg/repos"
)

type Edge struct {
	repos.BaseEntity `bson:",inline"`
	infraV1.Edge     `json:",inline" bson:",inline"`
}

func (edge *Edge) UnmarshalGQL(v interface{}) error {
	switch t := v.(type) {
	case map[string]any:
		b, err := json.Marshal(t)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(b, edge); err != nil {
			return err
		}

	case string:
		if err := json.Unmarshal([]byte(t), edge); err != nil {
			return err
		}
	}

	return nil
}

func (edge Edge) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(edge)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

var EdgeIndices = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "metadata.name", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
