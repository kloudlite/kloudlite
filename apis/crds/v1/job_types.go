package v1

import (
	"github.com/kloudlite/operator/pkg/constants"
	rApi "github.com/kloudlite/operator/pkg/operator"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type JobAction struct {
	BackOffLimit *int32         `json:"backOffLimit,omitempty"`
	PodSpec      corev1.PodSpec `json:"podSpec"`
}

// JobSpec defines the desired state of Job
type JobSpec struct {
	OnApply  JobAction  `json:"onApply"`
	OnDelete *JobAction `json:"onDelete,omitempty"`
}

type JobStatus struct {
	rApi.Status `json:",inline"`
	Succeeded   *bool `json:"succeeded,omitempty"`
	Running     *bool `json:"running,omitempty"`
	Failed      *bool `json:"failed,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:JSONPath=".status.lastReconcileTime",name=seen,type=date
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/checks",name=Checks,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/job\\.phase",name=phase,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.annotations.kloudlite\\.io\\/resource\\.ready",name=Ready,type=string
// +kubebuilder:printcolumn:JSONPath=".metadata.creationTimestamp",name=Age,type=date

// Job is the Schema for the jobs API
type Job struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JobSpec   `json:"spec,omitempty"`
	Status JobStatus `json:"status,omitempty"`
}

func (j *Job) EnsureGVK() {
	if j != nil {
		j.SetGroupVersionKind(GroupVersion.WithKind("Job"))
	}
}

func (p *Job) GetStatus() *rApi.Status {
	return &p.Status.Status
}

func (p *Job) GetEnsuredLabels() map[string]string {
	return map[string]string{
		constants.JobNameKey: p.Name,
	}
}

func (p *Job) GetEnsuredAnnotations() map[string]string {
	phase := "PENDING"
	if p.Status.Running != nil && *p.Status.Running {
		phase = "RUNNING"
	}
	if p.Status.Failed != nil && *p.Status.Failed {
		phase = "FAILED"
	}
	if p.Status.Succeeded != nil && *p.Status.Succeeded {
		phase = "SUCCEEDED"
	}

	return map[string]string{
		"kloudlite.io/job.phase": phase,
	}
}

//+kubebuilder:object:root=true

// JobList contains a list of Job
type JobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Job `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Job{}, &JobList{})
}
