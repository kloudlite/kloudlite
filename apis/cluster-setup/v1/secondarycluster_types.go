package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SecondaryClusterSpec defines the desired state of SecondaryCluster
type SecondaryClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of SecondaryCluster. Edit secondarycluster_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// SecondaryClusterStatus defines the observed state of SecondaryCluster
type SecondaryClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SecondaryCluster is the Schema for the secondaryclusters API
type SecondaryCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecondaryClusterSpec   `json:"spec,omitempty"`
	Status SecondaryClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SecondaryClusterList contains a list of SecondaryCluster
type SecondaryClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecondaryCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SecondaryCluster{}, &SecondaryClusterList{})
}
