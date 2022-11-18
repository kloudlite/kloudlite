package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ct "operators.kloudlite.io/apis/common-types"
	"operators.kloudlite.io/pkg/constants"
	rApi "operators.kloudlite.io/pkg/operator"
)

type ServiceSpec struct {
	// CloudProvider ct.CloudProvider `json:"cloudProvider"`
	Region string `json:"region"`

	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`

	// +kubebuidler:default=1
	// +kubebuilder:validation:Optional
	ReplicaCount int `json:"replicaCount"`

	// Storage   ct.Storage   `json:"storage"`
	Resources ct.Resources `json:"resources"`

	// +kubebuilder:default=true
	KibanaEnabled bool `json:"kibanaEnabled,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Region",type="string",JSONPath=".spec.region"
// +kubebuilder:printcolumn:name="ReplicaCount",type="integer",JSONPath=".spec.replicaCount"
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

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
	return map[string]string{constants.MsvcNameKey: s.Name}
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
