package v1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/lib/constants"
	rApi "operators.kloudlite.io/lib/operator"
)

type Route struct {
	Path   string `json:"path"`
	App    string `json:"app,omitempty"`
	Lambda string `json:"lambda,omitempty"`
	Port   uint16 `json:"port"`
}

type RateLimit struct {
	Enabled     bool `json:"enabled,omitempty"`
	Rps         int  `json:"rps,omitempty"`
	Rpm         int  `json:"rpm,omitempty"`
	Connections int  `json:"connections,omitempty"`
}

// RouterSpec defines the desired state of Router
type RouterSpec struct {
	ForceSSLRedirect bool               `json:"forceSSLRedirect,omitempty"`
	RateLimit        RateLimit          `json:"rateLimit,omitempty"`
	MaxBodySize      int                `json:"maxBodySize,omitempty"`
	Domains          []string           `json:"domains"`
	Routes           map[string][]Route `json:"routes"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Router is the Schema for the routers API
type Router struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouterSpec  `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

var RouterGroupVersionKind = GroupVersion.WithKind("Router")

func (r *Router) GetStatus() *rApi.Status {
	return &r.Status
}

func (r *Router) GetEnsuredLabels() map[string]string {
	return map[string]string{
		fmt.Sprintf("%s/ref", GroupVersion.Group): r.Name,
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
