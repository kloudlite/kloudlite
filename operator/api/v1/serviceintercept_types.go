package v1

import (
	"github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceInterceptPortMappings struct {
	// +kubebuilder:default="TCP"
	Protocol    corev1.Protocol `json:"protocol,omitempty"`
	ServicePort int32           `json:"servicePort"`
	DevicePort  int32           `json:"devicePort"`
}

// ServiceInterceptSpec defines the desired state of ServiceIntercept.
type ServiceInterceptSpec struct {
	ToHost       string                         `json:"toHost"`
	PortMappings []ServiceInterceptPortMappings `json:"portMappings"`
	ServiceRef   corev1.LocalObjectReference    `json:"serviceRef"`
}

type ServiceInterceptStatus struct {
	reconciler.Status `json:",inline"`
	Selector          map[string]string `json:"selector,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.operator\\.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.operator\\.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ServiceIntercept is the Schema for the serviceintercepts API.
type ServiceIntercept struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceInterceptSpec   `json:"spec,omitempty"`
	Status ServiceInterceptStatus `json:"status,omitempty"`
}

func (svci *ServiceIntercept) EnsureGVK() {
	if svci != nil {
		svci.SetGroupVersionKind(GroupVersion.WithKind("ServiceIntercept"))
	}
}

func (svci *ServiceIntercept) GetStatus() *reconciler.Status {
	return &svci.Status.Status
}

func (svci *ServiceIntercept) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (svci *ServiceIntercept) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

// +kubebuilder:object:root=true

// ServiceInterceptList contains a list of ServiceIntercept.
type ServiceInterceptList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceIntercept `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceIntercept{}, &ServiceInterceptList{})
}
