package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ct "operators.kloudlite.io/apis/common-types"
	rApi "operators.kloudlite.io/lib/operator"
)

type ServiceSpec struct {
	CloudProvider ct.CloudProvider `json:"cloudProvider"`

	// +kubebuilder:validation:optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// +kubebuidler:default=1
	// +kubebuilder:validation:Optional
	ReplicaCount int `json:"replicaCount"`

	Storage   ct.Storage   `json:"storage"`
	Resources ct.Resources `json:"resources"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Service is the Schema for the services API
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (s *Service) GetStatus() *rApi.Status {
	return &s.Status
}

func (s *Service) GetEnsuredLabels() map[string]string {
	return map[string]string{"kloudlite.io/msvc.name": s.Name}
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
