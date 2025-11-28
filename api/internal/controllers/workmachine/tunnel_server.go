package workmachine

import (
	"fmt"

	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	tunnelServerName = "tunnel-server"

	// CA secret location (created by kloudlite-ca CertificateAuthority)
	kloudliteCASecretNamespace = "kloudlite-ingress"
	kloudliteCASecretName      = "kloudlite-ca"

	// Wildcard TLS certificate (signed by kloudlite CA, has *.domain SANs)
	kloudliteWildcardCertName = "kloudlite-wildcard-cert-tls"
)

// ensureTunnelServer ensures the tunnel-server StatefulSet exists for WireGuard connectivity
func (r *WorkMachineReconciler) ensureTunnelServer(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace
	clusterRoleName := fmt.Sprintf("%s-%s", tunnelServerName, obj.Name)
	clusterRoleBindingName := fmt.Sprintf("%s-%s", tunnelServerName, obj.Name)

	labels := map[string]string{
		"app":                      tunnelServerName,
		"kloudlite.io/workmachine": obj.Name,
	}

	// Create ServiceAccount
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tunnelServerName,
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
		return check.Failed(fmt.Errorf("failed to create/update tunnel-server service account: %w", err))
	}

	// Create ClusterRole
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, clusterRole, func() error {
		if !fn.IsOwner(clusterRole, obj) {
			clusterRole.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		clusterRole.Labels = labels
		clusterRole.Rules = []rbacv1.PolicyRule{
			{
				APIGroups: []string{"networking.k8s.io"},
				Resources: []string{"ingresses"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"namespaces"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"domains.kloudlite.io"},
				Resources: []string{"domainrequests"},
				Verbs:     []string{"get", "list", "watch"},
			},
		}
		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update tunnel-server cluster role: %w", err))
	}

	// Create ClusterRoleBinding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, clusterRoleBinding, func() error {
		if !fn.IsOwner(clusterRoleBinding, obj) {
			clusterRoleBinding.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		clusterRoleBinding.Labels = labels
		clusterRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
		}
		clusterRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      tunnelServerName,
				Namespace: namespace,
			},
		}
		return nil
	}); err != nil {
		return check.Failed(fmt.Errorf("failed to create/update tunnel-server cluster role binding: %w", err))
	}

	// Create TLS secrets for tunnel-server (signed by kloudlite CA)
	if err := r.ensureTunnelServerTLSSecrets(check, obj, namespace, labels); err != nil {
		return check.Failed(fmt.Errorf("failed to create tunnel-server TLS secrets: %w", err))
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
					ServiceAccountName:            tunnelServerName,
					TerminationGracePeriodSeconds: fn.Ptr(int64(10)),
					SecurityContext: &corev1.PodSecurityContext{
						Sysctls: []corev1.Sysctl{
							{
								Name:  "net.ipv4.ip_forward",
								Value: "1",
							},
						},
					},
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
								"--tls-secret", "tunnel-server-tls",
								"--ca-cert-secret", "tunnel-server-ca",
								"--wireguard-target", "127.0.0.1:51820",
								"--watch-config",
								"--config-path", "/etc/wireguard/wg0.conf",
								"--namespace", namespace,
								"--router-service", "wm-ingress-controller",
							},
							Env: []corev1.EnvVar{
								{
									Name:  "PUBLIC_HOST",
									Value: obj.Status.PublicIP,
								},
								{
									Name: "POD_NAMESPACE",
									ValueFrom: &corev1.EnvVarSource{
										FieldRef: &corev1.ObjectFieldSelector{
											FieldPath: "metadata.namespace",
										},
									},
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
									HostPort:      443,
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
									Name:      "wireguard-config",
									MountPath: "/etc/wireguard",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
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

// cleanupTunnelServer deletes the tunnel-server StatefulSet, service, and RBAC resources
func (r *WorkMachineReconciler) cleanupTunnelServer(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespace := obj.Spec.TargetNamespace
	clusterRoleName := fmt.Sprintf("%s-%s", tunnelServerName, obj.Name)
	clusterRoleBindingName := fmt.Sprintf("%s-%s", tunnelServerName, obj.Name)

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

	// Delete ClusterRoleBinding
	if err := r.Delete(check.Context(), &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete tunnel-server cluster role binding: %w", err))
		}
	}

	// Delete ClusterRole
	if err := r.Delete(check.Context(), &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete tunnel-server cluster role: %w", err))
		}
	}

	// Delete ServiceAccount
	if err := r.Delete(check.Context(), &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tunnelServerName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete tunnel-server service account: %w", err))
		}
	}

	return check.Passed()
}

// ensureTunnelServerTLSSecrets copies the wildcard TLS and CA secrets for tunnel-server
func (r *WorkMachineReconciler) ensureTunnelServerTLSSecrets(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine, namespace string, labels map[string]string) error {
	ctx := check.Context()

	// Fetch the kloudlite CA secret
	caSecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: kloudliteCASecretNamespace,
		Name:      kloudliteCASecretName,
	}, caSecret); err != nil {
		return fmt.Errorf("failed to get kloudlite CA secret: %w", err)
	}

	caCertPEM, ok := caSecret.Data["ca.crt"]
	if !ok {
		return fmt.Errorf("kloudlite CA secret missing ca.crt")
	}

	// Create the CA secret for tunnel-server (so clients can verify)
	tunnelCASecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tunnel-server-ca",
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, tunnelCASecret, func() error {
		tunnelCASecret.Labels = labels
		if !fn.IsOwner(tunnelCASecret, obj) {
			tunnelCASecret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		tunnelCASecret.Data = map[string][]byte{
			"ca.crt": caCertPEM,
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to create/update tunnel-server-ca secret: %w", err)
	}

	// Fetch the kloudlite wildcard TLS certificate (has *.domain SANs)
	wildcardCertSecret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{
		Namespace: kloudliteCASecretNamespace,
		Name:      kloudliteWildcardCertName,
	}, wildcardCertSecret); err != nil {
		return fmt.Errorf("failed to get kloudlite wildcard cert secret: %w", err)
	}

	tlsCert, ok := wildcardCertSecret.Data["tls.crt"]
	if !ok {
		return fmt.Errorf("kloudlite wildcard cert secret missing tls.crt")
	}

	tlsKey, ok := wildcardCertSecret.Data["tls.key"]
	if !ok {
		return fmt.Errorf("kloudlite wildcard cert secret missing tls.key")
	}

	// Copy the wildcard TLS certificate to tunnel-server namespace
	tlsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tunnel-server-tls",
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, tlsSecret, func() error {
		tlsSecret.Labels = labels
		if !fn.IsOwner(tlsSecret, obj) {
			tlsSecret.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		tlsSecret.Type = corev1.SecretTypeTLS
		tlsSecret.Data = map[string][]byte{
			"tls.crt": tlsCert,
			"tls.key": tlsKey,
			"ca.crt":  caCertPEM,
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to create/update tunnel-server-tls secret: %w", err)
	}

	return nil
}

