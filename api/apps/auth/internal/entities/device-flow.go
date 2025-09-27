package entities

import (
	"time"

	"github.com/kloudlite/api/pkg/repos"
)

type DeviceFlow struct {
	repos.BaseEntity `bson:",inline" json:",inline"`
	DeviceCode       string    `bson:"deviceCode" json:"deviceCode"`
	UserCode         string    `bson:"userCode" json:"userCode"`
	ClientID         string    `bson:"clientId" json:"clientId"`
	UserID           string    `bson:"userId,omitempty" json:"userId,omitempty"`
	Authorized       bool      `bson:"authorized" json:"authorized"`
	ExpiresAt        time.Time `bson:"expiresAt" json:"expiresAt"`
}

var DeviceFlowIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "deviceCode", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "userCode", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	// Note: TTL index for expiresAt should be added manually in MongoDB
	// db.device_flows.createIndex({ "expiresAt": 1 }, { expireAfterSeconds: 0 })
}