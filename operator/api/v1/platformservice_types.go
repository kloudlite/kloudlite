package v1

import (
	"fmt"

	"github.com/kloudlite/kloudlite/operator/toolkit/plugin"
	"github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:generate=true
type PluginTemplate struct {
	APIVersion string                          `json:"apiVersion"`
	Kind       string                          `json:"kind"`
	Spec       map[string]apiextensionsv1.JSON `json:"spec,omitempty"`
	Export     plugin.Export                   `json:"export,omitempty"`
}

// PlatformServiceSpec defines the desired state of PlatformService.
type PlatformServiceSpec struct {
	Plugin PluginTemplate `json:"plugin"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/service-gvk",name=Service GVK,type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.operator\\.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.operator\\.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// PlatformService is the Schema for the platformservices API.
type PlatformService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PlatformServiceSpec `json:"spec,omitempty"`
	Status reconciler.Status   `json:"status,omitempty"`
}

func (m *PlatformService) EnsureGVK() {
	if m != nil {
		m.SetGroupVersionKind(GroupVersion.WithKind("PlatformService"))
	}
}

func (m *PlatformService) GetStatus() *reconciler.Status {
	return &m.Status
}

func (m *PlatformService) GetEnsuredLabels() map[string]string {
	return map[string]string{
		PlatformServiceNameKey: m.Name,
	}
}

func (m *PlatformService) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		PlatformServicePluginGVK: fmt.Sprintf("%s|%s", m.Spec.Plugin.APIVersion, m.Spec.Plugin.Kind),
	}
}

// +kubebuilder:object:root=true

// PlatformServiceList contains a list of PlatformService.
type PlatformServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PlatformService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PlatformService{}, &PlatformServiceList{})
}
