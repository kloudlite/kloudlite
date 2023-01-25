package entities

import (
	"encoding/json"
	"io"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"kloudlite.io/pkg/repos"
)

type ConfigStatus string

const (
	ConfigStateSyncing = ConfigStatus("sync-in-progress")
	ConfigStateLive    = ConfigStatus("live")
	ConfigStateError   = ConfigStatus("error")
	ConfigStateDown    = ConfigStatus("down")
)

type Entry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Config struct {
	repos.BaseEntity `bson:",inline"`
	crdsv1.Config    `json:",inline" bson:",inline"`
	//ClusterId        repos.ID           `json:"cluster_id" bson:"cluster_id"`
	//ProjectId        repos.ID           `json:"project_id" bson:"project_id"`
	//Namespace        string             `json:"namespace" bson:"namespace"`
	//Description      *string            `json:"description" bson:"description"`
	//Name             string             `json:"name" bson:"name"`
	//Data             []*Entry           `json:"data" bson:"data"`
	//Status           ConfigStatus       `json:"status" bson:"status"`
	//Conditions       []metav1.Condition `json:"conditions" bson:"conditions"`
}

func (c *Config) UnmarshalGQL(v interface{}) error {
	if err := json.Unmarshal([]byte(v.(string)), c); err != nil {
		return err
	}

	// if err := validator.Validate(*c); err != nil {
	//  return err
	// }

	return nil
}

func (c Config) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(c)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

var ConfigIndexes = []repos.IndexField{
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
