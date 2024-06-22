package v1

import (
	"fmt"

	ct "github.com/kloudlite/operator/apis/common-types"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceBindingSpec defines the desired state of ServiceBinding
type ServiceBindingSpec struct {
	GlobalIP   string                    `json:"globalIP"`
	ServiceIP  *string                   `json:"serviceIP,omitempty"`
	ServiceRef *ct.NamespacedResourceRef `json:"serviceRef,omitempty"`
	Ports      []corev1.ServicePort      `json:"ports,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
//+kubebuilder:printcolumn:JSONPath=".spec.globalIP",name=GlobalIP,type=string
//+kubebuilder:printcolumn:JSONPath=".spec.serviceIP",name=ServiceIP,type=string
//+kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/global\\.hostname",name=Host,type=string
//+kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
//+kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ServiceBinding is the Schema for the servicebindings API
type ServiceBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceBindingSpec `json:"spec,omitempty"`
	Status rApi.Status        `json:"status,omitempty"`
}

func (sb *ServiceBinding) EnsureGVK() {
	if sb != nil {
		sb.SetGroupVersionKind(GroupVersion.WithKind("ServiceBinding"))
	}
}

func (sb *ServiceBinding) GetStatus() *rApi.Status {
	return &sb.Status
}

func (sb *ServiceBinding) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (sb *ServiceBinding) GetEnsuredAnnotations() map[string]string {
	key := "kloudlite.io/servicebinding.reservation"
	v, ok := sb.GetLabels()[key]
	if !ok || v == "false" {
		return map[string]string{key: "UnReserved"}
	}

	if sb.Spec.ServiceRef == nil {
		return map[string]string{key: "Reserved"}
	}
	return map[string]string{key: fmt.Sprintf("Reserved (%s/%s)", sb.Spec.ServiceRef.Namespace, sb.Spec.ServiceRef.Name)}
}

//+kubebuilder:object:root=true

// ServiceBindingList contains a list of ServiceBinding
type ServiceBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServiceBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ServiceBinding{}, &ServiceBindingList{})
}
