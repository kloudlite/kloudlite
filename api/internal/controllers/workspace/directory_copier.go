package workspace

import (
	"context"
	"fmt"
	"time"

	workmachinev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// copierImage is the alpine image used for sender and receiver jobs
	copierImage = "alpine:latest"

	// copyJobTTLSeconds is the time to live for completed copy jobs (5 minutes)
	copyJobTTLSeconds = 300

	// senderHTTPPort is the port used by the sender to serve the tar.gz file
	senderHTTPPort = 8080
)

// WorkspaceDirectoryCopier handles copying workspace directories between WorkMachines
type WorkspaceDirectoryCopier struct {
	client.Client
	Logger *zap.Logger
}

// createSenderJob creates a job that tar.gz the workspace directory and serves it via HTTP
// The job runs on the source WorkMachine node and mounts the source workspace directory
func (c *WorkspaceDirectoryCopier) createSenderJob(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	sourceWorkspace *workspacev1.Workspace,
	logger *zap.Logger,
) (*batchv1.Job, error) {
	jobName := fmt.Sprintf("ws-clone-sender-%s", workspace.Name)
	namespace, err := c.getJobNamespace(sourceWorkspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get job namespace: %w", err)
	}
	sourceFolderPath := fmt.Sprintf("/var/lib/kloudlite/home/workspaces/%s", sourceWorkspace.Name)

	logger.Info("Creating sender job",
		zap.String("jobName", jobName),
		zap.String("namespace", namespace),
		zap.String("sourceFolderPath", sourceFolderPath),
		zap.String("sourceNode", sourceWorkspace.Spec.WorkmachineName))

	// Get source WorkMachine to determine the node
	sourceWorkmachine, err := c.getWorkMachine(ctx, sourceWorkspace.Spec.WorkmachineName)
	if err != nil {
		return nil, fmt.Errorf("failed to get source WorkMachine: %w", err)
	}

	ttlSeconds := int32(copyJobTTLSeconds)
	backoffLimit := int32(3)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				"kloudlite.io/job-type":         "workspace-clone-sender",
				"kloudlite.io/target-workspace": workspace.Name,
				"kloudlite.io/source-workspace": sourceWorkspace.Name,
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttlSeconds,
			BackoffLimit:            &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"kloudlite.io/job-type":         "workspace-clone-sender",
						"kloudlite.io/target-workspace": workspace.Name,
						"kloudlite.io/source-workspace": sourceWorkspace.Name,
					},
				},
				Spec: corev1.PodSpec{
					// Schedule on the source WorkMachine node
					NodeName:      sourceWorkmachine.Name,
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    "sender",
							Image:   copierImage,
							Command: []string{"sh", "-c"},
							Args: []string{
								fmt.Sprintf(`
set -e
echo "Installing required packages..."
apk add --no-cache python3 tar gzip

echo "Creating compressed archive from workspace directory..."
cd %s
tar czf /tmp/workspace-data.tar.gz .

echo "Archive created successfully"
ls -lh /tmp/workspace-data.tar.gz

echo "Starting HTTP server to serve the archive..."
cd /tmp
python3 -m http.server %d
`, sourceFolderPath, senderHTTPPort),
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: senderHTTPPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1000m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workspace-data",
									MountPath: sourceFolderPath,
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "workspace-data",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: sourceFolderPath,
									Type: fn.Ptr(corev1.HostPathDirectory),
								},
							},
						},
					},
				},
			},
		},
	}

	// Set workspace as owner
	if err := controllerutil.SetControllerReference(workspace, job, c.Scheme()); err != nil {
		return nil, fmt.Errorf("failed to set owner reference: %w", err)
	}

	if err := c.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create sender job: %w", err)
	}

	logger.Info("Successfully created sender job",
		zap.String("jobName", jobName),
		zap.String("namespace", namespace))

	return job, nil
}

// createReceiverJob creates a job that downloads and extracts the workspace directory
// The job runs on the target WorkMachine node and mounts the target workspace directory
func (c *WorkspaceDirectoryCopier) createReceiverJob(
	ctx context.Context,
	workspace *workspacev1.Workspace,
	senderPodIP string,
	logger *zap.Logger,
) (*batchv1.Job, error) {
	jobName := fmt.Sprintf("ws-clone-receiver-%s", workspace.Name)
	namespace, err := c.getJobNamespace(workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to get job namespace: %w", err)
	}
	targetFolderPath := fmt.Sprintf("/var/lib/kloudlite/home/workspaces/%s", workspace.Name)

	logger.Info("Creating receiver job",
		zap.String("jobName", jobName),
		zap.String("namespace", namespace),
		zap.String("targetFolderPath", targetFolderPath),
		zap.String("senderPodIP", senderPodIP),
		zap.String("targetNode", workspace.Spec.WorkmachineName))

	// Get target WorkMachine to determine the node
	targetWorkmachine, err := c.getWorkMachine(ctx, workspace.Spec.WorkmachineName)
	if err != nil {
		return nil, fmt.Errorf("failed to get target WorkMachine: %w", err)
	}

	ttlSeconds := int32(copyJobTTLSeconds)
	backoffLimit := int32(3)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				"kloudlite.io/job-type":    "workspace-clone-receiver",
				"kloudlite.io/workspace":   workspace.Name,
				"kloudlite.io/workmachine": workspace.Spec.WorkmachineName,
			},
		},
		Spec: batchv1.JobSpec{
			TTLSecondsAfterFinished: &ttlSeconds,
			BackoffLimit:            &backoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"kloudlite.io/job-type":    "workspace-clone-receiver",
						"kloudlite.io/workspace":   workspace.Name,
						"kloudlite.io/workmachine": workspace.Spec.WorkmachineName,
					},
				},
				Spec: corev1.PodSpec{
					// Schedule on the target WorkMachine node
					NodeName:      targetWorkmachine.Name,
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:    "receiver",
							Image:   copierImage,
							Command: []string{"sh", "-c"},
							Args: []string{
								fmt.Sprintf(`
set -e
echo "Installing required packages..."
apk add --no-cache curl tar gzip

echo "Waiting for sender to be ready..."
MAX_RETRIES=60
RETRY_INTERVAL=2
for i in $(seq 1 $MAX_RETRIES); do
  if curl -f -s http://%s:%d/ > /dev/null 2>&1; then
    echo "Sender is ready"
    break
  fi
  echo "Sender not ready yet, retrying... ($i/$MAX_RETRIES)"
  sleep $RETRY_INTERVAL
done

echo "Downloading workspace archive from sender..."
curl -f --progress-bar -o /tmp/workspace-data.tar.gz http://%s:%d/workspace-data.tar.gz

echo "Download completed successfully"
ls -lh /tmp/workspace-data.tar.gz

echo "Creating target directory..."
mkdir -p %s

echo "Extracting archive to target directory..."
tar xzf /tmp/workspace-data.tar.gz -C %s

echo "Setting ownership to kl user (1001:1001)..."
chown -R 1001:1001 %s

echo "Extraction completed successfully"
ls -la %s
`, senderPodIP, senderHTTPPort, senderPodIP, senderHTTPPort, targetFolderPath, targetFolderPath, targetFolderPath, targetFolderPath),
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1000m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "workspace-data",
									MountPath: targetFolderPath,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "workspace-data",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: targetFolderPath,
									Type: fn.Ptr(corev1.HostPathDirectoryOrCreate),
								},
							},
						},
					},
				},
			},
		},
	}

	// Set workspace as owner
	if err := controllerutil.SetControllerReference(workspace, job, c.Scheme()); err != nil {
		return nil, fmt.Errorf("failed to set owner reference: %w", err)
	}

	if err := c.Create(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create receiver job: %w", err)
	}

	logger.Info("Successfully created receiver job",
		zap.String("jobName", jobName),
		zap.String("namespace", namespace))

	return job, nil
}

// waitForSenderPodReady waits for the sender pod to get an IP address
// Returns the pod IP when ready, or an error if timeout
func (c *WorkspaceDirectoryCopier) waitForSenderPodReady(
	ctx context.Context,
	senderJobName string,
	namespace string,
	logger *zap.Logger,
) (string, error) {
	logger.Info("Waiting for sender pod to be ready",
		zap.String("jobName", senderJobName),
		zap.String("namespace", namespace))

	timeout := 5 * time.Minute
	pollInterval := 2 * time.Second
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		// List pods with the sender job label
		podList := &corev1.PodList{}
		if err := c.List(ctx, podList, client.InNamespace(namespace), client.MatchingLabels{
			"job-name":              senderJobName,
			"kloudlite.io/job-type": "workspace-clone-sender",
		}); err != nil {
			logger.Warn("Failed to list sender pods", zap.Error(err))
			time.Sleep(pollInterval)
			continue
		}

		if len(podList.Items) == 0 {
			logger.Debug("No sender pods found yet")
			time.Sleep(pollInterval)
			continue
		}

		pod := &podList.Items[0]
		if pod.Status.Phase == corev1.PodRunning && pod.Status.PodIP != "" {
			logger.Info("Sender pod is ready",
				zap.String("podName", pod.Name),
				zap.String("podIP", pod.Status.PodIP))
			return pod.Status.PodIP, nil
		}

		logger.Debug("Sender pod not ready yet",
			zap.String("podName", pod.Name),
			zap.String("phase", string(pod.Status.Phase)),
			zap.String("podIP", pod.Status.PodIP))

		time.Sleep(pollInterval)
	}

	return "", fmt.Errorf("timeout waiting for sender pod to be ready after %v", timeout)
}

// getSenderPodIPIfReady immediately checks if the sender pod is ready and returns its IP
// Returns error if pod doesn't exist or isn't ready yet
// This method is used for recovery when the status wasn't updated but the pod is already running
func (c *WorkspaceDirectoryCopier) getSenderPodIPIfReady(
	ctx context.Context,
	senderJobName string,
	namespace string,
	logger *zap.Logger,
) (string, error) {
	// List pods with the sender job label
	podList := &corev1.PodList{}
	if err := c.List(ctx, podList, client.InNamespace(namespace), client.MatchingLabels{
		"job-name":              senderJobName,
		"kloudlite.io/job-type": "workspace-clone-sender",
	}); err != nil {
		return "", fmt.Errorf("failed to list sender pods: %w", err)
	}

	if len(podList.Items) == 0 {
		return "", fmt.Errorf("no sender pods found")
	}

	pod := &podList.Items[0]
	if pod.Status.Phase == corev1.PodRunning && pod.Status.PodIP != "" {
		logger.Info("Recovered sender pod IP from existing pod",
			zap.String("podName", pod.Name),
			zap.String("podIP", pod.Status.PodIP))
		return pod.Status.PodIP, nil
	}

	return "", fmt.Errorf("sender pod not ready yet (phase=%s, ip=%s)", pod.Status.Phase, pod.Status.PodIP)
}

// getDirectoryCopyStatus checks the status of the receiver job
// Returns (completed, failed, error)
func (c *WorkspaceDirectoryCopier) getDirectoryCopyStatus(
	ctx context.Context,
	receiverJobName string,
	namespace string,
	logger *zap.Logger,
) (bool, bool, error) {
	job := &batchv1.Job{}
	if err := c.Get(ctx, client.ObjectKey{
		Name:      receiverJobName,
		Namespace: namespace,
	}, job); err != nil {
		return false, false, fmt.Errorf("failed to get receiver job: %w", err)
	}

	completed := job.Status.Succeeded > 0
	failed := job.Status.Failed > 0

	if completed {
		logger.Info("Receiver job completed successfully",
			zap.String("jobName", receiverJobName))
	} else if failed {
		logger.Warn("Receiver job failed",
			zap.String("jobName", receiverJobName),
			zap.Int32("failedPods", job.Status.Failed))
	}

	return completed, failed, nil
}

// cleanupCopyJobs deletes the sender and receiver jobs
func (c *WorkspaceDirectoryCopier) cleanupCopyJobs(
	ctx context.Context,
	senderJobName string,
	receiverJobName string,
	namespace string,
	logger *zap.Logger,
) error {
	logger.Info("Cleaning up copy jobs",
		zap.String("senderJob", senderJobName),
		zap.String("receiverJob", receiverJobName),
		zap.String("namespace", namespace))

	// Delete sender job
	senderJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      senderJobName,
			Namespace: namespace,
		},
	}
	if err := c.Delete(ctx, senderJob, &client.DeleteOptions{
		PropagationPolicy: fn.Ptr(metav1.DeletePropagationBackground),
	}); err != nil && !apierrors.IsNotFound(err) {
		logger.Warn("Failed to delete sender job", zap.Error(err))
	}

	// Delete receiver job
	receiverJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      receiverJobName,
			Namespace: namespace,
		},
	}
	if err := c.Delete(ctx, receiverJob, &client.DeleteOptions{
		PropagationPolicy: fn.Ptr(metav1.DeletePropagationBackground),
	}); err != nil && !apierrors.IsNotFound(err) {
		logger.Warn("Failed to delete receiver job", zap.Error(err))
	}

	logger.Info("Successfully cleaned up copy jobs")
	return nil
}

// getJobNamespace returns the namespace where copy jobs should be created
// For workspaces, we use the WorkMachine's target namespace
func (c *WorkspaceDirectoryCopier) getJobNamespace(workspace *workspacev1.Workspace) (string, error) {
	// Get the WorkMachine to find its target namespace
	wm, err := c.getWorkMachine(context.TODO(), workspace.Spec.WorkmachineName)
	if err != nil {
		return "", fmt.Errorf("failed to get WorkMachine for namespace: %w", err)
	}
	return wm.Spec.TargetNamespace, nil
}

// getWorkMachine fetches the WorkMachine resource by name
func (c *WorkspaceDirectoryCopier) getWorkMachine(ctx context.Context, name string) (*workmachinev1.WorkMachine, error) {
	wm := &workmachinev1.WorkMachine{}
	if err := c.Get(ctx, client.ObjectKey{Name: name}, wm); err != nil {
		return nil, fmt.Errorf("failed to get WorkMachine %s: %w", name, err)
	}
	return wm, nil
}
