package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ManagedServiceSpec defines the desired state of ManagedService
type ManagedServiceSpec struct {
	Type   string            `json:"type"`
	Inputs map[string]string `json:"inputs"`
}

// ManagedServiceStatus defines the observed state of ManagedService
type ManagedServiceStatus struct {
	Generation int64              `json:"generation,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ManagedService is the Schema for the managedservices API
type ManagedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedServiceSpec   `json:"spec,omitempty"`
	Status ManagedServiceStatus `json:"status,omitempty"`
}

func (msvc *ManagedService) DefaultStatus() {
	msvc.Status.Generation = msvc.Generation
}

func (msvc *ManagedService) IsNewGeneration() bool {
	return msvc.Generation > msvc.Status.Generation
}

func (msvc *ManagedService) HasToBeDeleted() bool {
	return msvc.GetDeletionTimestamp() != nil
}

func (msvc *ManagedService) BuildConditions() {
}

//+kubebuilder:object:root=true

// ManagedServiceList contains a list of ManagedService
type ManagedServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedService `json:"items"`
}

//+kubebuilder:object:root=true
type K8sYaml struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              unstructured.Unstructured `json:"spec"`
}

func init() {
	SchemeBuilder.Register(&ManagedService{}, &ManagedServiceList{}, &K8sYaml{})
}
