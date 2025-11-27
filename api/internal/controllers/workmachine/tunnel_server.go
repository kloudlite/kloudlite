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
	tunnelServerName = "tunnel-server"
)

// ensureTunnelServer ensures the tunnel-server StatefulSet exists for WireGuard connectivity
func (r *WorkMachineReconciler) ensureTunnelServer(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace

	labels := map[string]string{
		"app":                      tunnelServerName,
		"kloudlite.io/workmachine": obj.Name,
	}

	// Create StatefulSet for tunnel-server
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tunnelServerName,
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
			ServiceName:         tunnelServerName,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					TerminationGracePeriodSeconds: fn.Ptr(int64(10)),
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
							Name:            tunnelServerName,
							Image:           r.env.TunnelServerImage,
							ImagePullPolicy: corev1.PullAlways,
							Args: []string{
								"--listen", ":443",
								"--tls-cert", "/certs/tls.crt",
								"--tls-key", "/certs/tls.key",
								"--wireguard-target", "127.0.0.1:51820",
								"--watch-config",
								"--config-path", "/etc/wireguard/wg0.conf",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "PUBLIC_HOST",
									Value: obj.Status.PublicIP,
								},
							},
							SecurityContext: &corev1.SecurityContext{
								Capabilities: &corev1.Capabilities{
									Add: []corev1.Capability{
										"NET_ADMIN",
										"SYS_MODULE",
									},
								},
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("256Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("50m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "https",
									ContainerPort: 443,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "wireguard",
									ContainerPort: 51820,
									Protocol:      corev1.ProtocolUDP,
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/health",
										Port:   intstr.FromInt(443),
										Scheme: corev1.URISchemeHTTPS,
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
									HTTPGet: &corev1.HTTPGetAction{
										Path:   "/health",
										Port:   intstr.FromInt(443),
										Scheme: corev1.URISchemeHTTPS,
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds:       30,
								TimeoutSeconds:      5,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "tls-certs",
									MountPath: "/certs",
									ReadOnly:  true,
								},
								{
									Name:      "wireguard-config",
									MountPath: "/etc/wireguard",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "tls-certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: fmt.Sprintf("%s-tls", tunnelServerName),
								},
							},
						},
						{
							Name: "wireguard-config",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		}

		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update tunnel-server statefulset: %w", err))
	}

	// Create Service for tunnel-server
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tunnelServerName,
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
				Name:       "https",
				Protocol:   corev1.ProtocolTCP,
				Port:       443,
				TargetPort: intstr.FromInt32(443),
			},
			{
				Name:       "wireguard",
				Protocol:   corev1.ProtocolUDP,
				Port:       51820,
				TargetPort: intstr.FromInt32(51820),
			},
		}

		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update tunnel-server service: %w", err))
	}

	return check.Passed()
}

// cleanupTunnelServer deletes the tunnel-server StatefulSet and service
func (r *WorkMachineReconciler) cleanupTunnelServer(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace

	// Delete StatefulSet if it exists
	if err := r.Delete(check.Context(), &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tunnelServerName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete tunnel-server statefulset: %w", err))
		}
	}

	// Delete service
	if err := r.Delete(check.Context(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tunnelServerName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete tunnel-server service: %w", err))
		}
	}

	return check.Passed()
}
