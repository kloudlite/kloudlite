package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ct "operators.kloudlite.io/apis/common-types"
	"operators.kloudlite.io/pkg/constants"
	rApi "operators.kloudlite.io/pkg/operator"
)

type Admin struct {
	Username string `json:"username"`
	Bucket   string `json:"bucket"`
	Org      string `json:"org"`
}

// ServiceSpec defines the desired state of Service
type ServiceSpec struct {
	CloudProvider ct.CloudProvider `json:"cloudProvider"`
	Region        string           `json:"region"`
	// +kubebuilder:validation:Optional
	NodeSelector map[string]string   `json:"nodeSelector"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
	Admin        *Admin              `json:"admin,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	ReplicaCount int          `json:"replicaCount,omitempty"`
	Resources    ct.Resources `json:"resources"`
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
	return map[string]string{
		constants.MsvcNameKey: s.Name,
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
