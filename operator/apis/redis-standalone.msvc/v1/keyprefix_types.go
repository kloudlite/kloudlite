package v1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	rawJson "operators.kloudlite.io/lib/raw-json"

	fn "operators.kloudlite.io/lib/functions"
)

type KeyPrefixSpec struct {
	Inputs         rawJson.KubeRawJson `json:"inputs,omitempty"`
	ManagedSvcName string              `json:"managedSvcName"`
}

type KeyPrefixStatus struct {
	GeneratedVars rawJson.KubeRawJson `json:"generatedVars,omitempty"`
	Conditions    []metav1.Condition  `json:"conditions,omitempty"`
	OpsConditions []metav1.Condition  `json:"opsConditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// KeyPrefix is the Schema for the keyprefixes API
type KeyPrefix struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KeyPrefixSpec   `json:"spec,omitempty"`
	Status KeyPrefixStatus `json:"status,omitempty"`
}

func (s *KeyPrefix) NamespacedName() types.NamespacedName {
	return types.NamespacedName{Namespace: s.Namespace, Name: s.Name}
}

func (s *KeyPrefix) NameRef() string {
	return fmt.Sprintf("GroupRef=%s ResourceRef=%s", s.GroupVersionKind().String(), s.NamespacedName().String())
}

func (s KeyPrefix) LabelRef() (key, value string) {
	return "mres.kloudlite.io/group-ref", fmt.Sprintf("%s_%s", s.GroupVersionKind().Group, s.GroupVersionKind().Kind)
}

func (s *KeyPrefix) HasLabels() bool {
	key, value := s.LabelRef()
	if s.Labels[key] != value {
		return false
	}
	return true
}

func (s *KeyPrefix) EnsureLabels() {
	key, value := s.LabelRef()
	s.SetLabels(
		map[string]string{
			key:                     value,
			"msvc.kloudlite.io/ref": s.Spec.ManagedSvcName,
		},
	)
}

func (s *KeyPrefix) Hash() string {
	m := make(map[string]interface{}, 3)
	m["name"] = s.Name
	m["namespace"] = s.Namespace
	m["spec"] = s.Spec
	hash, _ := fn.Json.Hash(m)
	return hash
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
