package entities

import "kloudlite.io/pkg/repos"

type Device struct {
	repos.BaseEntity `bson:",inline"`
	Name             string   `json:"name" bson:"name"`
	ClusterId        repos.ID `json:"cluster_id" bson:"cluster_id"`
	UserId           repos.ID `json:"user_id" bson:"user_id"`
	PrivateKey       *string  `json:"private_key" bson:"private_key"`
	PublicKey        *string  `json:"public_key" bson:"public_key"`
}

/*
[Interface]
Address =
PrivateKey =
DNS =

[Peer]
PublicKey =
AllowedIPs = <, separated>
Endpoint =
*/
