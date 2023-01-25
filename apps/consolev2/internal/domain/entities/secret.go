package entities

import (
	"encoding/json"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"gopkg.in/validator.v2"
	"io"
	"kloudlite.io/pkg/repos"
)

type SecretStatus string

const (
	SecretStateSyncing = SecretStatus("sync-in-progress")
	SecretStateLive    = SecretStatus("live")
	SecretStateError   = SecretStatus("error")
	SecretStateDown    = SecretStatus("down")
)

type SecretData map[string][]byte

func (s SecretData) ToMap() map[string][]byte {
	return s
}

func (s *SecretData) UnmarshalGQL(v interface{}) error {
	if s == nil {
		return nil
	}
	if err := json.Unmarshal([]byte(v.(string)), s); err != nil {
		return err
	}

	if err := validator.Validate(*s); err != nil {
		return err
	}

	return nil
}

func (s SecretData) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(s)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

type Secret struct {
	repos.BaseEntity `bson:",inline"`
	crdsv1.Secret    `json:",inline" bson:",inline"`

	// ClusterId   repos.ID           `json:"cluster_id" bson:"cluster_id"`
	// ProjectId   repos.ID           `json:"project_id" bson:"project_id"`
	// Name        string             `json:"name" bson:"name"`
	// Namespace   string             `json:"namespace" bson:"namespace"`
	// Description *string            `json:"description" bson:"description"`
	// Data        []*Entry           `json:"data" bson:"data"`
	// Status      SecretStatus       `json:"status" bson:"status"`
	// Conditions  []metav1.Condition `json:"conditions" bson:"conditions"`
}

func (s *Secret) UnmarshalGQL(v interface{}) error {
	if err := json.Unmarshal([]byte(v.(string)), s); err != nil {
		return err
	}

	// if err := validator.Validate(*s); err != nil {
	//  return err
	// }

	return nil
}

func (s Secret) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(s)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

var SecretIndexes = []repos.IndexField{
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
