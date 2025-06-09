package v1

import (
	"fmt"

	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/toolkit/reconciler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterManagedServiceSpec defines the desired state of ClusterManagedService
type ClusterManagedServiceSpec struct {
	TargetNamespace string             `json:"targetNamespace" graphql:"noinput"`
	MSVCSpec        ManagedServiceSpec `json:"msvcSpec"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ClusterManagedService is the Schema for the clustermanagedservices API
type ClusterManagedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterManagedServiceSpec `json:"spec,omitempty"`
	Status rApi.Status               `json:"status,omitempty" graphql:"noinput"`
}

func (obj *ClusterManagedService) PatchWithDefaults() (hasPatched bool) {
	hasPatched = false

	if obj.Spec.TargetNamespace == "" {
		hasPatched = true
		obj.Spec.TargetNamespace = fmt.Sprintf("cmsvc-%s", obj.Name)
	}

	if obj.Spec.MSVCSpec.Plugin != nil && obj.Spec.MSVCSpec.Plugin.Export.ViaSecret == "" {
		hasPatched = true
		obj.Spec.MSVCSpec.Plugin.Export.ViaSecret = obj.Name + "-export"
	}

	return hasPatched
}

func (m *ClusterManagedService) EnsureGVK() {
	if m != nil {
		m.SetGroupVersionKind(GroupVersion.WithKind("ClusterManagedService"))
	}
}

func (m *ClusterManagedService) GetStatus() *rApi.Status {
	return &m.Status
}

func (m *ClusterManagedService) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.ClusterManagedServiceNameKey: m.Name,
	}
}

func (m *ClusterManagedService) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// ClusterManagedServiceList contains a list of ClusterManagedService
type ClusterManagedServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterManagedService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterManagedService{}, &ClusterManagedServiceList{})
}
