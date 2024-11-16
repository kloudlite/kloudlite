package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ManagedServicePluginSpec defines the desired state of ManagedServicePlugin
type ManagedServicePluginSpec struct {
	GVKs []metav1.GroupVersionKind `json:"gvks,omitempty"`
}

// ManagedServicePluginStatus defines the observed state of ManagedServicePlugin
type ManagedServicePluginStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ManagedServicePlugin is the Schema for the managedserviceplugins API
type ManagedServicePlugin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedServicePluginSpec   `json:"spec,omitempty"`
	Status ManagedServicePluginStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ManagedServicePluginList contains a list of ManagedServicePlugin
type ManagedServicePluginList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedServicePlugin `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedServicePlugin{}, &ManagedServicePluginList{})
}
