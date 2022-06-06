package v1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rApi "operators.kloudlite.io/lib/operator"
	rawJson "operators.kloudlite.io/lib/raw-json"
)

type KeyPrefixSpec struct {
	Inputs         rawJson.KubeRawJson `json:"inputs,omitempty"`
	ManagedSvcName string              `json:"managedSvcName"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// KeyPrefix is the Schema for the keyprefixes API
type KeyPrefix struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KeyPrefixSpec `json:"spec,omitempty"`
	Status rApi.Status   `json:"status,omitempty"`
}

func (s *KeyPrefix) GetStatus() *rApi.Status {
	return &s.Status
}

func (s *KeyPrefix) GetEnsuredLabels() map[string]string {
	return map[string]string{
		fmt.Sprintf("%s/ref", GroupVersion.Group): s.Name,
	}
}

// +kubebuilder:object:root=true

// KeyPrefixList contains a list of KeyPrefix
type KeyPrefixList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KeyPrefix `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KeyPrefix{}, &KeyPrefixList{})
}
