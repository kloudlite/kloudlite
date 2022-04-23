package v1

import (
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ManagedServiceSpec defines the desired state of ManagedService
type ManagedServiceSpec struct {
	// Input fields
	Inputs map[string]string `json:"inputs"`
	// Output Fields
	Outputs []Output `json:"outputs"`
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
	meta.SetStatusCondition(&msvc.Status.Conditions, metav1.Condition{
		Type:               "ApplyJobCheck",
		Status:             msvc.Status.ApplyJobCheck.ConditionStatus(),
		ObservedGeneration: msvc.Generation,
		Reason:             msvc.Status.ApplyJobCheck.Reason(),
		Message:            msvc.Status.ApplyJobCheck.Message,
	})

	meta.SetStatusCondition(&msvc.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             msvc.Status.ApplyJobCheck.ConditionStatus(),
		ObservedGeneration: msvc.Generation,
		Reason:             msvc.Status.ApplyJobCheck.Reason(),
		Message:            msvc.Status.ApplyJobCheck.Message,
	})
}

//+kubebuilder:object:root=true

// ManagedServiceList contains a list of ManagedService
type ManagedServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedService{}, &ManagedServiceList{})
}
