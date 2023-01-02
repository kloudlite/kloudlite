package entities

import "kloudlite.io/pkg/repos"

type Environment struct {
	repos.BaseEntity `bson:",inline"`
	// blueprint_id is project_id
	BlueprintId *repos.ID `bson:"blueprint_id,omitempty" json:"blueprint_id,omitempty"`
	Name        string    `bson:"name,omitempty" json:"name,omitempty"`
	ReadableId  string    `json:"readable_id" bson:"readable_id"`

	// ParentEnvironmentId *repos.ID `bson:"parent_environment_id,omitempty" json:"parent_environment_id,omitempty"`
}
