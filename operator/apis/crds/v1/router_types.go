package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/operator/toolkit/reconciler"
)

type Route struct {
	Host    string `json:"host"`
	Service string `json:"service"`
	Path    string `json:"path"`
	Port    uint16 `json:"port"`

	// +kubebuilder:default=false
	Rewrite bool `json:"rewrite,omitempty"`
}

type RateLimit struct {
	Enabled     bool `json:"enabled,omitempty"`
	Rps         int  `json:"rps,omitempty"`
	Rpm         int  `json:"rpm,omitempty"`
	Connections int  `json:"connections,omitempty"`
}

type Https struct {
	// +kubebuilder:default=true
	Enabled       bool   `json:"enabled"`
	ForceRedirect bool   `json:"forceRedirect,omitempty"`
	ClusterIssuer string `json:"clusterIssuer,omitempty"`
}

type BasicAuth struct {
	Enabled    bool   `json:"enabled"`
	Username   string `json:"username,omitempty"`
	SecretName string `json:"secretName,omitempty"`
}

type Cors struct {
	// +kubebuilder:default=false
	Enabled          bool     `json:"enabled,omitempty"`
	Origins          []string `json:"origins,omitempty"`
	AllowCredentials bool     `json:"allowCredentials,omitempty"`
}

// RouterSpec defines the desired state of Router
type RouterSpec struct {
	IngressClass    string  `json:"ingressClass,omitempty"`
	BackendProtocol *string `json:"backendProtocol,omitempty"`
	Https           *Https  `json:"https,omitempty"`

	RateLimit       *RateLimit `json:"rateLimit,omitempty"`
	MaxBodySizeInMB *int       `json:"maxBodySizeInMB,omitempty"`

	BasicAuth *BasicAuth `json:"basicAuth,omitempty"`
	Cors      *Cors      `json:"cors,omitempty"`

	// NginxIngressAnnotations is additional list of annotations on ingress resource
	// INFO: must be used when router does not have direct support for it
	NginxIngressAnnotations map[string]string `json:"nginxIngressAnnotations,omitempty"`

	Routes []Route `json:"routes,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/router\\.ingress-class",name=Class,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/router\\.domains",name=domains,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Router is the Schema for the routers API
type Router struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RouterSpec `json:"spec"`
	// +kubebuilder:default=true
	Enabled bool              `json:"enabled,omitempty"`
	Status  reconciler.Status `json:"status,omitempty" graphql:"noinput"`
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

// +kubebuilder:object:root=true

// RouterList contains a list of Router
type RouterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Router `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Router{}, &RouterList{})
}
