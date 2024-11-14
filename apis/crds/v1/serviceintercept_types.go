package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceInterceptSpec struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ServiceIntercept is the Schema for the serviceintercepts API
type ServiceIntercept struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceInterceptSpec `json:"spec,omitempty"`
	Status rApi.Status          `json:"status,omitempty" graphql:"noinput"`
}

func (svci *ServiceIntercept) EnsureGVK() {
	if svci != nil {
		svci.SetGroupVersionKind(GroupVersion.WithKind("App"))
	}
}

func (svci *ServiceIntercept) GetStatus() *rApi.Status {
	return &svci.Status
}

func (svci *ServiceIntercept) GetEnsuredLabels() map[string]string {
	m := map[string]string{
		"kloudlite.io/app.name": svci.Name,
	}

	return m
}

func (svci *ServiceIntercept) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("App").String(),
	}
}

//+kubebuilder:object:root=true

// ServiceInterceptList contains a list of ServiceIntercept
type ServiceInterceptList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceIntercept `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceIntercept{}, &ServiceInterceptList{})
}
