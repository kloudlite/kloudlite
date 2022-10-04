package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ct "operators.kloudlite.io/apis/common-types"
	rApi "operators.kloudlite.io/lib/operator"
)

type TopicSpec struct {
	AdminSecretRef ct.SecretRef `json:"adminSecretRef"`
	// +kubebuilder:default=5
	PartitionCount int `json:"partitionCount"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Topic is the Schema for the topics API
type Topic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopicSpec   `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (t *Topic) GetStatus() *rApi.Status {
	return &t.Status
}

func (t *Topic) GetEnsuredLabels() map[string]string {
	return map[string]string{
		"kloudlite.io/msvc.name": "redpanda",
	}
}

func (t *Topic) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

// +kubebuilder:object:root=true

// TopicList contains a list of Topic
type TopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Topic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Topic{}, &TopicList{})
}
