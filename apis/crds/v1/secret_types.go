package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date
// +kubebuilder:printcolumn:JSONPath=".type",name=Type,type=string

// Secret is the Schema for the secrets API
type Secret struct {
	corev1.Secret `json:",inline"`

	//metav1.TypeMeta   `json:",inline"`
	//metav1.ObjectMeta `json:"metadata,omitempty"`

	ProjectName string `json:"projectName,omitempty"`

	//Type       corev1.SecretType `json:"type,omitempty"`
	//Data       map[string][]byte `json:"data,omitempty"`
	//StringData map[string]string `json:"stringData,omitempty"`
	Overrides *JsonPatch `json:"overrides,omitempty"`

	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	Status rApi.Status `json:"status,omitempty"`
}

func (scrt *Secret) EnsureGVK() {
	if scrt != nil {
		scrt.SetGroupVersionKind(GroupVersion.WithKind("Secret"))
	}
}

func (scrt *Secret) GetGVK() schema.GroupVersionKind {
	return GroupVersion.WithKind("Secret")
}

func (scrt *Secret) GetStatus() *rApi.Status {
	return &scrt.Status
}

func (scrt *Secret) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (scrt *Secret) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: scrt.GroupVersionKind().String(),
	}
}

//+kubebuilder:object:root=true

// SecretList contains a list of Secret
type SecretList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Secret `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Secret{}, &SecretList{})
}
