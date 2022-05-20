package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ElasticSearchSpec defines the desired state of ElasticSearch
type ElasticSearchSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ElasticSearch. Edit elasticsearch_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// ElasticSearchStatus defines the observed state of ElasticSearch
type ElasticSearchStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ElasticSearch is the Schema for the elasticsearches API
type ElasticSearch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticSearchSpec   `json:"spec,omitempty"`
	Status ElasticSearchStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ElasticSearchList contains a list of ElasticSearch
type ElasticSearchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ElasticSearch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ElasticSearch{}, &ElasticSearchList{})
}
