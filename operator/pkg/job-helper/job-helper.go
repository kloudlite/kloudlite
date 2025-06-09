package job_helper

import (
	"context"
	"encoding/json"
	"fmt"

	fn "github.com/kloudlite/operator/pkg/functions"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiLabels "k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type JobStatus string

const (
	JobStatusUnknown JobStatus = "job-status-unknown"
	JobStatusRunning JobStatus = "job-status-running"
	JobStatusFailed  JobStatus = "job-status-failed"
	JobStatusSuccess JobStatus = "job-status-success"
)

func GetLatestPod(ctx context.Context, cli client.Client, jobNamespace string, jobName string) (*corev1.Pod, error) {
	var podList corev1.PodList

	if err := cli.List(ctx, &podList, &client.ListOptions{
		Namespace: jobNamespace,
		LabelSelector: apiLabels.SelectorFromValidatedSet(map[string]string{
			"job-name": jobName,
		}),
	}); err != nil {
		return nil, err
	}

	if len(podList.Items) == 0 {
		return nil, fmt.Errorf("no pods found")
	}

	var latestPod corev1.Pod
	for i := range podList.Items {
		if podList.Items[i].CreationTimestamp.Unix()-latestPod.CreationTimestamp.Unix() > 0 {
			latestPod = podList.Items[i]
		}
	}

	return &latestPod, nil
}

func HasJobFinished(ctx context.Context, cli client.Client, job *batchv1.Job) bool {
	for _, v := range job.Status.Conditions {
		if v.Type == batchv1.JobComplete && v.Status == "True" {
			return true
		}

		if v.Type == batchv1.JobFailed && v.Status == "True" {
			return true
		}

		if v.Type == batchv1.JobSuspended && v.Status == "True" {
			return true
		}
	}

	p, err := GetLatestPod(ctx, cli, job.Namespace, job.Name)
	if err != nil {
		return false
	}

	if p.Status.Phase == corev1.PodSucceeded || p.Status.Phase == corev1.PodFailed {
		return true
	}

	return false
}

func GetTerminationLog(ctx context.Context, cli client.Client, jobNamespace string, jobName string) string {
	m := map[string]string{}

	pod, err := GetLatestPod(ctx, cli, jobNamespace, jobName)
	if err != nil {
		return ""
	}

	for i := range pod.Status.ContainerStatuses {
		cs := pod.Status.ContainerStatuses[i]
		if cs.State.Terminated != nil {
			m[cs.Name] = cs.State.Terminated.Message
		}
	}

	b, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func DeleteJob(ctx context.Context, cli client.Client, jobNamespace string, jobName string) error {
	if err := cli.Delete(ctx, &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: jobNamespace,
		},
	}, &client.DeleteOptions{
		GracePeriodSeconds: fn.New(int64(10)),
		Preconditions:      &metav1.Preconditions{},
		PropagationPolicy:  fn.New(metav1.DeletePropagationBackground),
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return err
		}
	}

	return nil
}
