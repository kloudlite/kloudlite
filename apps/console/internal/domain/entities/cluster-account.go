package entities

import "kloudlite.io/pkg/repos"

type ClusterAccountStatus string

type WGAccount struct {
	repos.BaseEntity `bson:",inline"`
	AccountID        repos.ID             `bson:"account_id"`
	AccessDomain     string               `bson:"access_domain"`
	WgPort           string               `bson:"wg_port"`
	WgPubKey         string               `bson:"wg_pub_key"`
	WgPrivateKey     string               `bson:"wg_pvt_key"`
	Status           ClusterAccountStatus `json:"status" bson:"status"`
}

var WGAccountIndexes = []repos.IndexField{
	{
		Field: []repos.IndexKey{
			{Key: "id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
	{
		Field: []repos.IndexKey{
			{Key: "account_id", Value: repos.IndexAsc},
		},
		Unique: true,
	},
}
