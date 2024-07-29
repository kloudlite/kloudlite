package templates

import (
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GatewayDeploymentArgs struct {
	metav1.ObjectMeta

	ServiceAccountName string

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

type WebhookTemplateArgs struct {
	NamePrefix      string
	Namespace       string
	OwnerReferences []metav1.OwnerReference

	ServiceName string

	WebhookServerImage        string
	WebhookServerCertCABundle string

	WebhookNamespaceSelectorKey string
}

type GatewayRBACTemplateArgs struct {
	metav1.ObjectMeta
}
