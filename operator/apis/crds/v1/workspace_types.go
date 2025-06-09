package v1

import (
	"github.com/kloudlite/operator/toolkit/reconciler"
	rApi "github.com/kloudlite/operator/toolkit/reconciler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:validation:Enum=ON;OFF;
// +kubebuilder:default=ON
type WorkspaceState string

const (
	WorkspaceStateOn  WorkspaceState = "ON"
	WorkspaceStateOff WorkspaceState = "OFF"
)

// WorkspaceSpec defines the desired state of Workspace
type WorkspaceSpec struct {
	// Name of work machine
	WorkMachine string `json:"workMachine" graphql:"noinput"`

	// +kubebuilder:default=ON
	State              WorkspaceState `json:"state"`
	ServiceAccountName string         `json:"serviceAccountName,omitempty"`

	EnableTTYD            bool `json:"enableTTYD,omitempty"`
	EnableJupyterNotebook bool `json:"enableJupyterNotebook,omitempty"`
	EnableCodeServer      bool `json:"enableCodeServer,omitempty"`
	EnableVSCodeServer    bool `json:"enableVSCodeServer,omitempty"`

	// +kubebuilder:default=IfNotPresent
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	// Router RouterSpec `json:"router"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Workspace is the Schema for the workspaces API
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceSpec `json:"spec,omitempty"`
	Status rApi.Status   `json:"status,omitempty"`
}

func (r *Workspace) EnsureGVK() {
	if r != nil {
		r.SetGroupVersionKind(GroupVersion.WithKind("Workspace"))
	}
}

func (w *Workspace) GetStatus() *reconciler.Status {
	return &w.Status
}

func (w *Workspace) GetEnsuredLabels() map[string]string {
	return map[string]string{
		"kloudlite.io/workspace.name": w.Name,
	}
}

func (w *Workspace) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

// +kubebuilder:object:root=true

// WorkspaceList contains a list of Workspace
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workspace{}, &WorkspaceList{})
}
