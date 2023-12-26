package v1

import (
	ct "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/pkg/influx"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	Spec   BucketSpec  `json:"spec"`
	Status rApi.Status `json:"status,omitempty"`
}

func (b *Bucket) EnsureGVK() {
	if b != nil {
		b.SetGroupVersionKind(GroupVersion.WithKind("Bucket"))
	}
}

func (b *Bucket) GetStatus() *rApi.Status {
	return &b.Status
}

func (b *Bucket) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.MsvcNameKey:      b.Spec.MsvcRef.Name,
		constants.MsvcNamespaceKey: b.Spec.MsvcRef.Namespace,
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
