package templates

import (
	v1 "github.com/kloudlite/operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WebhookProxy struct {
	Enabled          bool
	ServiceName      string
	ServiceNamespace string
}

type WebhookTemplateArgs struct {
	WebhookProxy

	CaBundle         string
	ServiceName      string
	ServiceNamespace string
	ServiceHTTPSPort int
	ServiceSelector  map[string]string

	NamespaceSelector        metav1.LabelSelector
	InterceptorPodLabelKey   string
	InterceptorPodLabelValue string
}

type ServiceInterceptPodSpecParams struct {
	TargetHost   string
	PortMappings []v1.ServiceInterceptPortMappings
}
