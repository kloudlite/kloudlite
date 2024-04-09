package v1

import (
	ct "github.com/kloudlite/operator/apis/common-types"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GcpVPCSpec defines the desired state of GcpVPC
type GcpVPCSpec struct {
	GCPProjectID string `json:"gcpProjectID"`
	Region       string `json:"region"`

	// This secret will be unmarshalled into type GCPCredentials
	CredentialsRef ct.SecretRef `json:"credentialsRef"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// GcpVPC is the Schema for the gcpvpcs API
type GcpVPC struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GcpVPCSpec  `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (gv *GcpVPC) EnsureGVK() {
	if gv != nil {
		gv.SetGroupVersionKind(GroupVersion.WithKind("GcpVPC"))
	}
}

func (p *GcpVPC) GetStatus() *rApi.Status {
	return &p.Status
}

func (p *GcpVPC) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (p *GcpVPC) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// GcpVPCList contains a list of GcpVPC
type GcpVPCList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GcpVPC `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GcpVPC{}, &GcpVPCList{})
}
