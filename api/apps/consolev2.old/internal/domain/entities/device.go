package entities

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/pkg/repos"
)

type DeviceStatus string

const (
	DeviceStateSyncing  = DeviceStatus("sync-in-progress")
	DeviceStateError    = DeviceStatus("error")
	DeviceStateAttached = DeviceStatus("attached")
	DeviceStateDeleted  = DeviceStatus("deleted")
)

type Port struct {
	Port       int32  `json:"port" bson:"port"`
	TargetPort *int32 `json:"target_port" bson:"target_port"`
}

type Device struct {
	repos.BaseEntity `bson:",inline"`
	Index            int                `json:"index" bson:"index"`
	Name             string             `json:"name" bson:"name"`
	AccountId        repos.ID           `json:"account_id" bson:"account_id"`
	UserId           repos.ID           `json:"user_id" bson:"user_id"`
	Status           DeviceStatus       `json:"status" bson:"status"`
	ActiveRegion     *string            `json:"region" bson:"region"`
	ExposedPorts     []Port             `json:"exposed_ports" bson:"exposed_ports"`
	Conditions       []metav1.Condition `json:"conditions" bson:"conditions"`
}

var DeviceIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "account_id", Value: repos.IndexAsc},
			{Key: "name", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
