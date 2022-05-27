package v1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rawJson "operators.kloudlite.io/lib/raw-json"
)

type DatabaseSpec struct {
	ManagedSvcName string              `json:"managedSvcName,omitempty"`
	Inputs         rawJson.KubeRawJson `json:"inputs,omitempty"`
}

type DatabaseStatus struct {
	LastHash string `json:"lastHash,omitempty"`

	GeneratedVars rawJson.KubeRawJson `json:"generatedVars,omitempty"`
	Conditions    []metav1.Condition  `json:"conditions,omitempty"`
	OpsConditions []metav1.Condition  `json:"opsConditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Database is the Schema for the databases API
type Database struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseSpec   `json:"spec,omitempty"`
	Status DatabaseStatus `json:"status,omitempty"`
}

func (s *Database) NameRef() string {
	return fmt.Sprintf("%s/%s/%s", s.GroupVersionKind().Group, s.Namespace, s.Name)
}

func (s Database) LabelRef() (string, string) {
	return "mres.kloudlite.io/for", GroupVersion.Group
}

func (s *Database) HasLabels() bool {
	key, value := s.LabelRef()
	if s.Labels[key] != value {
		return false
	}
	return true
}

func (s *Database) EnsureLabels() {
	key, value := s.LabelRef()
	s.SetLabels(map[string]string{key: value})
}

// +kubebuilder:object:root=true

// DatabaseList contains a list of Database
type DatabaseList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Database `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Database{}, &DatabaseList{})
}
