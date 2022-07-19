package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type HarborProjectSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of HarborProject. Edit harborproject_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// HarborProjectStatus defines the observed state of HarborProject
type HarborProjectStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// HarborProject is the Schema for the harborprojects API
type HarborProject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborProjectSpec   `json:"spec,omitempty"`
	Status HarborProjectStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HarborProjectList contains a list of HarborProject
type HarborProjectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HarborProject `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HarborProject{}, &HarborProjectList{})
}
