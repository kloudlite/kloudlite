package entities

import "kloudlite.io/pkg/repos"

type Cluster struct {
	Id           repos.ID        `json:"id" bson:"id"`
	Name         string          `json:"name" bson:"name"`
	Address      string          `json:"address" bson:"address"`
	ListenPort   uint16          `json:"listenPort" bson:"listenPort"`
	PrivateKey   string          `json:"privateKey" bson:"privateKey"`
	Peers        map[string]Peer `json:"peers" bson:"peers"`
	NetInterface string          `json:"netInterface" bson:"NetInterface"`
}

/*
	[Interface]
	Address =
	SaveConfig = true
	ListenPort =
	PrivateKey =
	PostUp = iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o {{ .NetInterface }} -j MASQUERADE; ip6tables -A FORWARD -i wg0 -j ACCEPT; ip6tables -t nat -A POSTROUTING -o {{ .NetInterface }} -j MASQUERADE
	PostDown = iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o {{ .NetInterface }} -j MASQUERADE; ip6tables -D FORWARD -i wg0 -j ACCEPT; ip6tables -t nat -D POSTROUTING -o {{ .NetInterface }} -j MASQUERADE

	[Peer...]
	PublicKey =
	AllowedIPs =
*/

func (c Cluster) GetId() repos.ID {
	return c.Id
}

func (c Cluster) SetId(id repos.ID) repos.Entity {
	c.Id = id
	return c
}
