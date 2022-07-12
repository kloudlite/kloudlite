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
	PrivateKey       *string      `json:"private_key" bson:"private_key"`
	PublicKey        *string      `json:"public_key" bson:"public_key"`
	Ip               string       `json:"ip" bson:"ip"`
	Status           DeviceStatus `json:"status" bson:"status"`
}

var DeviceIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
