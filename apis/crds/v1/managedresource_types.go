package v1

import (
	"fmt"

	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MsvcNamedRef struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
}

type mresKind struct {
	Kind string `json:"kind"`
}

type MresResourceTemplate struct {
	metav1.TypeMeta `json:",inline"`
	MsvcRef         MsvcNamedRef                    `json:"msvcRef"`
	Spec            map[string]apiextensionsv1.JSON `json:"spec"`
}

// ManagedResourceSpec defines the desired state of ManagedResource
type ManagedResourceSpec struct {
	ResourceTemplate MresResourceTemplate `json:"resourceTemplate"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource-gvk",name=Resource_GVK,type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ManagedResource is the Schema for the managedresources API
type ManagedResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ManagedResourceSpec `json:"spec"`
	// +kubebuilder:default=true
	Enabled *bool       `json:"enabled,omitempty"`
	Status  rApi.Status `json:"status,omitempty" graphql:"noinput"`
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
		"kloudlite.io/msvc.name": m.Spec.ResourceTemplate.MsvcRef.Name,
		"kloudlite.io/mres.name": m.Name,
	}
}

func (m *ManagedResource) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("ManagedResource").String(),
		"kloudlite.io/resource-gvk":               m.Spec.ResourceTemplate.GroupVersionKind().String(),
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
