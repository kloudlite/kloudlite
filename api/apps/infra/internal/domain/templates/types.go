package templates

import (
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

type ClusterGatewayDeploymentTemplateVars struct {
	metav1.ObjectMeta
	GatewayWgSecretName          string
	GatewayGlobalIP              string
	GatewayDNSSuffix             string
	GatewayInternalDNSNameserver string
	GatewayWgExtraPeersHash      string
	GatewayDNSServers            string

	GatewayServiceType networkingv1.GatewayServiceType
	GatewayNodePort    int32

	ClusterCIDR string
	ServiceCIDR string

	IPManagerConfigName      string
	IPManagerConfigNamespace string

	ImageWebhookServer       string
	ImageIPManager           string
	ImageIPBindingController string
	ImageDNS                 string
	ImageLogsProxy           string
}
