package v1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MongoDBSpec defines the desired state of Project
// +kubebuilder:pruning:PreserveUnknownFields
type MongoDBSpec struct {
	// unstructured.Unstructured `json:",inline"`
}

// +kubebuilder:pruning:PreserveUnknownFields
type MongoDBStatus struct {
	// unstructured.Unstructured `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MongoDB is the Schema for the mongodbs API
type MongoDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongoDBSpec   `json:"spec,omitempty"`
	Status MongoDBStatus `json:"status,omitempty"`
}

func (m *MongoDB) DeploymentName() string {
	return fmt.Sprintf("%s-mongodb", m.Name)
}

//+kubebuilder:object:root=true

// MongoDBList contains a list of MongoDB
type MongoDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongoDB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongoDB{}, &MongoDBList{})
}
