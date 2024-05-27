package templates

type GVPNKloudliteDeviceTemplateVars struct {
	Name      string
	Namespace string
	WgConfig  string

	KubeReverseProxyImage string
	AuthzToken            string
}
