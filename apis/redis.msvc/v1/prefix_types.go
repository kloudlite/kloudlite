package v1

import (
	"fmt"

	ct "github.com/kloudlite/operator/apis/common-types"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PrefixOutput struct {
	Credentials ct.SecretRef `json:"credentials"`
}

// PrefixSpec defines the desired state of Prefix
type PrefixSpec struct {
	MsvcRef   ct.MsvcRef   `json:"msvcRef"`
	PrefixKey string       `json:"prefixKey,omitempty"`
	Output    PrefixOutput `json:"output,omitempty" graphql:"noinput"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/credentials-ref",name=Credentials,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Prefix is the Schema for the prefixes API
type Prefix struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PrefixSpec  `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (a *Prefix) EnsureGVK() {
	if a != nil {
		a.SetGroupVersionKind(GroupVersion.WithKind("Prefix"))
	}
}

func (a *Prefix) GetStatus() *rApi.Status {
	return &a.Status
}
func (a *Prefix) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (a *Prefix) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		"kloudlite.io/credentials-ref": fmt.Sprintf("secret:%s/%s", a.Spec.Output.Credentials.Namespace, a.Spec.Output.Credentials.Name),
	}
}

//+kubebuilder:object:root=true

// PrefixList contains a list of Prefix
type PrefixList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Prefix `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Prefix{}, &PrefixList{})
}
