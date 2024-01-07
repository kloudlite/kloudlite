package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ProjectManagedServiceSpec defines the desired state of ProjectManagedService
type ProjectManagedServiceSpec struct {
	TargetNamespace string             `json:"targetNamespace"`
	MSVCSpec        ManagedServiceSpec `json:"msvcSpec"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/service-gvk",name=Service GVK,type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ProjectManagedService is the Schema for the projectmanagedservices API
type ProjectManagedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectManagedServiceSpec `json:"spec,omitempty"`
	Status rApi.Status               `json:"status,omitempty" graphql:"noinput"`
}

func (m *ProjectManagedService) EnsureGVK() {
	if m != nil {
		m.SetGroupVersionKind(GroupVersion.WithKind("ProjectManagedService"))
	}
}

func (m *ProjectManagedService) GetStatus() *rApi.Status {
	return &m.Status
}

func (m *ProjectManagedService) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.ProjectManagedServiceNameKey: m.Name,
	}
}

func (m *ProjectManagedService) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// ProjectManagedServiceList contains a list of ProjectManagedService
type ProjectManagedServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectManagedService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProjectManagedService{}, &ProjectManagedServiceList{})
}
