package entities

import (
	"encoding/json"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"io"

	infraV1 "github.com/kloudlite/cluster-operator/apis/infra/v1"

	"kloudlite.io/pkg/repos"
)

type Secret struct {
	repos.BaseEntity `json:",inline"`
	crdsv1.Secret    `json:",inline"`
}

//goland:noinspection ALL
func (s *Secret) UnmarshalGQL(v interface{}) error {
	switch t := v.(type) {
	case map[string]any:
		b, err := json.Marshal(t)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(b, s); err != nil {
			return err
		}
	case string:
		if err := json.Unmarshal([]byte(t), s); err != nil {
			return err
		}
	}

	return nil
}

//goland:noinspection ALL
func (s Secret) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(s)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

var SecretIndices = []repos.IndexField{
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
}

type CloudProvider struct {
	repos.BaseEntity      `bson:",inline" json:",inline"`
	infraV1.CloudProvider `bson:",inline" json:",inline"`
}

//goland:noinspection ALL
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

//goland:noinspection ALL
func (cp CloudProvider) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(cp)
	if err != nil {
		w.Write([]byte(""))
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
			{Key: "spec.accountId", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
