package v1

import (
	"fmt"

	"github.com/kloudlite/operator/pkg/constants"
	"github.com/kloudlite/operator/toolkit/plugin"
	rApi "github.com/kloudlite/operator/toolkit/reconciler"

	fn "github.com/kloudlite/operator/pkg/functions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// +kubebuilder:object:generate=true
type ServiceTemplate struct {
	APIVersion string                          `json:"apiVersion"`
	Kind       string                          `json:"kind"`
	Spec       map[string]apiextensionsv1.JSON `json:"spec,omitempty"`
	Export     plugin.Export                   `json:"export,omitempty"`
}

func (s *ServiceTemplate) GroupVersionKind() schema.GroupVersionKind {
	return fn.ParseGVK(s.APIVersion, s.Kind)
}

// ManagedServiceSpec defines the desired state of ManagedService
type ManagedServiceSpec struct {
	ServiceTemplate *ServiceTemplate `json:"serviceTemplate,omitempty"`
	Plugin          *ServiceTemplate `json:"plugin,omitempty"`
}

func (obj *ManagedService) PatchWithDefaults() (hasPatched bool) {
	hasPatched = false

	if obj.Spec.Plugin != nil && obj.Spec.Plugin.Export.ViaSecret == "" {
		hasPatched = true
		obj.Spec.Plugin.Export.ViaSecret = obj.Name + "-export"
	}

	return hasPatched
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/service-gvk",name=Service GVK,type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ManagedService is the Schema for the managedservices API
type ManagedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ManagedServiceSpec `json:"spec"`

	Status rApi.Status `json:"status,omitempty" graphql:"noinput"`
}

func (m *ManagedService) EnsureGVK() {
	if m != nil {
		m.SetGroupVersionKind(GroupVersion.WithKind("ManagedService"))
	}
}

func (m *ManagedService) GetStatus() *rApi.Status {
	return &m.Status
}

func (m *ManagedService) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.MsvcNameKey: m.Name,
	}
}

func (m *ManagedService) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		"kloudlite.io/service-gvk": func() string {
			if m.Spec.ServiceTemplate != nil {
				return fmt.Sprintf("%s|%s", m.Spec.ServiceTemplate.APIVersion, m.Spec.ServiceTemplate.Kind)
				// return m.Spec.ServiceTemplate.GroupVersionKind().String()
			}

			if m.Spec.Plugin != nil {
				return fmt.Sprintf("%s|%s", m.Spec.Plugin.APIVersion, m.Spec.Plugin.Kind)
				// return m.Spec.Plugin.GroupVersionKind().String()
			}

			return ""
		}(),
	}
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
