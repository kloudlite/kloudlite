package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KeyPrefixSpec defines the desired state of KeyPrefix
type KeyPrefixSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of KeyPrefix. Edit keyprefix_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// KeyPrefixStatus defines the observed state of KeyPrefix
type KeyPrefixStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KeyPrefix is the Schema for the keyprefixes API
type KeyPrefix struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KeyPrefixSpec   `json:"spec,omitempty"`
	Status KeyPrefixStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KeyPrefixList contains a list of KeyPrefix
type KeyPrefixList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KeyPrefix `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KeyPrefix{}, &KeyPrefixList{})
}
