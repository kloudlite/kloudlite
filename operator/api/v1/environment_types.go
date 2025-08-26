package v1

import (
	"github.com/kloudlite/kloudlite/operator/toolkit/reconciler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EnvironmentSpec defines the desired state of Environment.
type EnvironmentSpec struct {
	// Paused pauses the environment
	Paused bool `json:"paused,omitempty"`

	// TargetNamespace is namespace under which all resources are created
	TargetNamespace string `json:"targetNamespace,omitempty"`

	// ServiceAccount is used for all apps deployed in this environment
	ServiceAccount string `json:"serviceAccount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:JSONPath=".spec.targetNamespace",name="target-ns",type=string
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.operator\\.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".spec.suspend",name=Paused,type=boolean
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.operator\\.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Environment is the Schema for the environments API.
type Environment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EnvironmentSpec   `json:"spec,omitempty"`
	Status reconciler.Status `json:"status,omitempty"`
}

func (e *Environment) EnsureGVK() {
	if e != nil {
		e.SetGroupVersionKind(GroupVersion.WithKind("Environment"))
	}
}

func (e *Environment) GetStatus() *reconciler.Status {
	return &e.Status
}

func (e *Environment) GetEnsuredLabels() map[string]string {
	return map[string]string{}
}

func (e *Environment) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

// +kubebuilder:object:root=true

// EnvironmentList contains a list of Environment.
type EnvironmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Environment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Environment{}, &EnvironmentList{})
}
