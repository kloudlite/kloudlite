package entities

import "kloudlite.io/pkg/repos"

type Cluster struct {
	repos.BaseEntity `bson:",inline"`
	Name             string            `json:"name" bson:"name"`
	Address          *string           `json:"address,omitempty" bson:"address,omitempty"`
	ListenPort       *uint16           `json:"listenPort,omitempty" bson:"listenPort,omitempty"`
	PrivateKey       *string           `json:"privateKey,omitempty" bson:"privateKey,omitempty"`
	PublicKey        *string           `json:"publicKey,omitempty" bson:"publicKey,omitempty"`
	Peers            map[repos.ID]Peer `json:"peers,omitempty" bson:"peers,omitempty"`
	NetInterface     *string           `json:"netInterface" bson:"netInterface,omitempty"`
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
