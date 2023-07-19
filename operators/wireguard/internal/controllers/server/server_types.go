package server

type Peer struct {
	PublicKey  string
	AllowedIps string
}

type data struct {
	ServerIp         string
	ServerPrivateKey string
	Peers            []Peer
}

type ConfigService struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	ServicePort int32  `json:"servicePort"`
	ProxyPort   int32  `json:"proxyPort"`
}
