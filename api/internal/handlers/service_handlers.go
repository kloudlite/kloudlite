package handlers

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/dto"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceHandlers handles HTTP requests for Kubernetes Service resources
type ServiceHandlers struct {
	k8sClient client.Client
	clientset *kubernetes.Clientset
	logger    *zap.Logger
}

// NewServiceHandlers creates a new ServiceHandlers
func NewServiceHandlers(k8sClient client.Client, clientset *kubernetes.Clientset, logger *zap.Logger) *ServiceHandlers {
	return &ServiceHandlers{
		k8sClient: k8sClient,
		clientset: clientset,
		logger:    logger,
	}
}

// ListServices handles GET /api/v1/namespaces/:namespace/services
// Lists deployments created from compositions as "services"
func (h *ServiceHandlers) ListServices(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Namespace is required",
		})
		return
	}

	// List all deployments managed by compositions
	deploymentList := &appsv1.DeploymentList{}
	if err := h.k8sClient.List(c.Request.Context(), deploymentList,
		client.InNamespace(namespace),
		client.MatchingLabels{"kloudlite.io/managed": "true"},
	); err != nil {
		h.logger.Error("Failed to list deployments",
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to list services",
			Details: err.Error(),
		})
		return
	}

	// List all services in the namespace to enrich deployment data
	serviceList := &corev1.ServiceList{}
	if err := h.k8sClient.List(c.Request.Context(), serviceList, client.InNamespace(namespace)); err != nil {
		h.logger.Warn("Failed to list k8s services, continuing without service data",
			zap.String("namespace", namespace),
			zap.Error(err))
	}

	// Create a map of services by name for quick lookup
	serviceMap := make(map[string]*corev1.Service)
	for i := range serviceList.Items {
		svc := &serviceList.Items[i]
		serviceMap[svc.Name] = svc
	}

	// Transform the deployments to service info format
	services := make([]dto.ServiceInfo, 0, len(deploymentList.Items))
	for _, deploy := range deploymentList.Items {
		// Get associated service if it exists
		svc, hasSvc := serviceMap[deploy.Name]

		var ports []dto.ServicePort
		var clusterIP string
		var svcType string

		if hasSvc {
			ports = make([]dto.ServicePort, 0, len(svc.Spec.Ports))
			for _, port := range svc.Spec.Ports {
				ports = append(ports, dto.ServicePort{
					Name:       port.Name,
					Protocol:   string(port.Protocol),
					Port:       port.Port,
					TargetPort: port.TargetPort.String(),
				})
			}
			clusterIP = svc.Spec.ClusterIP
			svcType = string(svc.Spec.Type)
		} else {
			// No service exists, extract ports from deployment container spec
			ports = make([]dto.ServicePort, 0)
			if len(deploy.Spec.Template.Spec.Containers) > 0 {
				for _, port := range deploy.Spec.Template.Spec.Containers[0].Ports {
					ports = append(ports, dto.ServicePort{
						Name:       port.Name,
						Protocol:   string(port.Protocol),
						Port:       port.ContainerPort,
						TargetPort: string(port.ContainerPort),
					})
				}
			}
			svcType = "None"
		}

		services = append(services, dto.ServiceInfo{
			Name:      deploy.Name,
			Namespace: deploy.Namespace,
			Type:      svcType,
			ClusterIP: clusterIP,
			Ports:     ports,
			Selector:  deploy.Spec.Selector.MatchLabels,
			Replicas:  deploy.Status.ReadyReplicas,
			Image:     getImageFromDeployment(&deploy),
		})
	}

	c.JSON(http.StatusOK, dto.ServiceListResponse{
		Services: services,
		Count:    len(services),
	})
}

// getImageFromDeployment extracts the image from the first container
func getImageFromDeployment(deploy *appsv1.Deployment) string {
	if len(deploy.Spec.Template.Spec.Containers) > 0 {
		return deploy.Spec.Template.Spec.Containers[0].Image
	}
	return ""
}

// GetServiceLogs handles GET /api/v1/namespaces/:namespace/services/:name/logs
// Streams logs from pods belonging to the service deployment via HTTP streaming
// Query parameters:
//   - follow: boolean, whether to follow logs (default: false)
//   - tailLines: int, number of lines to show from the end (default: 100)
//   - container: string, container name (optional, uses first container if not specified)
//   - timestamps: boolean, include timestamps (default: false)
func (h *ServiceHandlers) GetServiceLogs(c *gin.Context) {
	ctx := c.Request.Context()

	namespace := c.Param("namespace")
	serviceName := c.Param("name")

	if namespace == "" || serviceName == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Namespace and service name are required",
		})
		return
	}

	// Parse query parameters
	follow := c.Query("follow") == "true"
	timestamps := c.Query("timestamps") == "true"
	container := c.Query("container")

	tailLines := int64(100)
	if t := c.Query("tailLines"); t != "" {
		if parsed, err := strconv.ParseInt(t, 10, 64); err == nil && parsed > 0 {
			tailLines = parsed
		}
	}

	// Find the deployment for this service
	deployment := &appsv1.Deployment{}
	if err := h.k8sClient.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      serviceName,
	}, deployment); err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Error: fmt.Sprintf("Service deployment '%s' not found in namespace '%s'", serviceName, namespace),
			})
			return
		}
		h.logger.Error("Failed to get deployment",
			zap.String("namespace", namespace),
			zap.String("service", serviceName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get service deployment",
			Details: err.Error(),
		})
		return
	}

	// Find pods for this deployment
	podList := &corev1.PodList{}
	if err := h.k8sClient.List(ctx, podList,
		client.InNamespace(namespace),
		client.MatchingLabels(deployment.Spec.Selector.MatchLabels),
	); err != nil {
		h.logger.Error("Failed to list pods",
			zap.String("namespace", namespace),
			zap.String("service", serviceName),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to list pods",
			Details: err.Error(),
		})
		return
	}

	if len(podList.Items) == 0 {
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error: fmt.Sprintf("No pods found for service '%s'", serviceName),
		})
		return
	}

	// Use the first running pod, or fallback to first pod
	var targetPod *corev1.Pod
	for i := range podList.Items {
		pod := &podList.Items[i]
		if pod.Status.Phase == corev1.PodRunning {
			targetPod = pod
			break
		}
	}
	if targetPod == nil {
		targetPod = &podList.Items[0]
	}

	// Determine container name
	if container == "" && len(targetPod.Spec.Containers) > 0 {
		container = targetPod.Spec.Containers[0].Name
	}

	// Build log options
	logOptions := &corev1.PodLogOptions{
		Container:  container,
		Follow:     follow,
		TailLines:  &tailLines,
		Timestamps: timestamps,
	}

	// Get log stream from Kubernetes
	req := h.clientset.CoreV1().Pods(namespace).GetLogs(targetPod.Name, logOptions)
	stream, err := req.Stream(ctx)
	if err != nil {
		h.logger.Error("Failed to get pod logs stream",
			zap.String("namespace", namespace),
			zap.String("pod", targetPod.Name),
			zap.String("container", container),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to get logs stream",
			Details: err.Error(),
		})
		return
	}
	defer stream.Close()

	// Set streaming headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Header("X-Pod-Name", targetPod.Name)
	c.Header("X-Container-Name", container)

	// Stream logs line by line as SSE events
	scanner := bufio.NewScanner(stream)
	// Increase buffer size for potentially long log lines
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			line := scanner.Text()
			// Write as SSE event
			fmt.Fprintf(c.Writer, "data: %s\n\n", line)
			c.Writer.Flush()
		}
	}

	if err := scanner.Err(); err != nil {
		h.logger.Debug("Log stream ended",
			zap.String("pod", targetPod.Name),
			zap.Error(err))
	}
}
