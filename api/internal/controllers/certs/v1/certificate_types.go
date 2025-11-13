package v1

import (
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CertificateSpec struct {
	CA string `json:"ca"`
}

type CertificateStatus struct {
	reconciler.Status `json:",inline"`
	TLSSecretName     string `json:"tlsSecretName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories={kloudlite}
// +kubebuilder:printcolumn:name="Seen",type=date,JSONPath=".status.lastReconcileTime"
// +kubebuilder:printcolumn:name=Ready,type=string,JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.resource\\.ready"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Certificate is the Schema for the certificate API
type Certificate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertificateSpec   `json:"spec,omitempty"`
	Status CertificateStatus `json:"status,omitempty"`
}

func (ca *Certificate) GetStatus() *reconciler.Status {
	return &ca.Status.Status
}

// +kubebuilder:object:root=true

type CertificateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Certificate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Certificate{}, &CertificateList{})
}
