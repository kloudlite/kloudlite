package v1

import (
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AccountS3BucketSpec defines the desired state of AccountS3Bucket
type AccountS3BucketSpec struct {
	AccountName  string `json:"accountName"`
	BucketRegion string `json:"bucketRegion"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// AccountS3Bucket is the Schema for the accounts3buckets API
type AccountS3Bucket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccountS3BucketSpec `json:"spec,omitempty"`
	Status rApi.Status         `json:"status,omitempty" graphql:"noinput"`
}

func (b *AccountS3Bucket) GetStatus() *rApi.Status {
	return &b.Status
}

func (b *AccountS3Bucket) EnsureGVK() {
	if b != nil {
		b.SetGroupVersionKind(GroupVersion.WithKind("AccountS3Bucket"))
	}
}

func (b *AccountS3Bucket) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (b *AccountS3Bucket) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// AccountS3BucketList contains a list of AccountS3Bucket
type AccountS3BucketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccountS3Bucket `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AccountS3Bucket{}, &AccountS3BucketList{})
}
