package v1

import (
	"fmt"

	ct "operators.kloudlite.io/apis/common-types"
	"operators.kloudlite.io/lib/constants"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	libOperator "operators.kloudlite.io/lib/operator"
)

// ServiceSpec defines the desired state of Service
type ServiceSpec struct {
	CloudProvider ct.CloudProvider `json:"cloudProvider"`
	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	ReplicaCount int `json:"replicaCount,omitempty"`
	// Inputs    rawJson.KubeRawJson `json:"inputs,omitempty"`
	Storage   ct.Storage   `json:"storage"`
	Resources ct.Resources `json:"resources"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Service is the Schema for the services API
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec        `json:"spec,omitempty"`
	Status libOperator.Status `json:"status,omitempty"`
}

func (s *Service) GetStatus() *libOperator.Status {
	return &s.Status
}

func (s *Service) GetEnsuredLabels() map[string]string {
	return map[string]string{
		fmt.Sprintf("%s/ref", GroupVersion.Group): s.Name,
	}
}

func (m *Service) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("Service").String(),
	}
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
