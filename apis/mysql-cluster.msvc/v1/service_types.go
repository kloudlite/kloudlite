package v1

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rApi "operators.kloudlite.io/lib/operator"
	rawJson "operators.kloudlite.io/lib/raw-json"
)

// ServiceSpec defines the desired state of Service
type ServiceSpec struct {
	Inputs rawJson.RawJson `json:"inputs,omitempty"`
}

// ServiceStatus defines the observed state of Service

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Service is the Schema for the services API
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (s *Service) NameRef() string {
	return fmt.Sprintf("%s/%s/%s", s.GroupVersionKind().Group, s.Namespace, s.Name)
}

func (s *Service) GetStatus() *rApi.Status {
	return &s.Status
}

func (s *Service) GetEnsuredLabels() map[string]string {
	return map[string]string{
		fmt.Sprintf("%s/ref", GroupVersion.Group): s.Name,
	}
}

func (s *Service) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
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
