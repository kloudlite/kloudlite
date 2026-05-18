package functions

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

type jobStatus string

func (j jobStatus) Message() string {
	switch j {
	case JobStatusPending:
		return "waiting for job to start"
	case JobStatusRunning:
		return "waiting for job to finish"
	case JobStatusSucceeded:
		return "job succeeded"
	case JobStatusFailed:
		return "job failed"
	case JobStatusSuspended:
		return "job suspended"
	default:
		return "invalid job status"
	}
}

const (
	JobStatusPending   jobStatus = "PENDING"
	JobStatusRunning   jobStatus = "RUNNING"
	JobStatusFailed    jobStatus = "FAILED"
	JobStatusSucceeded jobStatus = "SUCCEEDED"
	JobStatusSuspended jobStatus = "SUSPENDED"
)

func ParseJobStatus(job *batchv1.Job) jobStatus {
	for _, v := range job.Status.Conditions {
		if v.Type == batchv1.JobComplete && v.Status == corev1.ConditionTrue {
			return JobStatusSucceeded
		}

		if v.Type == batchv1.JobFailed && v.Status == corev1.ConditionTrue {
			return JobStatusFailed
		}

		if v.Type == batchv1.JobSuspended && v.Status == corev1.ConditionTrue {
			return JobStatusSuspended
		}
	}

	if job.Status.Active > 0 {
		return JobStatusRunning
	}

	return JobStatusPending
}
