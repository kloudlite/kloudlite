package entities

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/pkg/repos"
)

type Environment struct {
	repos.BaseEntity `bson:",inline"`
	// blueprint_id is project_id
	BlueprintId repos.ID           `bson:"blueprint_id,omitempty" json:"blueprint_id,omitempty"`
	Name        string             `bson:"name,omitempty" json:"name,omitempty"`
	ReadableId  string             `json:"readable_id" bson:"readable_id"`
	Status      ProjectStatus      `json:"status" bson:"status"`
	Conditions  []metav1.Condition `json:"conditions" bson:"conditions"`
	IsDeleted   bool               `json:"is_deleted" bson:"is_deleted"`

	// ParentEnvironmentId *repos.ID `bson:"parent_environment_id,omitempty" json:"parent_environment_id,omitempty"`
}

var EnvironmentIndexs = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "readable_id", Value: repos.IndexAsc},
			{Key: "blueprint_id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
