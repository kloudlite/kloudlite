package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AccountSpec struct {
	HarborProjectName string `json:"harborProjectName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Account is the Schema for the accounts API
type Account struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccountSpec `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (acc *Account) EnsureGVK() {
	if acc != nil {
		acc.SetGroupVersionKind(GroupVersion.WithKind("Account"))
	}
}

func (acc *Account) GetStatus() *rApi.Status {
	return &acc.Status
}

func (acc *Account) GetEnsuredLabels() map[string]string {
	m := map[string]string{
		"kloudlite.io/account.name": acc.Name,
	}

	return m
}

func (acc *Account) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("Account").String(),
	}
}

//+kubebuilder:object:root=true

// AccountList contains a list of Account
type AccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Account `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Account{}, &AccountList{})
}
