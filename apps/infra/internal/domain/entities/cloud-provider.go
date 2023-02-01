package entities

import (
	"encoding/json"
	"io"

	infrav1 "github.com/kloudlite/internal_operator_v2/apis/infra/v1"
	"kloudlite.io/pkg/repos"
)

type ProviderSecrets map[string]string

func (ps *ProviderSecrets) UnmarshalGQL(v interface{}) error {
	switch t := v.(type) {
	case map[string]any:
		b, err := json.Marshal(t)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(b, ps); err != nil {
			return err
		}

	case string:
		if err := json.Unmarshal([]byte(t), ps); err != nil {
			return err
		}
	}

	return nil
}

func (ps ProviderSecrets) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(ps)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

type CloudProvider struct {
	repos.BaseEntity      `bson:",inline" json:",inline"`
	infrav1.CloudProvider `bson:",inline" json:",inline"`
}

func (cp *CloudProvider) UnmarshalGQL(v interface{}) error {
	switch t := v.(type) {
	case map[string]any:
		b, err := json.Marshal(t)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(b, cp); err != nil {
			return err
		}

	case string:
		if err := json.Unmarshal([]byte(t), cp); err != nil {
			return err
		}
	}

	return nil
}

func (cp CloudProvider) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(cp)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

var CloudProviderIndices = []repos.IndexField{
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
