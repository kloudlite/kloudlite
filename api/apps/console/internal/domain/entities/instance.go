package entities

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
)

type OperationType string
type InstanceStatus string

const (
	OperationReplace OperationType = "replace"
	OperationRemove  OperationType = "remove"
	OperationAdd     OperationType = "add"
)

const (
	InstanceStateSyncing    = InstanceStatus("sync-in-progress")
	InstanceStateRestarting = InstanceStatus("restarting")
	InstanceStateFrozen     = InstanceStatus("frozen")
	InstanceStateDeleting   = InstanceStatus("deleting")
	InstanceStateLive       = InstanceStatus("live")
	InstanceStateError      = InstanceStatus("error")
	InstanceStateDown       = InstanceStatus("down")
)

type ResInstance struct {
	repos.BaseEntity `bson:",inline"`
	Overrides        string              `bson:"overrides,omitempty" json:"overrides,omitempty"`
	ResourceId       repos.ID            `bson:"resource_id" json:"resource_id"`
	EnvironmentId    repos.ID            `bson:"environment_id" json:"environment_id"`
	BlueprintId      *repos.ID           `bson:"blueprint_id,omitempty" json:"blueprint_id,omitempty"` // blueprint_id is project_id
	ResourceType     common.ResourceType `bson:"resource_type" json:"resource_type"`
	Status           InstanceStatus      `json:"status" bson:"status"`
	Conditions       []metav1.Condition  `json:"conditions" bson:"conditions"`
}

var ResourceIndexs = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "resource_id", Value: repos.IndexAsc},
			{Key: "environment_id", Value: repos.IndexAsc},
			{Key: "blueprint_id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
