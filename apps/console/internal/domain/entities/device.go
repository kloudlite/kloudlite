package entities

import "kloudlite.io/pkg/repos"

type DeviceStatus string

const (
	DeviceStateSyncing  = DeviceStatus("sync-in-progress")
	DeviceStateAttached = DeviceStatus("attached")
)

type Device struct {
	repos.BaseEntity `bson:",inline"`
	Index            int          `json:"index" bson:"index"`
	Name             string       `json:"name" bson:"name"`
	ClusterId        repos.ID     `json:"cluster_id" bson:"cluster_id"`
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
	{
		Field: []repos.IndexKey{
			{Key: "cluster_id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "user_id", Value: repos.IndexAsc},
		},
		Unique: false,
	},
	{
		Field: []repos.IndexKey{
			{Key: "ip", Value: repos.IndexAsc},
			{Key: "cluster_id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "index", Value: repos.IndexAsc},
			{Key: "cluster_id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
