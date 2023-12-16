package v1

import (
	ct "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DatabaseOutput struct {
	Credentials ct.SecretRef `json:"credentials"`
}

// DatabaseSpec defines the desired state of Database
type DatabaseSpec struct {
	MsvcRef      ct.MsvcRef     `json:"msvcRef"`
	ResourceName string         `json:"resourceName"`
	Output       DatabaseOutput `json:"output,omitempty" graphql:"noinput"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/msvc-gvk",name=Msvc GVK,type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Database is the Schema for the databases API
type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseSpec `json:"spec"`
	Status rApi.Status  `json:"status,omitempty"`
}

func (db *Database) EnsureGVK() {
	if db != nil {
		db.SetGroupVersionKind(GroupVersion.WithKind("Database"))
	}
}

func (d *Database) GetStatus() *rApi.Status {
	return &d.Status
}

func (d *Database) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.MsvcNameKey: d.Spec.MsvcRef.Name,
	}
}

func (d *Database) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("Database").String(),
		"kloudlite.io/msvc-gvk":                   d.Spec.MsvcRef.GroupVersionKind().String(),
	}
}

// +kubebuilder:object:root=true

// DatabaseList contains a list of Database
type DatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Database `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Database{}, &DatabaseList{})
}
