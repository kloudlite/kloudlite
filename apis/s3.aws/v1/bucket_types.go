package v1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/pkg/constants"
	rApi "operators.kloudlite.io/pkg/operator"
)

type BucketSpec struct {
	Region        string   `json:"region"`
	PublicFolders []string `json:"publicFolders"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Bucket is the Schema for the buckets API
type Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketSpec  `json:"spec"`
	Status rApi.Status `json:"status,omitempty"`
}

func (b *Bucket) GetStatus() *rApi.Status {
	return &b.Status
}

func (b *Bucket) GetEnsuredLabels() map[string]string {
	return map[string]string{
		fmt.Sprintf("%s/ref", GroupVersion.Group): b.Name,
	}
}

func (m *Bucket) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("Bucket").String(),
	}
}

// +kubebuilder:object:root=true

// BucketList contains a list of Bucket
type BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Bucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Bucket{}, &BucketList{})
}
