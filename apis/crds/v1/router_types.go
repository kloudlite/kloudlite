package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/pkg/constants"
	rApi "operators.kloudlite.io/pkg/operator"
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
	Enabled       bool `json:"enabled"`
	ForceRedirect bool `json:"forceRedirect,omitempty"`
}

type BasicAuth struct {
	Enabled    bool   `json:"enabled"`
	SecretName string `json:"secretName"`
}

// RouterSpec defines the desired state of Router
type RouterSpec struct {
	Region string `json:"region,omitempty"`
	Https  Https  `json:"https,omitempty"`
	// +kubebuilder:validation:Optional
	RateLimit       RateLimit `json:"rateLimit,omitempty"`
	MaxBodySizeInMB int       `json:"maxBodySizeInMB,omitempty"`
	Domains         []string  `json:"domains"`
	Routes          []Route   `json:"routes,omitempty"`
	BasicAuth       BasicAuth `json:"basicAuth,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Router is the Schema for the routers API
type Router struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouterSpec  `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (r *Router) GetStatus() *rApi.Status {
	return &r.Status
}

func (r *Router) GetEnsuredLabels() map[string]string {
	return map[string]string{}
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
