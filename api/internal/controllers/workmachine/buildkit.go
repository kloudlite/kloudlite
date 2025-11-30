package workmachine

import (
	"fmt"

	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	buildkitName  = "buildkitd"
	buildkitImage = "moby/buildkit:latest"
	buildkitPort  = 1234
)

// ensureBuildKit ensures the BuildKit StatefulSet exists for container image builds
func (r *WorkMachineReconciler) ensureBuildKit(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace

	labels := map[string]string{
		"app":                      buildkitName,
		"kloudlite.io/workmachine": obj.Name,
	}

	// Create StatefulSet for buildkitd
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildkitName,
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, statefulSet, func() error {
		statefulSet.SetLabels(fn.MapMerge(statefulSet.GetLabels(), labels))

		if !fn.IsOwner(statefulSet, obj) {
			statefulSet.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		statefulSet.Spec = appsv1.StatefulSetSpec{
			Replicas:            fn.Ptr(int32(1)),
			ServiceName:         buildkitName,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: fn.Ptr(int64(30)),
					NodeSelector: map[string]string{
						"kloudlite.io/workmachine": obj.Name,
					},
					Tolerations: []corev1.Toleration{
						{
							Key:      "kloudlite.io/workmachine",
							Operator: corev1.TolerationOpExists,
							Effect:   corev1.TaintEffectNoSchedule,
						},
						{
							Key:               "node.kubernetes.io/not-ready",
							Operator:          corev1.TolerationOpExists,
							Effect:            corev1.TaintEffectNoExecute,
							TolerationSeconds: fn.Ptr(int64(0)),
						},
						{
							Key:               "node.kubernetes.io/unreachable",
							Operator:          corev1.TolerationOpExists,
							Effect:            corev1.TaintEffectNoExecute,
							TolerationSeconds: fn.Ptr(int64(0)),
						},
					},
					// Init container to clean up stale lock files from previous runs
					InitContainers: []corev1.Container{
						{
							Name:  "cleanup-lock",
							Image: "busybox:latest",
							Command: []string{
								"sh", "-c",
								"rm -f /var/lib/buildkit/buildkitd.lock",
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "buildkit-cache",
									MountPath: "/var/lib/buildkit",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            buildkitName,
							Image:           buildkitImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Args: []string{
								"--root", "/var/lib/buildkit",
								"--addr", fmt.Sprintf("tcp://0.0.0.0:%d", buildkitPort),
								"--addr", "unix:///run/buildkit/buildkitd.sock",
								"--oci-worker-no-process-sandbox",
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: fn.Ptr(true),
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("4"),
									corev1.ResourceMemory: resource.MustParse("8Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "buildkit",
									ContainerPort: buildkitPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"buildctl", "debug", "workers"},
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"buildctl", "debug", "workers"},
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       30,
								TimeoutSeconds:      5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "buildkit-cache",
									MountPath: "/var/lib/buildkit",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "buildkit-cache",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: fmt.Sprintf("/var/lib/kloudlite/buildkit/%s", obj.Name),
									Type: fn.Ptr(corev1.HostPathDirectoryOrCreate),
								},
							},
						},
					},
				},
			},
		}

		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update buildkitd statefulset: %w", err))
	}

	// Create Service for buildkitd
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildkitName,
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, svc, func() error {
		svc.SetLabels(fn.MapMerge(svc.GetLabels(), labels))

		if !fn.IsOwner(svc, obj) {
			svc.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		svc.Spec.Type = corev1.ServiceTypeClusterIP
		svc.Spec.Selector = labels
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "buildkit",
				Protocol:   corev1.ProtocolTCP,
				Port:       buildkitPort,
				TargetPort: intstr.FromInt32(buildkitPort),
			},
		}

		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update buildkitd service: %w", err))
	}

	return check.Passed()
}

// cleanupBuildKit deletes the BuildKit StatefulSet and service
func (r *WorkMachineReconciler) cleanupBuildKit(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace

	// Delete StatefulSet if it exists
	if err := r.Delete(check.Context(), &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildkitName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete buildkitd statefulset: %w", err))
		}
	}

	// Delete service
	if err := r.Delete(check.Context(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildkitName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete buildkitd service: %w", err))
		}
	}

	return check.Passed()
}
