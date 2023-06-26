package v1

import (
	"fmt"

	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rawJson "github.com/kloudlite/operator/pkg/raw-json"
)

type msvcNamedRefTT struct {
	APIVersion string `json:"apiVersion"`
	// +kubebuilder:default=Service
	// +kubebuilder:validation:Optional
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type mresKind struct {
	Kind string `json:"kind"`
}

// ManagedResourceSpec defines the desired state of ManagedResource
type ManagedResourceSpec struct {
	MsvcRef  msvcNamedRefTT  `json:"msvcRef"`
	MresKind mresKind        `json:"mresKind"`
	Inputs   rawJson.RawJson `json:"inputs,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ManagedResource is the Schema for the managedresources API
type ManagedResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ManagedResourceSpec `json:"spec,omitempty"`
	// +kubebuilder:default=true
	Enabled *bool       `json:"enabled,omitempty"`
	Status  rApi.Status `json:"status,omitempty"`
}

func (m *ManagedResource) EnsureGVK() {
	if m != nil {
		m.SetGroupVersionKind(GroupVersion.WithKind("ManagedResource"))
	}
}

func (m *ManagedResource) NameRef() string {
	return fmt.Sprintf("%s/%s/%s", m.GroupVersionKind().Group, m.Namespace, m.Name)
}

func (m *ManagedResource) GetStatus() *rApi.Status {
	return &m.Status
}

func (m *ManagedResource) GetEnsuredLabels() map[string]string {
	return map[string]string{
		"kloudlite.io/msvc.name": m.Spec.MsvcRef.Name,
		"kloudlite.io/mres.name": m.Name,
	}
}

func (m *ManagedResource) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("ManagedResource").String(),
	}
}

func (m *ManagedResource) OwnedByMsvc(svc *ManagedService) bool {
	for _, c := range m.OwnerReferences {
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
