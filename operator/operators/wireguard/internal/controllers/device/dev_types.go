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
