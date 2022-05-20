package v1

import (
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	t "operators.kloudlite.io/lib/types"
)

// ServiceSpec defines the desired state of Service
type ServiceSpec struct {
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	Inputs json.RawMessage `json:"inputs,omitempty"`
}

// ServiceStatus defines the observed state of Service
type ServiceStatus struct {
	Conditions t.Conditions `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Service is the Schema for the services API
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec   `json:"spec,omitempty"`
	Status ServiceStatus `json:"status,omitempty"`
}

func (s *Service) NameRef() string {
	return fmt.Sprintf("%s/%s/%s", s.GroupVersionKind().Group, s.Namespace, s.Name)
}

func (s Service) LabelRef() (key, value string) {
	return "msvc.kloudlite.io/service", GroupVersion.Group
}

func (s *Service) HasLabels() bool {
	key, value := s.LabelRef()
	if s.Labels[key] != value {
		return false
	}
	return true
}

func (s *Service) EnsureLabels() {
	key, value := s.LabelRef()
	s.SetLabels(map[string]string{key: value})
}

// +kubebuilder:object:root=true

// ServiceList contains a list of Service
type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Service `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Service{}, &ServiceList{})
}
