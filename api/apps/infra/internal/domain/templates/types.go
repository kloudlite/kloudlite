package templates

type GVPNKloudliteDeviceTemplateVars struct {
	Name          string
	Namespace     string
	WgConfig      string
	WireguardPort uint16

	KloudliteAccount string

	EnableKubeReverseProxy bool
	KubeReverseProxyImage  string
	AuthzToken             string

	GatewayDNSServers   string
	GatewayServiceHosts string
}

type GatewayServiceTemplateVars struct {
	Name          string
	Namespace     string
	WireguardPort uint16
	Selector      map[string]string
	ServiceType   string
}
