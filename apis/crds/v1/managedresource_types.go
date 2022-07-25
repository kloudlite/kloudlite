package v1

import (
	"fmt"
	"operators.kloudlite.io/lib/constants"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	libOperator "operators.kloudlite.io/lib/operator"
	rawJson "operators.kloudlite.io/lib/raw-json"
)

type msvcNamedRefTT struct {
	APIVersion string `json:"apiVersion"`
	// +kubebuilder:default=Service
	// +kubebuilder:validation:Optional
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type resRefTT struct {
	Kind string `json:"kind"`
}

// ManagedResourceSpec defines the desired state of ManagedResource
type ManagedResourceSpec struct {
	MsvcRef msvcNamedRefTT `json:"msvcRef"`
	ResRef  resRefTT       `json:"resRef"`
	// ApiVersion     string              `json:"apiVersion"`
	// Kind           string              `json:"kind"`
	// ManagedSvcName string              `json:"managedSvcName"`
	Inputs rawJson.KubeRawJson `json:"inputs,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ManagedResource is the Schema for the managedresources API
type ManagedResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ManagedResourceSpec `json:"spec,omitempty"`
	Status            libOperator.Status  `json:"status,omitempty"`
}

func (m *ManagedResource) NameRef() string {
	return fmt.Sprintf("%s/%s/%s", m.GroupVersionKind().Group, m.Namespace, m.Name)
}

func (m *ManagedResource) GetStatus() *libOperator.Status {
	return &m.Status
}

func (m *ManagedResource) GetEnsuredLabels() map[string]string {
	splits := strings.Split(m.Spec.MsvcRef.APIVersion, "/")
	return map[string]string{
		"msvc.kloudlite.io/ref.group":   splits[0],
		"msvc.kloudlite.io/ref.version": splits[1],
		"msvc.kloudlite.io/ref.name":    m.Name,
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
