package v1

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"operators.kloudlite.io/lib"
)

// ManagedResourceSpec defines the desired state of ManagedResource
type ManagedResourceSpec struct {
	Type       string            `json:"type"`
	ManagedSvc string            `json:"managedSvc"`
	Inputs     map[string]string `json:"inputs,omitempty"`
}

// ManagedResourceStatus defines the observed state of ManagedResource
type ManagedResourceStatus struct {
	Conditions []metav1.Condition `json:"conditions,omitempty"`
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

func (m *ManagedResource) LogRef() string {
	return fmt.Sprintf("%s/%s/%s", m.Namespace, m.Kind, m.Name)
}

const (
	OfMsvcLabelKey lib.LabelKey = "mres.kloudlite.io/of-msvc"
)

func (m *ManagedResource) HasLabels() bool {
	if s := m.Labels[OfMsvcLabelKey.String()]; s != m.Spec.ManagedSvc {
		return false
	}
	return true
}

func (m *ManagedResource) EnsureLabels() {
	m.SetLabels(map[string]string{
		OfMsvcLabelKey.String(): m.Spec.ManagedSvc,
	})
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
