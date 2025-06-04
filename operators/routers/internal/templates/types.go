package templates

import (
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type IngressTemplateArgs struct {
	Metadata         metav1.ObjectMeta
	IngressClassName string

	HttpsEnabled bool

	WildcardDomains    []string
	NonWildcardDomains []string

	Routes map[string][]crdsv1.Route
}
