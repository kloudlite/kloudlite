package v1

import (
	"github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RouterHttps struct {
	Enabled       *bool  `json:"enabled"`
	ForceRedirect bool   `json:"forceRedirect,omitempty"`
	ClusterIssuer string `json:"clusterIssuer,omitempty"`
}

func (h *RouterHttps) IsEnabled() bool {
	return h != nil && (h.Enabled == nil || *h.Enabled)
}

type RouterBasicAuth struct {
	Enabled    *bool  `json:"enabled"`
	Username   string `json:"username,omitempty"`
	SecretName string `json:"secretName,omitempty"`
}

func (h *RouterBasicAuth) IsEnabled() bool {
	return h != nil && (h.Enabled == nil || *h.Enabled)
}

type RouterCors struct {
	Enabled          *bool    `json:"enabled,omitempty"`
	Origins          []string `json:"origins,omitempty"`
	AllowCredentials bool     `json:"allowCredentials,omitempty"`
}

func (h *RouterCors) IsEnabled() bool {
	return h != nil && (h.Enabled == nil || *h.Enabled)
}

type RouterRateLimit struct {
	Enabled     *bool `json:"enabled,omitempty"`
	Rps         int   `json:"rps,omitempty"`
	Rpm         int   `json:"rpm,omitempty"`
	Connections int   `json:"connections,omitempty"`
}

func (h *RouterRateLimit) IsEnabled() bool {
	return h != nil && (h.Enabled == nil || *h.Enabled)
}

type RouterRoute struct {
	Host    string `json:"host"`
	Service string `json:"service"`
	Path    string `json:"path"`
	Port    uint16 `json:"port"`

	// +kubebuilder:default=false
	Rewrite bool `json:"rewrite,omitempty"`
}

// RouterSpec defines the desired state of Router.
type RouterSpec struct {
	IngressClass    string       `json:"ingressClass,omitempty"`
	BackendProtocol *string      `json:"backendProtocol,omitempty"`
	Https           *RouterHttps `json:"https,omitempty"`

	RateLimit       *RouterRateLimit `json:"rateLimit,omitempty"`
	MaxBodySizeInMB *int             `json:"maxBodySizeInMB,omitempty"`

	BasicAuth *RouterBasicAuth `json:"basicAuth,omitempty"`
	Cors      *RouterCors      `json:"cors,omitempty"`

	// NginxIngressAnnotations is additional list of annotations on ingress resource
	// INFO: must be used when router does not have direct support in spec
	NginxIngressAnnotations map[string]string `json:"nginxIngressAnnotations,omitempty"`

	Routes []RouterRoute `json:"routes,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.operator\\.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.operator\\.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Router is the Schema for the routers API.
type Router struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouterSpec        `json:"spec,omitempty"`
	Status reconciler.Status `json:"status,omitempty"`
}

func (r *Router) EnsureGVK() {
	if r != nil {
		r.SetGroupVersionKind(GroupVersion.WithKind("Router"))
	}
}

func (r *Router) GetStatus() *reconciler.Status {
	return &r.Status
}

func (r *Router) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (m *Router) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

func (m *Router) IsBasicAuthEnabled() bool {
	return m.Spec.BasicAuth != nil && (m.Spec.BasicAuth.Enabled == nil || *m.Spec.BasicAuth.Enabled)
}

// +kubebuilder:object:root=true

// RouterList contains a list of Router.
type RouterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Router `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Router{}, &RouterList{})
}
