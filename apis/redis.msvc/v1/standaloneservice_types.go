package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ct "operators.kloudlite.io/apis/common-types"
	"operators.kloudlite.io/pkg/constants"
	rApi "operators.kloudlite.io/pkg/operator"
)

// StandaloneServiceSpec defines the desired state of StandaloneService
type StandaloneServiceSpec struct {
	Region       string              `json:"region"`
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:default=1
	ReplicaCount int          `json:"replicaCount,omitempty"`
	Resources    ct.Resources `json:"resources"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Cloud",type="string",JSONPath=".spec.cloudProvider.cloud"
// +kubebuilder:printcolumn:name="ReplicaCount",type="integer",JSONPath=".spec.replicaCount"
// +kubebuilder:printcolumn:name="Status",type="boolean",JSONPath=".status.isReady"

// StandaloneService is the Schema for the standaloneservices API
type StandaloneService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StandaloneServiceSpec `json:"spec,omitempty"`
	Status rApi.Status           `json:"status,omitempty"`
}

func (ss *StandaloneService) GetStatus() *rApi.Status {
	return &ss.Status
}

func (ss *StandaloneService) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.MsvcNameKey: ss.Name,
	}
}

func (ss *StandaloneService) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("StandaloneService").String(),
	}
}

// +kubebuilder:object:root=true

// StandaloneServiceList contains a list of StandaloneService
type StandaloneServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StandaloneService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StandaloneService{}, &StandaloneServiceList{})
}
