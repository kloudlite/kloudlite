package templates

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WebhookTemplateArgs struct {
	CaBundle         string
	ServiceName      string
	ServiceNamespace string
	ServiceHTTPSPort int
	ServiceSelector  map[string]string

	NamespaceSelector metav1.LabelSelector
}

type ServiceInterceptPodSpecParams struct {
	DeviceHost      string
	TCPPortMappings map[int32]int32
	UDPPortMappings map[int32]int32
}
