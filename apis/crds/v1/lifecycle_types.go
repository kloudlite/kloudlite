package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type LifecycleAction struct {
	BackOffLimit *int32         `json:"backOffLimit,omitempty"`
	PodSpec      corev1.PodSpec `json:"podSpec"`
}

// LifecycleSpec defines the desired state of Lifecycle
type LifecycleSpec struct {
	//+kubebuilder:default=true
	RetryOnFailure bool `json:"retryOnFailure,omitempty"`
	//+kubebuilder:default="30s"
	RetryOnFailureDelay metav1.Duration  `json:"retryOnFailureDelay,omitempty"`
	OnApply             LifecycleAction  `json:"onApply"`
	OnDelete            *LifecycleAction `json:"onDelete,omitempty"`
}

type JobPhase string

const (
	JobPhasePending   JobPhase = "PENDING"
	JobPhaseRunning   JobPhase = "RUNNING"
	JobPhaseFailed    JobPhase = "FAILED"
	JobPhaseSucceeded JobPhase = "SUCCEEDED"
)

type LifecycleStatus struct {
	rApi.Status `json:",inline"`
	Phase       JobPhase `json:"phase"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=Seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".status.phase",name=phase,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/operator\\.resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Lifecycle is the Schema for the Lifecycle API
type Lifecycle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LifecycleSpec   `json:"spec,omitempty"`
	Status LifecycleStatus `json:"status,omitempty"`
}

func (lf *Lifecycle) EnsureGVK() {
	if lf != nil {
		lf.SetGroupVersionKind(GroupVersion.WithKind("Lifecycle"))
	}
}

func (lf *Lifecycle) GetStatus() *rApi.Status {
	return &lf.Status.Status
}

func (lf *Lifecycle) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.JobNameKey: lf.Name,
	}
}

func (lf *Lifecycle) GetEnsuredAnnotations() map[string]string {
	return map[string]string{}
}

func (lf *Lifecycle) HasCompleted() bool {
	return lf.Status.Phase == JobPhaseFailed || lf.Status.Phase == JobPhaseSucceeded
}

//+kubebuilder:object:root=true

// LifecycleList contains a list of Lifecycle
type LifecycleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Lifecycle `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Lifecycle{}, &LifecycleList{})
}
