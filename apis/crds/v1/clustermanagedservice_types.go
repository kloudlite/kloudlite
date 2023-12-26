package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterManagedServiceSpec defines the desired state of ClusterManagedService
type ClusterManagedServiceSpec struct {
	Namespace string             `json:"namespace"`
	MSVCSepec ManagedServiceSpec `json:"msvcSpec"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/service-gvk",name=Service GVK,type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ClusterManagedService is the Schema for the clustermanagedservices API
type ClusterManagedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterManagedServiceSpec `json:"spec,omitempty"`
	Status rApi.Status               `json:"status,omitempty" graphql:"noinput"`
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
		"kloudlite.io/cmsvc.name": m.Name,
	}
}

func (m *ClusterManagedService) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("ClusterManagedService").String(),
	}
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
