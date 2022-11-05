package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ct "operators.kloudlite.io/apis/common-types"
	"operators.kloudlite.io/lib/constants"
	rApi "operators.kloudlite.io/lib/operator"
)

// ClusterServiceSpec defines the desired state of ClusterService
type ClusterServiceSpec struct {
	Region       string              `json:"region"`
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	ReplicaCount int          `json:"replicaCount,omitempty"`
	Resources    ct.Resources `json:"resources"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ClusterService is the Schema for the clusterservices API
type ClusterService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterServiceSpec `json:"spec,omitempty"`
	Status rApi.Status        `json:"status,omitempty"`
}

func (c *ClusterService) GetStatus() *rApi.Status {
	return &c.Status
}

func (c *ClusterService) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.MsvcNameKey: c.Name,
	}
}

func (c *ClusterService) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

// +kubebuilder:object:root=true

// ClusterServiceList contains a list of ClusterService
type ClusterServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterService{}, &ClusterServiceList{})
}
