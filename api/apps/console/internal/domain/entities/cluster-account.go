package entities

import "kloudlite.io/pkg/repos"

type ClusterAccountStatus string

const (
	ClusterAccountStateSyncing = ClusterAccountStatus("sync-in-progress")
	ClusterAccountStateLive    = ClusterAccountStatus("live")
	ClusterAccountStateError   = ClusterAccountStatus("error")
	ClusterAccountStateDown    = ClusterAccountStatus("down")
)

type ClusterAccount struct {
	repos.BaseEntity `bson:",inline"`
	ClusterID        repos.ID             `bson:"cluster_id"`
	AccountID        repos.ID             `bson:"account_id"`
	WgIp             string               `bson:"wg_ip"`
	WgPort           string               `bson:"wg_port"`
	WgPubKey         string               `bson:"wg_pub_key"`
	Index            int                  `bson:"index"`
	Status           ClusterAccountStatus `json:"status" bson:"status"`
}

var ClusterAccountIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "cluster_id", Value: repos.IndexAsc},
			{Key: "account_id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
