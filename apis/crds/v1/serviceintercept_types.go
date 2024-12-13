package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SvcInterceptPortMappings struct {
	ServicePort uint16 `json:"servicePort"`
	DevicePort  uint16 `json:"devicePort"`
}

type ServiceInterceptSpec struct {
	ToAddr       string                     `json:"toAddress"`
	PortMappings []SvcInterceptPortMappings `json:"portMappings"`
}

type ServiceInterceptStatus struct {
	rApi.Status `json:",inline"`
	Selector    map[string]string `json:"selector,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date
// ServiceIntercept is the Schema for the serviceintercepts API
type ServiceIntercept struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceInterceptSpec   `json:"spec,omitempty"`
	Status ServiceInterceptStatus `json:"status,omitempty" graphql:"noinput"`
}

func (svci *ServiceIntercept) EnsureGVK() {
	if svci != nil {
		svci.SetGroupVersionKind(GroupVersion.WithKind("ServiceIntercept"))
	}
}

func (svci *ServiceIntercept) GetStatus() *rApi.Status {
	return &svci.Status.Status
}

func (svci *ServiceIntercept) GetEnsuredLabels() map[string]string {
	m := map[string]string{
		"kloudlite.io/svci.name": svci.Name,
	}

	return m
}

func (svci *ServiceIntercept) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("ServiceIntercept").String(),
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
