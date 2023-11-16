package device

type deviceConfig struct {
	DeviceIp        string
	DevicePvtKey    string
	ServerPublicKey string
	ServerEndpoint  string
	DNS             string
	PodCidr         string
	SvcCidr         string
}

type Peer struct {
	PublicKey  string
	AllowedIps string
}

type Data struct {
	ServerIp         string
	ServerPrivateKey string
	Peers            []Peer
}
