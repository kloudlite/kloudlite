package v1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rApi "operators.kloudlite.io/lib/operator"
)

type Route struct {
	Path   string `json:"path"`
	App    string `json:"app,omitempty"`
	Lambda string `json:"lambda,omitempty"`
	Port   uint16 `json:"port"`
}

// RouterSpec defines the desired state of Router
type RouterSpec struct {
	ForceSSLRedirect bool             `json:"forceSSLRedirect,omitempty"`
	Domains          []string         `json:"domains"`
	Routes           map[string]Route `json:"routes"`
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

func (r *Router) NameRef() string {
	return ""
	// return fmt.Sprintf("%s")
}

func (r *Router) GetStatus() *rApi.Status {
	return &r.Status
}

func (r *Router) GetEnsuredLabels() map[string]string {
	return map[string]string{
		fmt.Sprintf("%s/ref", GroupVersion.Group): r.Name,
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
