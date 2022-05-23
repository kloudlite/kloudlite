package v1

import (
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"operators.kloudlite.io/lib"
	fn "operators.kloudlite.io/lib/functions"
	t "operators.kloudlite.io/lib/types"
)

// ManagedResourceSpec defines the desired state of ManagedResource
type ManagedResourceSpec struct {
	ApiVersion     string `json:"apiVersion"`
	Kind           string `json:"kind"`
	ManagedSvcName string `json:"managedSvcName"`
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	Inputs json.RawMessage `json:"inputs,omitempty"`
}

// ManagedResourceStatus defines the observed state of ManagedResource
type ManagedResourceStatus struct {
	LastHash   string       `json:"lastHash,omitempty"`
	Conditions t.Conditions `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ManagedResource is the Schema for the managedresources API
type ManagedResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedResourceSpec   `json:"spec,omitempty"`
	Status ManagedResourceStatus `json:"status,omitempty"`
}

func (m *ManagedResource) NameRef() string {
	return fmt.Sprintf("%s/%s/%s", m.GroupVersionKind().Group, m.Namespace, m.Name)
}

const (
	OfMsvcLabelKey lib.LabelKey = "mres.kloudlite.io/of-msvc"
)

func (m ManagedResource) LabelRef() (key, value string) {
	return "mres.kloudlite.io/for", GroupVersion.Group
}

func (m *ManagedResource) HasLabels() bool {
	k, v := m.LabelRef()
	if v != m.Labels[k] {
		return false
	}
	return true
}

func (m *ManagedResource) EnsureLabels() {
	k, v := m.LabelRef()
	m.SetLabels(map[string]string{k: v})
}

func (s *ManagedResource) Hash() string {
	m := make(map[string]interface{}, 3)
	m["name"] = s.Name
	m["namespace"] = s.Namespace
	m["spec"] = s.Spec
	hash, _ := fn.Json.Hash(m)
	return hash
}

func (mr *ManagedResource) OwnedByMsvc(svc *ManagedService) bool {
	for _, c := range mr.OwnerReferences {
		if c.APIVersion == svc.APIVersion && c.Kind == svc.Kind && c.Name == svc.Name && c.UID == svc.UID {
			return true
		}
	}
	return false
}

// +kubebuilder:object:root=true

// ManagedResourceList contains a list of ManagedResource
type ManagedResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ManagedResource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ManagedResource{}, &ManagedResourceList{})
}
