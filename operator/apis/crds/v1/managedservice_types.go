package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rawJson "github.com/kloudlite/operator/pkg/raw-json"
)

type msvcKind struct {
	APIVersion string `json:"apiVersion"`
	// +kubebuilder:default=Service
	// +kubebuilder:validation:Optional
	Kind string `json:"kind"`
}

// ManagedServiceSpec defines the desired state of ManagedService
type ManagedServiceSpec struct {
	Region string `json:"region,omitempty"`

	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
	MsvcKind     msvcKind            `json:"msvcKind"`

	Inputs rawJson.RawJson `json:"inputs,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".spec.msvcKind.apiVersion",name=~API,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.msvcKind.kind",name=~Kind,type=string
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ManagedService is the Schema for the managedservices API
type ManagedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ManagedServiceSpec `json:"spec,omitempty"`

	// +kubebuilder:default=true
	Enabled *bool       `json:"enabled,omitempty"`
	Status  rApi.Status `json:"status,omitempty"`
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
		"kloudlite.io/msvc.name": m.Name,
	}
}

func (m *ManagedService) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.AnnotationKeys.GroupVersionKind: GroupVersion.WithKind("ManagedService").String(),
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
