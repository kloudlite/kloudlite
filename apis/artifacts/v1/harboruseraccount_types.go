package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/lib/constants"
	rApi "operators.kloudlite.io/lib/operator"
)

// HarborUserAccountSpec defines the desired state of HarborUserAccount
type HarborUserAccountSpec struct {
	Disable    bool   `json:"disable,omitempty"`
	ProjectRef string `json:"projectRef"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// HarborUserAccount is the Schema for the harboruseraccounts API
type HarborUserAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HarborUserAccountSpec `json:"spec,omitempty"`
	Status rApi.Status           `json:"status,omitempty"`
}

func (h *HarborUserAccount) GetStatus() *rApi.Status {
	return &h.Status
}

func (h *HarborUserAccount) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.LabelKeys.HarborProjectRef: h.Spec.ProjectRef,
	}
}

func (h *HarborUserAccount) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

// +kubebuilder:object:root=true

// HarborUserAccountList contains a list of HarborUserAccount
type HarborUserAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HarborUserAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HarborUserAccount{}, &HarborUserAccountList{})
}
