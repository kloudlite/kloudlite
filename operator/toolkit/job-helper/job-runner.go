package job_helper

import (
	"context"
	"errors"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type JobTracker struct {
	ctx    context.Context
	client client.Client

	job *batchv1.Job

	isTargetJob func(job *batchv1.Job) bool
}

type JobTrackerArgs struct {
	client client.Client

	JobNamespace string
	JobName      string

	IsTargetJob func(job *batchv1.Job) bool
}

func NewJobTracker(ctx context.Context, kcli client.Client, args JobTrackerArgs) (*JobTracker, error) {
	var job batchv1.Job
	if err := kcli.Get(ctx, types.NamespacedName{Name: args.JobName, Namespace: args.JobNamespace}, &job); err != nil {
		return nil, err
	}

	return &JobTracker{
		ctx:         ctx,
		client:      kcli,
		job:         &job,
		isTargetJob: args.IsTargetJob,
	}, nil
}

func (jr *JobTracker) HasJobFinished() bool {
	for _, v := range jr.job.Status.Conditions {
		if v.Type == batchv1.JobComplete && v.Status == corev1.ConditionTrue {
			return true
		}

		if v.Type == batchv1.JobFailed && v.Status == corev1.ConditionTrue {
			return true
		}

		if v.Type == batchv1.JobSuspended && v.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}

func (jr *JobTracker) StartTracking(ctx context.Context) (phase JobPhase, message string, err error) {
	if !jr.isTargetJob(jr.job) {
		if !jr.HasJobFinished() {
			return JobPhasePending, "waiting for previous jobs to finish execution", nil
		}

		if err := DeleteJob(jr.ctx, jr.client, jr.job.Namespace, jr.job.Name); err != nil {
			return JobPhasePending, "", errors.Join(errors.New("failed to delete job"), err)
		}

		return JobPhasePending, "waiting for job to start", nil
	}

	if !jr.HasJobFinished() {
		return JobPhaseRunning, "waiting for running job to finish", nil
	}

	// check.Message = job_manager.GetTerminationLog(ctx, r.Client, job.Namespace, job.Name)
	if jr.job.Status.Succeeded > 0 {
		return JobPhaseSucceeded, "", nil
	}

	if jr.job.Status.Active > 0 {
		return JobPhaseRunning, "waiting for job to complete", nil
	}

	if jr.job.Status.Failed > 0 {
		return JobPhaseFailed, "", errors.New("install or upgrade job failed")
	}

	return JobPhasePending, "", nil
}
