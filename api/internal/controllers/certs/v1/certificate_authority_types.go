package v1

import (
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CertificateAuthoritySpec struct {
	SANs []string `json:"san"`
}

type SecretKeyRef struct {
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	CaBundleKey  string `json:"caBundleKey"`
	CaPrivateKey string `json:"caPrivateKey"`
}

type CertificateAuthorityStatus struct {
	reconciler.Status `json:",inline"`

	SecretRef SecretKeyRef `json:"secretRef,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={kloudlite}
// +kubebuilder:printcolumn:name="Seen",type=date,JSONPath=".status.lastReconcileTime"
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.resource\\.ready"
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// CertificateAuthority is the Schema for the certificate authority API
type CertificateAuthority struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CertificateAuthoritySpec   `json:"spec,omitempty"`
	Status CertificateAuthorityStatus `json:"status,omitempty"`
}

func (ca *CertificateAuthority) GetStatus() *reconciler.Status {
	return &ca.Status.Status
}

// +kubebuilder:object:root=true

// ServiceInterceptList contains a list of ServiceIntercept
type CertificateAuthorityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CertificateAuthority `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CertificateAuthority{}, &CertificateAuthorityList{})
}
