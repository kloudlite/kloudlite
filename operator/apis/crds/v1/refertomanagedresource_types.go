package v1

import (
	"fmt"

	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	ct "github.com/kloudlite/operator/apis/common-types"
)

// ReferToManagedResourceSpec defines the desired state of ReferToManagedResource
type ReferToManagedResourceSpec struct {
	ManagedResourceRef ct.NamespacedResourceRef `json:"managedResourceRef"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ReferToManagedResource is the Schema for the refertomanagedresources API
type ReferToManagedResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReferToManagedResourceSpec `json:"spec,omitempty"`
	Status rApi.Status                `json:"status,omitempty"`
}

func (rtmr *ReferToManagedResource) EnsureGVK() {
	if rtmr != nil {
		rtmr.SetGroupVersionKind(GroupVersion.WithKind("ReferToManagedResource"))
	}
}

func (p *ReferToManagedResource) GetStatus() *rApi.Status {
	return &p.Status
}

func (p *ReferToManagedResource) GetEnsuredLabels() map[string]string {
	return map[string]string{
		"kloudlite.io/mres.ref": fmt.Sprintf("%s.%s", p.GetNamespace(), p.GetName()),
	}
}

func (p *ReferToManagedResource) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// ReferToManagedResourceList contains a list of ReferToManagedResource
type ReferToManagedResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReferToManagedResource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ReferToManagedResource{}, &ReferToManagedResourceList{})
}
