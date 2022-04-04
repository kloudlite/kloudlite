package entities

import "kloudlite.io/pkg/repos"

type Peer struct {
	Id         repos.ID `json:"id" bson:"id"`
	Address    *string  `json:"address" bson:"address"`
	PublicKey  *string  `json:"public_key" bson:"public_key"`
	AllowedIPs []string `json:"allowed_i_ps" bson:"allowed_i_ps"`
}

type Device struct {
	repos.BaseEntity `bson:",inline"`
	Name             string          `json:"name" bson:"name"`
	ClusterId        repos.ID        `json:"cluster_id" bson:"cluster_id"`
	UserId           repos.ID        `json:"user_id" bson:"user_id"`
	PrivateKey       *string         `json:"private_key" bson:"private_key"`
	PublicKey        *string         `json:"public_key" bson:"public_key"`
	Peers            map[string]Peer `json:"peers" bson:"peers"`
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
