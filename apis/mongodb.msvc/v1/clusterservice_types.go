package v1

import (
	ct "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterServiceOutput struct {
	Credentials ct.SecretRef `json:"credentials,omitempty"`
	HelmSecret  ct.SecretRef `json:"helmSecret,omitempty"`
}

// ClusterServiceSpec defines the desired state of ClusterService
type ClusterServiceSpec struct {
	NodeSelector              map[string]string                 `json:"nodeSelector,omitempty"`
	Tolerations               []corev1.Toleration               `json:"tolerations,omitempty"`
	TopologySpreadConstraints []corev1.TopologySpreadConstraint `json:"topologySpreadConstraints,omitempty"`

	Replicas  int                  `json:"replicas"`
	Resources ct.Resources         `json:"resources"`
	Output    ClusterServiceOutput `json:"output,omitempty" graphql:"noinput"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ClusterService is the Schema for the clusterservices API
type ClusterService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterServiceSpec `json:"spec"`
	Status rApi.Status        `json:"status,omitempty"`
}

func (cs *ClusterService) EnsureGVK() {
	if cs != nil {
		cs.SetGroupVersionKind(GroupVersion.WithKind("ClusterService"))
	}
}

func (c *ClusterService) GetStatus() *rApi.Status {
	return &c.Status
}

func (c *ClusterService) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.MsvcNameKey: c.Name,
	}
}

func (c *ClusterService) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("ClusterService").String(),
	}
}

// +kubebuilder:object:root=true

// ClusterServiceList contains a list of ClusterService
type ClusterServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterService{}, &ClusterServiceList{})
}
