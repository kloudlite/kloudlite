package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ct "operators.kloudlite.io/apis/common-types"
	"operators.kloudlite.io/lib/constants"
	rApi "operators.kloudlite.io/lib/operator"
)

type adminTT struct {
	// +kubebuilder:default=admin
	// +kubebuidler:validation:Optional
	Username string `json:"username"`
	// +kubebuidler:validation:Optional
	// +kubebuilder:default=primary
	Bucket string `json:"bucket"`
	// +kubebuilder:default=primary
	// +kubebuidler:validation:Optional
	Org string `json:"org"`
}

// ServiceSpec defines the desired state of Service
type ServiceSpec struct {
	CloudProvider ct.CloudProvider `json:"cloudProvider"`
	// +kubebuilder:validation:Optional
	NodeSelector map[string]string `json:"nodeSelector"`
	Admin        adminTT           `json:"admin"`
	Storage      ct.Storage        `json:"storage"`
	Resources    ct.Resources      `json:"resources"`
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
	return map[string]string{}
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
