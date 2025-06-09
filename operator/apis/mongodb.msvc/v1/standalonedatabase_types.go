package v1

import (
	ct "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StandaloneDatabaseSpec defines the desired state of StandaloneDatabase
type StandaloneDatabaseSpec struct {
	MsvcRef ct.MsvcRef `json:"msvcRef"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// StandaloneDatabase is the Schema for the standalonedatabases API
type StandaloneDatabase struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StandaloneDatabaseSpec `json:"spec,omitempty"`
	Status rApi.Status            `json:"status,omitempty"`

	Output ct.ManagedResourceOutput `json:"output,omitempty"`
}

func (db *StandaloneDatabase) EnsureGVK() {
	if db != nil {
		db.SetGroupVersionKind(GroupVersion.WithKind("StandaloneDatabase"))
	}
}

func (db *StandaloneDatabase) GetStatus() *rApi.Status {
	return &db.Status
}

func (db *StandaloneDatabase) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.MsvcNameKey:      db.Spec.MsvcRef.Name,
		constants.MsvcNamespaceKey: db.Spec.MsvcRef.Namespace,
	}
}

func (p *StandaloneDatabase) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// StandaloneDatabaseList contains a list of StandaloneDatabase
type StandaloneDatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StandaloneDatabase `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StandaloneDatabase{}, &StandaloneDatabaseList{})
}
