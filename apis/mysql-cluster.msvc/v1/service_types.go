package v1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rawJson "operators.kloudlite.io/lib/raw-json"

	fn "operators.kloudlite.io/lib/functions"
)

// ServiceSpec defines the desired state of Service
type ServiceSpec struct {
	Inputs rawJson.KubeRawJson `json:"inputs,omitempty"`
}

// ServiceStatus defines the observed state of Service
type ServiceStatus struct {
	LastHash      string              `json:"lastHash,omitempty"`
	GeneratedVars rawJson.KubeRawJson `json:"generatedVars,omitempty"`
	Conditions    []metav1.Condition  `json:"conditions,omitempty"`
	OpsConditions []metav1.Condition  `json:"opsConditions,omitempty"`
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

func (s *Service) Hash() string {
	m := make(map[string]interface{}, 3)
	m["name"] = s.Name
	m["namespace"] = s.Namespace
	m["spec"] = s.Spec
	hash, _ := fn.Json.Hash(m)
	return hash
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
