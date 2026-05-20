package machinescoped

import (
	"context"
	"fmt"

	workmachineshared "github.com/kloudlite/kloudlite/controllers/workmachine/shared"

	fn "github.com/kloudlite/kloudlite/pkg/operator-toolkit/functions"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// ensureHostManagerPod ensures the workmachine-host-manager StatefulSet exists
// This function is called when the WorkMachine is in running state
func (r *MachineScopedReconciler) ensureHostManagerPod(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	namespace := obj.Spec.TargetNamespace
	hostManagerName := "host-manager"

	labels := map[string]string{
		"app":                       hostManagerName,
		"kloudlite.io/package-mgmt": "true",
		"kloudlite.io/workmachine":  obj.Name,
	}

	// Create StatefulSet
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostManagerName,
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, statefulSet, func() error {
		statefulSet.SetLabels(fn.MapMerge(statefulSet.GetLabels(), labels))

		if !fn.IsOwner(statefulSet, obj) {
			statefulSet.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		statefulSet.Spec = appsv1.StatefulSetSpec{
			Replicas:            fn.Ptr(int32(1)),
			ServiceName:         hostManagerName,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName:            "host-manager",
					TerminationGracePeriodSeconds: fn.Ptr(int64(5)),
					NodeSelector:                  workmachineshared.WorkMachineAddOnPlacement(obj.Name).NodeSelector,
					Tolerations:                   workmachineshared.WorkMachineAddOnPlacement(obj.Name).Tolerations,
					HostPID:                       true,
					InitContainers: []corev1.Container{
						{
							Name:            "setup-nix",
							Image:           r.env.HostManagerImage,
							ImagePullPolicy: corev1.PullAlways,
							Command: []string{
								"sh",
								"-c",
								"if [ -z \"$(ls -A /nix-shared)\" ]; then echo 'Nix store is empty, copying...'; cp -r /nix/* /nix-shared/; else echo 'Nix store already exists, skipping copy'; fi",
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: fn.Ptr(true),
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "nix-store",
									MountPath: "/nix-shared",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "host-manager",
							Image:           r.env.HostManagerImage,
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name:  "NAMESPACE",
									Value: obj.Spec.TargetNamespace,
								},
								{
									Name:  "WORKMACHINE_NAME",
									Value: obj.Name,
								},
								{
									Name:  "SNAPSHOT_REGISTRY_ENDPOINT",
									Value: r.env.SnapshotRegistryEndpoint,
								},
								{
									Name:  "SNAPSHOT_REGISTRY_PREFIX",
									Value: r.env.SnapshotRegistryPrefix,
								},
								{
									Name:  "SNAPSHOT_REGISTRY_INSECURE",
									Value: r.env.SnapshotRegistryInsecure,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: fn.Ptr(true),
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "metrics",
									ContainerPort: 8081,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "nix-store",
									MountPath: "/nix",
								},
								{
									Name:      "kloudlite-data",
									MountPath: "/var/lib/kloudlite",
								},
								{
									Name:      "host-sys",
									MountPath: "/host/sys",
									ReadOnly:  true,
								},
								{
									Name:      "host-dev",
									MountPath: "/host/dev",
									ReadOnly:  true,
								},
								{
									Name:      "host-proc",
									MountPath: "/host/proc",
									ReadOnly:  true,
								},
								{
									Name:      "host-lib-modules",
									MountPath: "/lib/modules",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "nix-store",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kloudlite/nix-store",
									Type: fn.Ptr(corev1.HostPathDirectoryOrCreate),
								},
							},
						},
						{
							Name: "kloudlite-data",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kloudlite",
									Type: fn.Ptr(corev1.HostPathDirectoryOrCreate),
								},
							},
						},
						{
							Name: "host-sys",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/sys",
									Type: fn.Ptr(corev1.HostPathDirectory),
								},
							},
						},
						{
							Name: "host-dev",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/dev",
									Type: fn.Ptr(corev1.HostPathDirectory),
								},
							},
						},
						{
							Name: "host-proc",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/proc",
									Type: fn.Ptr(corev1.HostPathDirectory),
								},
							},
						},
						{
							Name: "host-lib-modules",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/lib/modules",
									Type: fn.Ptr(corev1.HostPathDirectory),
								},
							},
						},
					},
				},
			},
		}

		return nil
	}); err != nil {
		return markMachineFailed(session, workmachineshared.ConditionHostManagerReady, fmt.Errorf("failed to create/update host-manager statefulset: %w", err))
	}

	// Create Service for SSH access
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostManagerName,
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.SetLabels(fn.MapMerge(svc.GetLabels(), labels))

		if !fn.IsOwner(svc, obj) {
			svc.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		svc.Spec.Selector = labels
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "metrics",
				Protocol:   corev1.ProtocolTCP,
				Port:       8081,
				TargetPort: intstr.FromInt32(8081),
			},
		}

		return nil
	}); err != nil {
		return markMachineFailed(session, workmachineshared.ConditionHostManagerReady, fmt.Errorf("failed to create/update host-manager service: %w", err))
	}

	return markMachineReady(session, workmachineshared.ConditionHostManagerReady, "host manager is ready")
}

// cleanupHostManagerPod deletes the host-manager StatefulSet and service
// This function is called when the WorkMachine is not in running state
func (r *MachineScopedReconciler) cleanupHostManagerPod(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	namespace := obj.Spec.TargetNamespace
	hostManagerName := "host-manager"

	// Delete StatefulSet if it exists (this will cascade delete pods)
	if err := r.Delete(ctx, &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostManagerName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionHostManagerReady, fmt.Errorf("failed to delete host-manager statefulset: %w", err))
		}
	}

	// Delete service
	if err := r.Delete(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostManagerName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionHostManagerReady, fmt.Errorf("failed to delete host-manager service: %w", err))
		}
	}

	return markMachineReady(session, workmachineshared.ConditionHostManagerReady, "host manager is cleaned up")
}
