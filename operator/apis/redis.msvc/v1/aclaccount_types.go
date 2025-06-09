package v1

import (
	ct "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ACLAccountSpec defines the desired state of ACLAccount
type ACLAccountSpec struct {
	KeyPrefix    string     `json:"keyPrefix,omitempty"`
	MsvcRef      ct.MsvcRef `json:"msvcRef"`
	ResourceName string     `json:"resourceName,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ACLAccount is the Schema for the aclaccounts API
type ACLAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ACLAccountSpec `json:"spec"`
	Status rApi.Status    `json:"status,omitempty"`
}

func (a *ACLAccount) EnsureGVK() {
	if a != nil {
		a.SetGroupVersionKind(GroupVersion.WithKind("ACLAccount"))
	}
}

func (a *ACLAccount) GetStatus() *rApi.Status {
	return &a.Status
}

func (a *ACLAccount) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (a *ACLAccount) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("ACLAccount").String(),
	}
}

// +kubebuilder:object:root=true

// ACLAccountList contains a list of ACLAccount
type ACLAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ACLAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ACLAccount{}, &ACLAccountList{})
}
