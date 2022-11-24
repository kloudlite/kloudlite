package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ct "operators.kloudlite.io/apis/common-types"
	"operators.kloudlite.io/pkg/constants"
	rApi "operators.kloudlite.io/pkg/operator"
)

type Redpanda struct {
	AdminSecretRef     ct.SecretRef `json:"adminSecretRef"`
	Topics             []string     `json:"topics,omitempty"`
	PushAccessToTopics []string     `json:"pushAccessToTopics,omitempty"`
	ClusterDomain      []string     `json:"clusterDomain,omitempty"`
}

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	Redpanda   Redpanda      `json:"redpanda,omitempty"`
	KubeConfig *ct.SecretRef `json:"kubeConfig,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (c *Cluster) GetStatus() *rApi.Status {
	return &c.Status
}

func (c *Cluster) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (c *Cluster) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("Cluster").String(),
	}
}

// +kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
