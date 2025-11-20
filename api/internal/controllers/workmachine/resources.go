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

// ensureHostManagerPod ensures the workmachine-host-manager StatefulSet exists
// This function is called when the WorkMachine is in running state
func (r *WorkMachineReconciler) ensureHostManagerPod(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
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

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, statefulSet, func() error {
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
					HostPID: true,
					InitContainers: []corev1.Container{
						{
							Name:  "setup-nix",
							Image: r.env.HostManagerImage,
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
							Name:  "host-manager",
							Image: r.env.HostManagerImage,
							Env: []corev1.EnvVar{
								{
									Name:  "NAMESPACE",
									Value: obj.Spec.TargetNamespace,
								},
								{
									Name:  "WORKMACHINE_NAME",
									Value: obj.Name,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Privileged: fn.Ptr(true),
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("1000m"),
									corev1.ResourceMemory: resource.MustParse("1Gi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("100m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
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
		return check.Failed(fmt.Errorf("failed to create/update host-manager statefulset: %w", err))
	}

	// Create Service for SSH access
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostManagerName,
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, svc, func() error {
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
		return check.Failed(fmt.Errorf("failed to create/update host-manager service: %w", err))
	}

	return check.Passed()
}

// cleanupHostManagerPod deletes the host-manager StatefulSet and service
// This function is called when the WorkMachine is not in running state
func (r *WorkMachineReconciler) cleanupHostManagerPod(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace
	hostManagerName := "host-manager"

	// Delete StatefulSet if it exists (this will cascade delete pods)
	if err := r.Delete(check.Context(), &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostManagerName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete host-manager statefulset: %w", err))
		}
	}

	// Delete service
	if err := r.Delete(check.Context(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostManagerName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete host-manager service: %w", err))
		}
	}

	return check.Passed()
}
