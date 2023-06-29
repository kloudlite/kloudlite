package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DriverSpec defines the desired state of Driver
type DriverSpec struct {
	// +kubebuilder:validation:Enum=do;aws;gcp;azure;k3s-local
	Provider     string              `json:"provider"`
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
	SecretRef    string              `json:"secretRef"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// Driver is the Schema for the drivers API
type Driver struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DriverSpec  `json:"spec"`
	Status rApi.Status `json:"status,omitempty"`
}

func (d *Driver) GetStatus() *rApi.Status {
	return &d.Status
}

func (d *Driver) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.CsiDriverNameKey: d.Name,
	}
}

func (d *Driver) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("Driver").String(),
	}
}

// +kubebuilder:object:root=true

// DriverList contains a list of Driver
type DriverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Driver `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Driver{}, &DriverList{})
}
