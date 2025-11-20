package domainrequest

import (
	"context"
	"fmt"

	domainrequestsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/domainrequest/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// generateHAProxyConfig generates HAProxy configuration for routing traffic
func (r *DomainRequestReconciler) generateHAProxyConfig(domainRequest *domainrequestsv1.DomainRequest) string {
	config := `global
    maxconn 4096
    stats socket /var/run/haproxy.sock mode 660 level admin expose-fd listeners

defaults
    mode tcp
    timeout connect 5000ms
    timeout client 50000ms
    timeout server 50000ms
`

	// Use a single frontend on port 443 that handles both HTTPS and SSH
	if domainRequest.Spec.SSHBackend != nil {
		// Multiplex HTTPS and SSH on port 443 using protocol detection
		config += `
# Unified frontend on port 443 for both HTTPS and SSH traffic
frontend unified_frontend
    mode tcp
    bind *:443
    tcp-request inspect-delay 5s
    tcp-request content accept if { req.ssl_hello_type 1 } or { req.payload(0,7) -m str SSH-2.0 }

    # Detect SSH protocol (starts with "SSH-2.0")
    acl is_ssh req.payload(0,7) -m str SSH-2.0

    # Detect TLS/HTTPS traffic (TLS handshake)
    acl is_tls req.ssl_hello_type 1

    # Route SSH traffic to SSH backend, TLS traffic to HTTPS backend
    use_backend ssh_backend if is_ssh
    use_backend https_backend if is_tls

    # Default to HTTPS backend (shouldn't happen with proper detection)
    default_backend https_backend

# HTTPS backend - forwards to local SSL termination frontend
backend https_backend
    mode tcp
    server https_processor 127.0.0.1:8443

# Local HTTPS frontend for SSL termination and HTTP routing
frontend https_processor
    mode http
    bind 127.0.0.1:8443 ssl crt /etc/haproxy/certs/tls.pem
    timeout client 50000ms
    option forwardfor
    option http-server-close
`
	} else {
		// No SSH backend, use simple HTTPS frontend
		config += `
frontend https_frontend
    mode http
    bind *:443 ssl crt /etc/haproxy/certs/tls.pem
    option forwardfor
    option http-server-close
`
	}

	// Add domain-based routing to the appropriate frontend
	// Domain routing is added to the https_processor frontend when SSH is enabled,
	// or to the https_frontend when SSH is not enabled
	if len(domainRequest.Spec.DomainRoutes) > 0 {
		for i, route := range domainRequest.Spec.DomainRoutes {
			aclName := fmt.Sprintf("is_domain_%d", i)
			backendName := fmt.Sprintf("domain_backend_%d", i)
			config += fmt.Sprintf("\n    # Route for domain: %s\n", route.Domain)
			config += fmt.Sprintf("    acl %s hdr(host) -i %s\n", aclName, route.Domain)
			config += fmt.Sprintf("    use_backend %s if %s\n", backendName, aclName)
		}
		// Use first domain route as default backend
		config += "\n    default_backend domain_backend_0\n"
	} else if domainRequest.Spec.IngressBackend != nil {
		// Use IngressBackend as default if no domain routes
		config += "\n    default_backend service_backend\n"
	} else {
		// Fallback to default backend
		config += "\n    default_backend default_backend\n"
	}

	// Add backends for domain routes
	for i, route := range domainRequest.Spec.DomainRoutes {
		backendName := fmt.Sprintf("domain_backend_%d", i)
		config += fmt.Sprintf("\nbackend %s\n", backendName)
		config += "    mode http\n"
		config += fmt.Sprintf("    server backend%d %s.%s.svc.cluster.local:%d check\n",
			i, route.ServiceName, route.ServiceNamespace, route.ServicePort)
	}

	// Add IngressBackend if configured
	if domainRequest.Spec.IngressBackend != nil {
		backend := domainRequest.Spec.IngressBackend
		config += "\nbackend service_backend\n"
		config += "    mode http\n"
		config += fmt.Sprintf("    server backend1 %s.%s.svc.cluster.local:%d check\n",
			backend.ServiceName, backend.ServiceNamespace, backend.ServicePort)
	}

	// Add SSH backend if configured
	if domainRequest.Spec.SSHBackend != nil {
		backend := domainRequest.Spec.SSHBackend
		config += "\n# SSH Backend (TCP mode for raw SSH traffic)\n"
		config += "backend ssh_backend\n"
		config += "    mode tcp\n"
		config += "    timeout server 1h\n"
		config += "    timeout connect 10s\n"
		config += fmt.Sprintf("    server ssh1 %s.%s.svc.cluster.local:%d check\n",
			backend.ServiceName, backend.ServiceNamespace, backend.ServicePort)
	}

	// Add default backend if no other backends configured
	if len(domainRequest.Spec.DomainRoutes) == 0 && domainRequest.Spec.IngressBackend == nil {
		config += "\nbackend default_backend\n"
		config += "    mode http\n"
		config += "    server default1 frontend.kloudlite.svc.cluster.local:3000 check\n"
	}

	return config
}

// createHAProxyConfigMap creates or updates the HAProxy ConfigMap
func (r *DomainRequestReconciler) createHAProxyConfigMap(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) error {
	configMapName := fmt.Sprintf("%s-haproxy-config", domainRequest.Name)
	haproxyConfig := r.generateHAProxyConfig(domainRequest)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: domainRequest.Spec.WorkloadNamespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, configMap, func() error {
		if configMap.Data == nil {
			configMap.Data = make(map[string]string)
		}
		configMap.Data["haproxy.cfg"] = haproxyConfig

		// Set owner reference
		if err := controllerutil.SetControllerReference(domainRequest, configMap, r.Scheme); err != nil {
			return err
		}

		return nil
	})

	return err
}

// createHAProxyPod creates or updates the HAProxy pod with hostNetwork
func (r *DomainRequestReconciler) createHAProxyPod(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) error {
	// Check if origin certificate secret exists
	secretName := fmt.Sprintf("%s-origin-cert", domainRequest.Name)
	secret := &corev1.Secret{}
	if err := r.Get(ctx, client.ObjectKey{Name: secretName, Namespace: domainRequest.Spec.WorkloadNamespace}, secret); err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("origin certificate secret not yet created")
		}
		return fmt.Errorf("failed to check origin certificate secret: %w", err)
	}

	podName := fmt.Sprintf("%s-haproxy", domainRequest.Name)
	configMapName := fmt.Sprintf("%s-haproxy-config", domainRequest.Name)

	// Check if pod already exists
	existingPod := &corev1.Pod{}
	err := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: domainRequest.Spec.WorkloadNamespace}, existingPod)
	if err == nil {
		// Pod exists, delete it to force recreate (pods are immutable)
		logger.Info("Deleting existing HAProxy pod to recreate with new configuration")
		if err := r.Delete(ctx, existingPod); err != nil && !errors.IsNotFound(err) {
			return fmt.Errorf("failed to delete existing HAProxy pod: %w", err)
		}
		// Return and let the next reconcile create the new pod
		return nil
	} else if !errors.IsNotFound(err) {
		return fmt.Errorf("failed to check if HAProxy pod exists: %w", err)
	}

	// Create new pod
	// Single port 443 for both HTTPS and SSH (multiplexed)
	containerPorts := []corev1.ContainerPort{
		{
			Name:          "https",
			ContainerPort: 443,
			HostPort:      443,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	if domainRequest.Spec.SSHBackend != nil {
		logger.Info("SSH backend configured, SSH traffic will be multiplexed on port 443")
	}

	podSpec := corev1.PodSpec{
		HostNetwork: true,
		DNSPolicy:   corev1.DNSClusterFirstWithHostNet,
		Containers: []corev1.Container{
			{
				Name:  "haproxy",
				Image: "haproxy:2.8-alpine",
				SecurityContext: &corev1.SecurityContext{
					RunAsUser:  pointerInt64(0),
					RunAsGroup: pointerInt64(0),
					Capabilities: &corev1.Capabilities{
						Add: []corev1.Capability{
							"NET_BIND_SERVICE",
						},
					},
				},
				Ports: containerPorts,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "haproxy-config",
						MountPath: "/usr/local/etc/haproxy",
						ReadOnly:  true,
					},
					{
						Name:      "tls-certs",
						MountPath: "/etc/haproxy/certs",
						ReadOnly:  true,
					},
				},
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    *NewQuantity("100m"),
						corev1.ResourceMemory: *NewQuantity("128Mi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    *NewQuantity("500m"),
						corev1.ResourceMemory: *NewQuantity("256Mi"),
					},
				},
			},
		},
		Volumes: []corev1.Volume{
			{
				Name: "haproxy-config",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: configMapName,
						},
					},
				},
			},
			{
				Name: "tls-certs",
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: secretName,
					},
				},
			},
		},
	}

	// Use node selector and tolerations if NodeName is specified
	if domainRequest.Spec.NodeName != "" {
		podSpec.NodeSelector = map[string]string{
			"kubernetes.io/hostname": domainRequest.Spec.NodeName,
		}
		podSpec.Tolerations = []corev1.Toleration{
			{
				Key:      "kloudlite.io/workmachine",
				Operator: corev1.TolerationOpEqual,
				Value:    domainRequest.Spec.NodeName,
				Effect:   corev1.TaintEffectNoSchedule,
			},
		}
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: domainRequest.Spec.WorkloadNamespace,
			Labels: map[string]string{
				"app":                     "haproxy",
				"domain-request":          domainRequest.Name,
				"kloudlite.io/managed-by": "domainrequest-controller",
			},
		},
		Spec: podSpec,
	}

	// Set owner reference so pod is automatically cleaned up when DomainRequest is deleted
	if err := controllerutil.SetControllerReference(domainRequest, pod, r.Scheme); err != nil {
		return fmt.Errorf("failed to set owner reference: %w", err)
	}

	// Create the pod (cannot use CreateOrUpdate because pods are immutable)
	if err := r.Create(ctx, pod); err != nil {
		return fmt.Errorf("failed to create HAProxy pod: %w", err)
	}

	logger.Info("HAProxy pod created successfully", zap.String("podName", podName))
	return nil
}

// checkHAProxyReady checks if the HAProxy pod is ready
func (r *DomainRequestReconciler) checkHAProxyReady(ctx context.Context, podNamespace, podName string, logger *zap.Logger) (bool, error) {
	pod := &corev1.Pod{}
	err := r.Get(ctx, client.ObjectKey{Name: podName, Namespace: podNamespace}, pod)
	if err != nil {
		return false, err
	}

	// Check if pod is running and all containers are ready
	if pod.Status.Phase != corev1.PodRunning {
		return false, nil
	}

	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true, nil
		}
	}

	return false, nil
}

// deleteHAProxyResources deletes HAProxy pod and ConfigMap
func (r *DomainRequestReconciler) deleteHAProxyResources(ctx context.Context, domainRequest *domainrequestsv1.DomainRequest, logger *zap.Logger) error {
	podName := fmt.Sprintf("%s-haproxy", domainRequest.Name)
	configMapName := fmt.Sprintf("%s-haproxy-config", domainRequest.Name)

	// Delete pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: domainRequest.Spec.WorkloadNamespace,
		},
	}
	if err := r.Delete(ctx, pod); err != nil && !errors.IsNotFound(err) {
		logger.Error("Failed to delete HAProxy pod", zap.Error(err))
		return err
	}

	// Delete ConfigMap
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: domainRequest.Spec.WorkloadNamespace,
		},
	}
	if err := r.Delete(ctx, configMap); err != nil && !errors.IsNotFound(err) {
		logger.Error("Failed to delete HAProxy ConfigMap", zap.Error(err))
		return err
	}

	return nil
}

// Helper function to create resource quantities
func NewQuantity(value string) *resource.Quantity {
	q, _ := resource.ParseQuantity(value)
	return &q
}

// pointerInt64 returns a pointer to an int64 value
func pointerInt64(i int64) *int64 {
	return &i
}
