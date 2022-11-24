package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ct "operators.kloudlite.io/apis/common-types"
	"operators.kloudlite.io/pkg/constants"
	"operators.kloudlite.io/pkg/influx"
	rApi "operators.kloudlite.io/pkg/operator"
)

type BucketSpec struct {
	MsvcRef ct.MsvcRef     `json:"msvcRef"`
	Bucket  *influx.Bucket `json:"bucketRef"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Bucket is the Schema for the buckets API
type Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BucketSpec  `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (b *Bucket) GetStatus() *rApi.Status {
	return &b.Status
}

func (b *Bucket) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.MsvcNameKey: b.Spec.MsvcRef.Name,
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
