package workmachine

import (
	"fmt"

	domainrequestv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
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
	dockerDindName  = "docker-dind"
	dockerDindImage = "docker:dind"
	dockerDindPort  = 2375
)

// ensureBuildKit ensures the Docker dind StatefulSet exists for container image builds
func (r *WorkMachineReconciler) ensureBuildKit(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace

	labels := map[string]string{
		"app":                      dockerDindName,
		"kloudlite.io/workmachine": obj.Name,
	}

	// Fetch DomainRequest to get subdomain for image registry host
	var imageRegistryHost string
	domainRequest := &domainrequestv1.DomainRequest{}
	if err := r.Get(check.Context(), fn.NN("", "installation-domain"), domainRequest); err == nil && domainRequest.Status.Subdomain != "" {
		imageRegistryHost = fmt.Sprintf("cr.%s", domainRequest.Status.Subdomain)
	}

	// Get the wm-ingress-controller service ClusterIP for /etc/hosts entry
	var ingressControllerIP string
	ingressSvc := &corev1.Service{}
	if err := r.Get(check.Context(), fn.NN(namespace, "wm-ingress-controller"), ingressSvc); err == nil {
		ingressControllerIP = ingressSvc.Spec.ClusterIP
	}

	// Build HostAliases for docker-dind to resolve the image registry hostname
	var hostAliases []corev1.HostAlias
	if imageRegistryHost != "" && ingressControllerIP != "" {
		hostAliases = []corev1.HostAlias{
			{
				IP:        ingressControllerIP,
				Hostnames: []string{imageRegistryHost},
			},
		}
	}

	// Create StatefulSet for docker dind
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dockerDindName,
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
			ServiceName:         dockerDindName,
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
					HostAliases:                   hostAliases,
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
							Name:            dockerDindName,
							Image:           dockerDindImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{
									// Allow insecure connections (no TLS) for simplicity within cluster
									Name:  "DOCKER_TLS_CERTDIR",
									Value: "",
								},
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
									Name:          "docker",
									ContainerPort: dockerDindPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									Exec: &corev1.ExecAction{
										Command: []string{"docker", "info"},
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
										Command: []string{"docker", "info"},
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
									Name:      "docker-storage",
									MountPath: "/var/lib/docker",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "docker-storage",
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
		return check.Failed(fmt.Errorf("failed to create/update docker-dind statefulset: %w", err))
	}

	// Create Service for docker dind
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dockerDindName,
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
				Name:       "docker",
				Protocol:   corev1.ProtocolTCP,
				Port:       dockerDindPort,
				TargetPort: intstr.FromInt32(dockerDindPort),
			},
		}

		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update docker-dind service: %w", err))
	}

	// Cleanup old buildkitd resources if they exist
	r.cleanupOldBuildKit(check, obj)

	return check.Passed()
}

// cleanupOldBuildKit cleans up old buildkitd resources from the previous implementation
func (r *WorkMachineReconciler) cleanupOldBuildKit(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) {
	namespace := obj.Spec.TargetNamespace

	// Delete old buildkitd StatefulSet if it exists
	_ = r.Delete(check.Context(), &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "buildkitd",
			Namespace: namespace,
		},
	})

	// Delete old buildkitd service if it exists
	_ = r.Delete(check.Context(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "buildkitd",
			Namespace: namespace,
		},
	})
}

// cleanupBuildKit deletes the Docker dind StatefulSet and service
func (r *WorkMachineReconciler) cleanupBuildKit(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace

	// Delete StatefulSet if it exists
	if err := r.Delete(check.Context(), &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dockerDindName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete docker-dind statefulset: %w", err))
		}
	}

	// Delete service
	if err := r.Delete(check.Context(), &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      dockerDindName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete docker-dind service: %w", err))
		}
	}

	// Also cleanup old buildkitd resources
	r.cleanupOldBuildKit(check, obj)

	return check.Passed()
}
