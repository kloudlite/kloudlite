package entities

import "kloudlite.io/pkg/repos"

type Peer struct {
	Id         string   `json:"id" bson:"id"`
	PublicKey  string   `json:"public_key" bson:"public_key"`
	AllowedIPs []string `json:"allowed_i_ps" bson:"allowed_i_ps"`
}

type Device struct {
	Id         repos.ID        `json:"id" bson:"id"`
	Address    string          `json:"address" bson:"address"`
	PrivateKey string          `json:"private_key" bson:"private_key"`
	PublicKey  string          `json:"public_key" bson:"public_key"`
	Peers      map[string]Peer `json:"peers" bson:"peers"`
	AllowedIPs []string        `json:"allowed_ips" bson:"allowed_ips"`
	Endpoint   string          `json:"endpoint" bson:"endpoint"`
}

func (d Device) GetId() repos.ID {
	return d.Id
}

func (d Device) SetId(id repos.ID) repos.Entity {
	d.Id = id
	return d
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
