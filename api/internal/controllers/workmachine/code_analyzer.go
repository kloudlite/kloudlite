package workmachine

import (
	"fmt"

	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	codeAnalyzerName = "code-analyzer"
	codeAnalyzerPort = 8082
)

// ensureCodeAnalyzer ensures the code-analyzer StatefulSet exists
func (r *WorkMachineReconciler) ensureCodeAnalyzer(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace

	labels := map[string]string{
		"app":                      codeAnalyzerName,
		"kloudlite.io/workmachine": obj.Name,
	}

	// Create ServiceAccount
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      codeAnalyzerName,
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, serviceAccount, func() error {
		if !fn.IsOwner(serviceAccount, obj) {
			serviceAccount.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		serviceAccount.Labels = labels
		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update code-analyzer service account: %w", err))
	}

	// Create StatefulSet
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      codeAnalyzerName,
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
			ServiceName:         codeAnalyzerName,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName:            codeAnalyzerName,
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
					Containers: []corev1.Container{
						{
							Name:            codeAnalyzerName,
							Image:           r.env.CodeAnalyzerImage,
							ImagePullPolicy: corev1.PullAlways,
							Env: []corev1.EnvVar{
								{
									Name:  "NAMESPACE",
									Value: namespace,
								},
								{
									Name:  "WORKMACHINE_NAME",
									Value: obj.Name,
								},
								{
									Name:  "CLAUDE_API_URL",
									Value: "https://console.kloudlite.io/api/anthropic/v1/messages",
								},
								{
									Name:  "CLAUDE_API_KEY",
									Value: r.env.InstallationSecret,
								},
								{
									Name:  "WORKSPACES_PATH",
									Value: "/var/lib/kloudlite/home/workspaces",
								},
								{
									Name:  "REPORTS_PATH",
									Value: "/var/lib/kloudlite/code-analysis",
								},
								{
									Name:  "DEBOUNCE_SECONDS",
									Value: "45",
								},
								{
									Name:  "MAX_CONCURRENT_ANALYSES",
									Value: "2",
								},
								{
									Name:  "MAX_CONCURRENT_SCANS",
									Value: "3",
								},
								{
									Name:  "HTTP_PORT",
									Value: fmt.Sprintf("%d", codeAnalyzerPort),
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: codeAnalyzerPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(codeAnalyzerPort),
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/health",
										Port: intstr.FromInt(codeAnalyzerPort),
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       30,
								TimeoutSeconds:      10,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "kloudlite-data",
									MountPath: "/var/lib/kloudlite",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "kloudlite-data",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/var/lib/kloudlite",
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
		return check.Failed(fmt.Errorf("failed to create/update code-analyzer statefulset: %w", err))
	}

	// Create Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      codeAnalyzerName,
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, service, func() error {
		service.SetLabels(fn.MapMerge(service.GetLabels(), labels))

		if !fn.IsOwner(service, obj) {
			service.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		service.Spec = corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       codeAnalyzerPort,
					TargetPort: intstr.FromInt(codeAnalyzerPort),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		}
		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update code-analyzer service: %w", err))
	}

	return check.Passed()
}

// cleanupCodeAnalyzer removes the code-analyzer resources
func (r *WorkMachineReconciler) cleanupCodeAnalyzer(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace

	// Delete Service
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      codeAnalyzerName,
			Namespace: namespace,
		},
	}
	if err := r.Client.Delete(check.Context(), service); err != nil && !apiErrors.IsNotFound(err) {
		return check.Failed(fmt.Errorf("failed to delete code-analyzer service: %w", err))
	}

	// Delete StatefulSet
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      codeAnalyzerName,
			Namespace: namespace,
		},
	}
	if err := r.Client.Delete(check.Context(), statefulSet); err != nil && !apiErrors.IsNotFound(err) {
		return check.Failed(fmt.Errorf("failed to delete code-analyzer statefulset: %w", err))
	}

	// Delete ServiceAccount
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      codeAnalyzerName,
			Namespace: namespace,
		},
	}
	if err := r.Client.Delete(check.Context(), serviceAccount); err != nil && !apiErrors.IsNotFound(err) {
		return check.Failed(fmt.Errorf("failed to delete code-analyzer service account: %w", err))
	}

	return check.Passed()
}
