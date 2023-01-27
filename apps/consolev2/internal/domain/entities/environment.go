package entities

import (
	"encoding/json"
	"fmt"
	"io"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"kloudlite.io/pkg/repos"
)

type Environment struct {
	repos.BaseEntity `bson:",inline"`
	crdsv1.Env       `json:",inline"`
	// blueprint_id is project_id
	//BlueprintId repos.ID           `bson:"blueprint_id,omitempty" json:"blueprint_id,omitempty"`
	//Name        string             `bson:"name,omitempty" json:"name,omitempty"`
	//ReadableId  string             `json:"readable_id" bson:"readable_id"`
	//Status      ProjectStatus      `json:"status" bson:"status"`
	//Conditions  []metav1.Condition `json:"conditions" bson:"conditions"`
	//IsDeleted   bool               `json:"is_deleted" bson:"is_deleted"`

	// ParentEnvironmentId *repos.ID `bson:"parent_environment_id,omitempty" json:"parent_environment_id,omitempty"`
}

func (env *Environment) UnmarshalGQL(v interface{}) error {
	fmt.Println("v type is %T", v)
	if err := json.Unmarshal([]byte(v.(string)), env); err != nil {
		return err
	}

	// if err := validator.Validate(*c); err != nil {
	//  return err
	// }

	return nil
}

func (env Environment) MarshalGQL(w io.Writer) {
	b, err := json.Marshal(env)
	if err != nil {
		w.Write([]byte("{}"))
	}
	w.Write(b)
}

var EnvironmentIndexes = []repos.IndexField{
	{
		Field:  []repos.IndexKey{{Key: "metadata.name", Value: repos.IndexAsc}},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{{Key: "spec.projectName", Value: repos.IndexAsc}},
	},
	{
		Field: []repos.IndexKey{{Key: "spec.accountId", Value: repos.IndexAsc}},
	},
}
