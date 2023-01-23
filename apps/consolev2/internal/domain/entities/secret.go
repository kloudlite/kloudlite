package entities

import (
	"encoding/json"
	"gopkg.in/validator.v2"
	"io"
	corev1 "k8s.io/api/core/v1"
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
	if err := json.Unmarshal([]byte(v.(string)), c); err != nil {
		return err
	}

	if err := validator.Validate(*c); err != nil {
		return err
	}

	return nil
}

func (s SecretData) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(c)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

type Secret struct {
	repos.BaseEntity `bson:",inline"`
	corev1.Secret    `bson:",inline" json:",inline"`

	// ClusterId   repos.ID           `json:"cluster_id" bson:"cluster_id"`
	// ProjectId   repos.ID           `json:"project_id" bson:"project_id"`
	// Name        string             `json:"name" bson:"name"`
	// Namespace   string             `json:"namespace" bson:"namespace"`
	// Description *string            `json:"description" bson:"description"`
	// Data        []*Entry           `json:"data" bson:"data"`
	// Status      SecretStatus       `json:"status" bson:"status"`
	// Conditions  []metav1.Condition `json:"conditions" bson:"conditions"`
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
			// {Key: "cluster_id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
