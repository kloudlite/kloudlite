package entities

import (
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

type OperationType string

const (
	OperationReplace OperationType = "replace"
	OperationRemove  OperationType = "remove"
	OperationAdd     OperationType = "add"
)

// const (
// 	ResourceApp    ResourceType = "app"
// 	ResourceConfig ResourceType = "config"
// 	ResourceSecret ResourceType = "secret"
// 	ResourceMres   ResourceType = "mres"
// 	ResourceMsvc   ResourceType = "msvc"
// )

// type Overrides struct {
// 	Operation OperationType `bson:"op" json:"op"`
// 	Path      string        `bson:"path" json:"path"`
// 	Value     bson.      `bson:"value,omitempty" json:"value,omitempty"`
// }

type ResInstance struct {
	repos.BaseEntity `bson:",inline"`
	Overrides        string   `bson:"overrides,omitempty" json:"overrides,omitempty"`
	ResourceId       repos.ID `bson:"resource_id" json:"resource_id"`
	EnvironmentId    repos.ID `bson:"environment_id" json:"environment_id"`
	// blueprint_id is project_id
	BlueprintId  *repos.ID           `bson:"blueprint_id,omitempty" json:"blueprint_id,omitempty"`
	ResourceType common.ResourceType `bson:"resource_type" json:"resource_type"`

	// ParentEnvironmentId *repos.ID `bson:"parent_environment_id,omitempty" json:"parent_environment_id,omitempty"`
}

var ResourceIndexs = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "resource_id", Value: repos.IndexAsc},
			{Key: "environment_id", Value: repos.IndexAsc},
			{Key: "blueprint_id", Value: repos.IndexAsc},
			// {Key: "instance_type", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
