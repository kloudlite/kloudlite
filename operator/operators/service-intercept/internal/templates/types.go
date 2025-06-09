package templates

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HPATemplateVars struct {
	Metadata metav1.ObjectMeta
	*crdsv1.HPA
}

type WebhookTemplateArgs struct {
	CaBundle          string
	ServiceName       string
	ServiceNamespace  string
	ServiceHTTPSPort  uint16
	ServiceSelector   map[string]string
	NamespaceSelector metav1.LabelSelector
}
