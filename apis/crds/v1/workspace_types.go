package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
)

type WorkspaceSpec struct {
	ProjectName     string `json:"projectName"`
	TargetNamespace string `json:"targetNamespace,omitempty"`
	IsEnvironment   *bool  `json:"isEnvironment,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".spec.projectName",name=Project,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.targetNamespace",name="target-namespace",type=string
// +kubebuilder:printcolumn:JSONPath=".status.isReady",name=Ready,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Workspace is the Schema for the envs API
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceSpec `json:"spec,omitempty"`
	Status rApi.Status   `json:"status,omitempty" graphql:"noinput"`
}

func (e *Workspace) EnsureGVK() {
	if e != nil {
		e.SetGroupVersionKind(GroupVersion.WithKind("Workspace"))
	}
}

func (e *Workspace) GetStatus() *rApi.Status {
	return &e.Status
}

func (e *Workspace) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.ProjectNameKey:     e.Spec.ProjectName,
		constants.TargetNamespaceKey: e.Spec.TargetNamespace,
	}
}

func (e *Workspace) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

//+kubebuilder:object:root=true

// EnvList contains a list of Workspace
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workspace{}, &WorkspaceList{})
}
