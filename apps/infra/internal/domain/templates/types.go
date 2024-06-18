package templates

type GVPNKloudliteDeviceTemplateVars struct {
	Name          string
	Namespace     string
	WgConfig      string
	WireguardPort uint16

	KubeReverseProxyImage string
	AuthzToken            string

	GatewayDNSServers   string
	GatewayServiceHosts string
}
