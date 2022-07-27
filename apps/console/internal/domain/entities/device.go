package entities

import "kloudlite.io/pkg/repos"

type DeviceStatus string

const (
	DeviceStateSyncing  = DeviceStatus("sync-in-progress")
	DeviceStateError    = DeviceStatus("error")
	DeviceStateAttached = DeviceStatus("attached")
	DeviceStateDeleted  = DeviceStatus("deleted")
)

type Device struct {
	repos.BaseEntity `bson:",inline"`
	Index            int          `json:"index" bson:"index"`
	Name             string       `json:"name" bson:"name"`
	AccountId        repos.ID     `json:"account_id" bson:"account_id"`
	UserId           repos.ID     `json:"user_id" bson:"user_id"`
	Status           DeviceStatus `json:"status" bson:"status"`
	ActiveRegion     *string      `json:"region" bson:"region"`
	ExposedPorts     []int32      `json:"exposed_ports" bson:"exposed_ports"`
}

var DeviceIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
