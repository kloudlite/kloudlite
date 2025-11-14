package environment

import (
	"context"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// PVCCopier handles copying data between PVCs across nodes with compression
type PVCCopier struct {
	client          client.Client
	sourceNamespace string
	targetNamespace string
}

// NewPVCCopier creates a new PVC copier
func NewPVCCopier(client client.Client, sourceNamespace, targetNamespace string) *PVCCopier {
	return &PVCCopier{
		client:          client,
		sourceNamespace: sourceNamespace,
		targetNamespace: targetNamespace,
	}
}

// CopyPVC copies data from source PVC to target PVC with compression
func (c *PVCCopier) CopyPVC(ctx context.Context, sourcePVC, targetPVC string, owner metav1.Object) error {
	// Get source PVC to extract node binding information
	pvc := &corev1.PersistentVolumeClaim{}
	if err := c.client.Get(ctx, types.NamespacedName{
		Name:      sourcePVC,
		Namespace: c.sourceNamespace,
	}, pvc); err != nil {
		return fmt.Errorf("failed to get source PVC: %w", err)
	}

	// Get bound node name from PV
	var boundNodeName string
	if pvc.Spec.VolumeName != "" {
		pv := &corev1.PersistentVolume{}
		if err := c.client.Get(ctx, types.NamespacedName{Name: pvc.Spec.VolumeName}, pv); err == nil {
			// Extract node from PV node affinity
			if pv.Spec.NodeAffinity != nil && pv.Spec.NodeAffinity.Required != nil {
				for _, term := range pv.Spec.NodeAffinity.Required.NodeSelectorTerms {
					for _, expr := range term.MatchExpressions {
						if expr.Key == "kubernetes.io/hostname" && len(expr.Values) > 0 {
							boundNodeName = expr.Values[0]
							break
						}
					}
					if boundNodeName != "" {
						break
					}
				}
			}
		}
	}

	// Create source sender job with node affinity
	senderJob := c.createSenderJob(sourcePVC, targetPVC, boundNodeName, owner)
	if err := controllerutil.SetControllerReference(owner, senderJob, c.client.Scheme()); err != nil {
		return fmt.Errorf("failed to set owner reference for sender job: %w", err)
	}
	if err := c.client.Create(ctx, senderJob); err != nil {
		return fmt.Errorf("failed to create sender job: %w", err)
	}

	// Wait for sender pod to be ready
	senderPodIP, err := c.waitForSenderReady(ctx, senderJob.Name)
	if err != nil {
		return fmt.Errorf("sender pod failed to become ready: %w", err)
	}

	// Create target receiver job with sender IP
	receiverJob := c.createReceiverJob(sourcePVC, targetPVC, senderPodIP, owner)
	if err := controllerutil.SetControllerReference(owner, receiverJob, c.client.Scheme()); err != nil {
		return fmt.Errorf("failed to set owner reference for receiver job: %w", err)
	}
	if err := c.client.Create(ctx, receiverJob); err != nil {
		return fmt.Errorf("failed to create receiver job: %w", err)
	}

	return nil
}

// createSenderJob creates a job that compresses and serves PVC data via HTTP
func (c *PVCCopier) createSenderJob(sourcePVC, targetPVC, boundNodeName string, owner metav1.Object) *batchv1.Job {
	jobName := fmt.Sprintf("pvc-copy-sender-%s", targetPVC)

	podSpec := corev1.PodSpec{
		RestartPolicy: corev1.RestartPolicyOnFailure,
		Containers: []corev1.Container{
			{
				Name:    "sender",
				Image:   "alpine:latest",
				Command: []string{"/bin/sh", "-c"},
				Args: []string{
					`
# Install required tools
apk add --no-cache python3 tar gzip

# Create compressed archive (source must be fully suspended before this runs)
echo "Creating compressed archive of source PVC..."
cd /source-data
tar czf /tmp/data.tar.gz . 2>/dev/null || tar czf /tmp/data.tar.gz --warning=no-file-changed .

# Get archive size for progress tracking
ARCHIVE_SIZE=$(stat -c%s /tmp/data.tar.gz)
echo "Archive created successfully. Size: $ARCHIVE_SIZE bytes"

# Start HTTP server to serve the archive
echo "Starting HTTP server on port 8080..."
cd /tmp
python3 -m http.server 8080
`,
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "source-volume",
						MountPath: "/source-data",
						ReadOnly:  true,
					},
				},
				Ports: []corev1.ContainerPort{
					{
						ContainerPort: 8080,
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
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "source-volume",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: sourcePVC,
						ReadOnly:  true,
					},
				},
			},
		},
	}

	// Add node selector if bound node is known
	if boundNodeName != "" {
		podSpec.NodeSelector = map[string]string{
			"kubernetes.io/hostname": boundNodeName,
		}
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: c.sourceNamespace,
			Labels: map[string]string{
				"app":                     "pvc-copier",
				"role":                    "sender",
				"kloudlite.io/source-pvc": sourcePVC,
				"kloudlite.io/target-pvc": targetPVC,
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            ptr(int32(3)),
			TTLSecondsAfterFinished: ptr(int32(300)), // Clean up after 5 minutes
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                     "pvc-copier",
						"role":                    "sender",
						"kloudlite.io/source-pvc": sourcePVC,
						"kloudlite.io/target-pvc": targetPVC,
					},
				},
				Spec: podSpec,
			},
		},
	}
}

// createReceiverJob creates a job that downloads and extracts PVC data
func (c *PVCCopier) createReceiverJob(sourcePVC, targetPVC, senderIP string, owner metav1.Object) *batchv1.Job {
	jobName := fmt.Sprintf("pvc-copy-receiver-%s", targetPVC)
	senderURL := fmt.Sprintf("http://%s:8080/data.tar.gz", senderIP)

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: c.targetNamespace,
			Labels: map[string]string{
				"app":                     "pvc-copier",
				"role":                    "receiver",
				"kloudlite.io/source-pvc": sourcePVC,
				"kloudlite.io/target-pvc": targetPVC,
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            ptr(int32(3)),
			TTLSecondsAfterFinished: ptr(int32(300)), // Clean up after 5 minutes
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":                     "pvc-copier",
						"role":                    "receiver",
						"kloudlite.io/source-pvc": sourcePVC,
						"kloudlite.io/target-pvc": targetPVC,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyOnFailure,
					Containers: []corev1.Container{
						{
							Name:    "receiver",
							Image:   "alpine:latest",
							Command: []string{"/bin/sh", "-c"},
							Args: []string{
								fmt.Sprintf(`
# Install required tools
apk add --no-cache curl tar gzip

# Wait for sender to be ready with retries
echo "Waiting for sender to be ready at %s..."
MAX_RETRIES=60
RETRY_COUNT=0
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -sf --head %s > /dev/null 2>&1; then
        echo "Sender is ready!"
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo "Attempt $RETRY_COUNT/$MAX_RETRIES: Sender not ready yet, waiting..."
    sleep 2
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    echo "ERROR: Sender did not become ready within timeout"
    exit 1
fi

# Download compressed archive with progress
echo "Downloading archive from sender..."
curl -f --progress-bar -o /tmp/data.tar.gz %s
DOWNLOAD_SIZE=$(stat -c%%s /tmp/data.tar.gz)
echo "Downloaded $DOWNLOAD_SIZE bytes"

# Extract to target PVC
echo "Extracting archive to target volume..."
cd /target-data
tar xzf /tmp/data.tar.gz

# Verify extraction
echo "Verifying extraction..."
FILE_COUNT=$(find /target-data -type f | wc -l)
DIR_SIZE=$(du -sh /target-data | cut -f1)
echo "Extraction complete: $FILE_COUNT files, $DIR_SIZE total"

echo "Copy completed successfully"
`, senderURL, senderURL, senderURL),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "target-volume",
									MountPath: "/target-data",
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
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "target-volume",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: targetPVC,
								},
							},
						},
					},
				},
			},
		},
	}
}

// waitForSenderReady waits for the sender pod to be running and returns its IP
func (c *PVCCopier) waitForSenderReady(ctx context.Context, jobName string) (string, error) {
	timeout := time.After(5 * time.Minute)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return "", fmt.Errorf("timeout waiting for sender pod to be ready")
		case <-ticker.C:
			// Get pods for the sender job
			podList := &corev1.PodList{}
			if err := c.client.List(ctx, podList,
				client.InNamespace(c.sourceNamespace),
				client.MatchingLabels{
					"job-name": jobName,
					"role":     "sender",
				}); err != nil {
				continue
			}

			// Check if any pod is running and has an IP
			for _, pod := range podList.Items {
				if pod.Status.Phase == corev1.PodRunning && pod.Status.PodIP != "" {
					return pod.Status.PodIP, nil
				}
			}
		}
	}
}

// GetCopyStatus checks the status of an ongoing copy operation
func (c *PVCCopier) GetCopyStatus(ctx context.Context, targetPVC string) (completed bool, failed bool, err error) {
	// Check receiver job status
	receiverJobName := fmt.Sprintf("pvc-copy-receiver-%s", targetPVC)
	receiverJob := &batchv1.Job{}
	if err := c.client.Get(ctx, types.NamespacedName{
		Name:      receiverJobName,
		Namespace: c.targetNamespace,
	}, receiverJob); err != nil {
		return false, false, err
	}

	// Check if job completed successfully
	if receiverJob.Status.Succeeded > 0 {
		return true, false, nil
	}

	// Check if job failed
	if receiverJob.Status.Failed > 0 {
		return false, true, fmt.Errorf("copy job failed after %d attempts", receiverJob.Status.Failed)
	}

	return false, false, nil
}

// GetBytesTransferred estimates bytes transferred by checking receiver pod logs
func (c *PVCCopier) GetBytesTransferred(ctx context.Context, targetPVC string) (int64, error) {
	// This is a placeholder - in a real implementation, you would:
	// 1. Read receiver pod logs
	// 2. Parse the DOWNLOAD_SIZE from the logs
	// 3. Return the actual bytes transferred
	// For now, we return 0 to indicate progress tracking is in progress
	return 0, nil
}

// CleanupCopyJobs removes copy jobs for a specific PVC
func (c *PVCCopier) CleanupCopyJobs(ctx context.Context, targetPVC string) error {
	// Delete sender job from source namespace
	senderJobName := fmt.Sprintf("pvc-copy-sender-%s", targetPVC)
	senderJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      senderJobName,
			Namespace: c.sourceNamespace,
		},
	}
	if err := c.client.Delete(ctx, senderJob, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("failed to delete sender job: %w", err)
		}
	}

	// Delete receiver job from target namespace
	receiverJobName := fmt.Sprintf("pvc-copy-receiver-%s", targetPVC)
	receiverJob := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      receiverJobName,
			Namespace: c.targetNamespace,
		},
	}
	if err := c.client.Delete(ctx, receiverJob, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return fmt.Errorf("failed to delete receiver job: %w", err)
		}
	}

	return nil
}

// Helper function to create pointer to int32
func ptr[T any](v T) *T {
	return &v
}
