package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IPBindSpec defines the desired state of IPBind
type IPBindSpec struct {
}

// IPBindStatus defines the observed state of IPBind
type IPBindStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// IPBind is the Schema for the ipbinds API
type IPBind struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IPBindSpec   `json:"spec,omitempty"`
	Status IPBindStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IPBindList contains a list of IPBind
type IPBindList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPBind `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IPBind{}, &IPBindList{})
}
