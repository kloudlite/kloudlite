package v1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rApi "operators.kloudlite.io/lib/operator"
	rawJson "operators.kloudlite.io/lib/raw-json"
)

// ACLAccountSpec defines the desired state of ACLAccount
type ACLAccountSpec struct {
	Inputs         rawJson.KubeRawJson `json:"inputs,omitempty"`
	ManagedSvcName string              `json:"managedSvcName"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ACLAccount is the Schema for the aclaccounts API
type ACLAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ACLAccountSpec `json:"spec,omitempty"`
	Status rApi.Status    `json:"status,omitempty"`
}

func (ac *ACLAccount) GetStatus() *rApi.Status {
	return &ac.Status
}

func (ac *ACLAccount) GetEnsuredLabels() map[string]string {
	return map[string]string{
		fmt.Sprintf("%s/ref", GroupVersion.Group): ac.Name,
	}
}

//+kubebuilder:object:root=true

// ACLAccountList contains a list of ACLAccount
type ACLAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ACLAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ACLAccount{}, &ACLAccountList{})
}
