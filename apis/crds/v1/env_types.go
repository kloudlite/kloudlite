package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

type EnvSpec struct {
	ProjectName     string `json:"projectName"`
	TargetNamespace string `json:"targetNamespace,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date
// +kubebuilder:printcolumn:JSONPath=".spec.projectName",name=Project,type=string

// Env is the Schema for the envs API
type Env struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvSpec     `json:"spec,omitempty"`
	Status rApi.Status `json:"status,omitempty"`
}

func (e *Env) EnsureGVK() {
	if e != nil {
		e.SetGroupVersionKind(GroupVersion.WithKind("Env"))
	}
}

func (e *Env) GetStatus() *rApi.Status {
	return &e.Status
}

func (e *Env) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.ProjectNameKey: e.Spec.ProjectName,
	}
}

func (e *Env) GetEnsuredAnnotations() map[string]string {
	return map[string]string{
		constants.GVKKey: e.GroupVersionKind().String(),
	}
}

//+kubebuilder:object:root=true

// EnvList contains a list of Env
type EnvList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Env `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Env{}, &EnvList{})
}
