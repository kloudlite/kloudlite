package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceTemplate struct {
	metav1.TypeMeta `json:",inline"`
	Spec            map[string]apiextensionsv1.JSON `json:"spec"`
}

// ManagedServiceSpec defines the desired state of ManagedService
type ManagedServiceSpec struct {
	ServiceTemplate ServiceTemplate `json:"serviceTemplate"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/service-gvk",name=Service GVK,type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// ManagedService is the Schema for the managedservices API
type ManagedService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ManagedServiceSpec `json:"spec"`
	// json.RawMessage

	// +kubebuilder:default=true
	Enabled *bool       `json:"enabled,omitempty"`
	Status  rApi.Status `json:"status,omitempty" graphql:"noinput"`
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
		"kloudlite.io/service-gvk":                m.Spec.ServiceTemplate.GroupVersionKind().String(),
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
