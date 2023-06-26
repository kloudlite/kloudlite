package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

type Route struct {
	App    string `json:"app,omitempty"`
	Lambda string `json:"lambda,omitempty"`
	Path   string `json:"path"`
	Port   uint16 `json:"port"`
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
	ClusterIssuer string `json:"clusterIssuer,omitempty"`
	ForceRedirect bool   `json:"forceRedirect,omitempty"`
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
	Region          string  `json:"region,omitempty"`
	IngressClass    string  `json:"ingressClass,omitempty"`
	BackendProtocol *string `json:"backendProtocol,omitempty"`
	Https           *Https  `json:"https,omitempty"`
	// +kubebuilder:validation:Optional

	RateLimit       *RateLimit `json:"rateLimit,omitempty"`
	MaxBodySizeInMB *int       `json:"maxBodySizeInMB,omitempty"`
	Domains         []string   `json:"domains"`
	Routes          []Route    `json:"routes,omitempty"`
	BasicAuth       *BasicAuth `json:"basicAuth,omitempty"`
	Cors            *Cors      `json:"cors,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Router is the Schema for the routers API
type Router struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RouterSpec `json:"spec,omitempty"`
	// +kubebuilder:default=true
	Enabled bool        `json:"enabled,omitempty"`
	Status  rApi.Status `json:"status,omitempty"`
}

func (r *Router) EnsureGVK() {
	if r != nil {
		r.SetGroupVersionKind(GroupVersion.WithKind("Router"))
	}
}

func (r *Router) GetStatus() *rApi.Status {
	return &r.Status
}

func (r *Router) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.RouterNameKey: r.Name,
	}
}

func (m *Router) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("Router").String(),
	}
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
