package entities

import (
	"encoding/json"
	"io"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"kloudlite.io/pkg/repos"
)

type ManagedServiceStatus string

const (
	ManagedServiceStateSyncing  = ManagedServiceStatus("sync-in-progress")
	ManagedServiceStateDeleting = ManagedServiceStatus("deleting")
	ManagedServiceStateLive     = ManagedServiceStatus("live")
	ManagedServiceStateError    = ManagedServiceStatus("error")
	ManagedServiceStateDown     = ManagedServiceStatus("down")
)

type ManagedService struct {
	repos.BaseEntity      `bson:",inline" json:",inline"`
	crdsv1.ManagedService `json:",inline" bson:",inline"`
	//ClusterId        repos.ID             `json:"cluster_id" bson:"cluster_id"`
	//ProjectId        repos.ID             `json:"project_id" bson:"project_id"`
	//Name             string               `json:"name" bson:"name"`
	//Namespace        string               `json:"namespace" bson:"namespace"`
	//ServiceType      ManagedServiceType   `json:"service_type" bson:"service_type"`
	//Values           map[string]any       `json:"values" bson:"values"`
	//Status           ManagedServiceStatus `json:"status" bson:"status"`
	//Conditions       []metav1.Condition   `json:"conditions" bson:"conditions"`
}

func (obj *ManagedService) UnmarshalGQL(v interface{}) error {
  if err := json.Unmarshal([]byte(v.(string)), obj); err != nil {
    return err
  }

  // if err := validator.Validate(*objobj); err != nil {
  //  return err
  // }

  return nil
}

func (obj ManagedService) MarshalGQL(w io.Writer) {
  b, err := json.Marshal(obj)
  if err != nil {
    w.Write([]byte("{}"))
  }
  w.Write(b)
}

var ManagedServiceIndexes = []repos.IndexField{
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
	//{
	//	Field: []repos.IndexKey{
	//		{Key: "name", Value: repos.IndexAsc},
	//		{Key: "namespace", Value: repos.IndexAsc},
	//		{Key: "cluster_id", Value: repos.IndexAsc},
	//	},
	//	Unique: true,
	//},
}
