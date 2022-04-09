package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Json string

// ManagedServiceSpec defines the desired state of ManagedService
type ManagedServiceSpec struct {
	// Contains LastApplied Value for update clauses
	LastApplied Json `json:"lastApplied,omitempty"`

	// Values define the managed services values that correspond to a particular templateId
	Values Json `json:"values"`

	// Version define the resource version currently in use
	Version uint16 `json:"version"`

	// managed svc template Id
	TemplateName string `json:"templateName"`
}

// ManagedServiceStatus defines the observed state of ManagedService
type ManagedServiceStatus struct {
	MarkedDeleted bool `json:"markedDeleted"`
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
