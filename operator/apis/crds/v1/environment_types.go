package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnvironmentSpec defines the desired state of Environment
type EnvironmentSpec struct {
	ProjectName     string `json:"projectName"`
	TargetNamespace string `json:"targetNamespace,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:JSONPath=".spec.projectName",name=Project,type=string
//+kubebuilder:printcolumn:JSONPath=".spec.targetNamespace",name="target-namespace",type=string
//+kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Last_Reconciled_At,type=date
//+kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
//+kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Environment is the Schema for the environments API
type Environment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvironmentSpec `json:"spec,omitempty"`
	Status rApi.Status     `json:"status,omitempty" graphql:"noinput"`
}

func (e *Environment) EnsureGVK() {
	if e != nil {
		e.SetGroupVersionKind(GroupVersion.WithKind("Environment"))
	}
}

func (e *Environment) GetStatus() *rApi.Status {
	return &e.Status
}

func (e *Environment) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.ProjectNameKey:     e.Spec.ProjectName,
		constants.TargetNamespaceKey: e.Spec.TargetNamespace,
	}
}

func (e *Environment) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// EnvironmentList contains a list of Environment
type EnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Environment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Environment{}, &EnvironmentList{})
}
