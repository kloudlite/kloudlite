package v1

import (
	"encoding/json"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	fn "operators.kloudlite.io/lib/functions"
)

// ManagedServiceSpec defines the desired state of ManagedService
type ManagedServiceSpec struct {
	ApiVersion string `json:"apiVersion"`
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	Inputs json.RawMessage `json:"inputs,omitempty"`
}

// ManagedServiceStatus defines the observed state of ManagedService
type ManagedServiceStatus struct {
	LastHash   string             `json:"lastHash,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ManagedService is the Schema for the managedservices API
type ManagedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedServiceSpec   `json:"spec,omitempty"`
	Status ManagedServiceStatus `json:"status,omitempty"`
}

func (m *ManagedService) NameRef() string {
	return fmt.Sprintf("%s/%s/%s", m.Namespace, m.Kind, m.Name)
}

func (m ManagedService) LabelRef() (key, value string) {
	return "msvc.kloudlite.io/for", strings.Split(m.Spec.ApiVersion, "/")[0]
}

func (m *ManagedService) HasLabels() bool {
	key, value := m.LabelRef()
	if m.Labels[key] != value {
		return false
	}
	return true
}

func (m *ManagedService) EnsureLabels() {
	key, value := m.LabelRef()
	m.SetLabels(map[string]string{key: value})
}

func (s *ManagedService) Hash() (string, error) {
	m := make(map[string]interface{}, 3)
	m["name"] = s.Name
	m["namespace"] = s.Namespace
	m["spec"] = s.Spec
	hash, err := fn.Json.Hash(m)
	return hash, err
}

// +kubebuilder:object:root=true

// ManagedServiceList contains a list of ManagedService
type ManagedServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedService{}, &ManagedServiceList{})
}
