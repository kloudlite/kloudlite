package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/dto"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceHandlers handles HTTP requests for Kubernetes Service resources
type ServiceHandlers struct {
	k8sClient client.Client
	logger    *zap.Logger
}

// NewServiceHandlers creates a new ServiceHandlers
func NewServiceHandlers(k8sClient client.Client, logger *zap.Logger) *ServiceHandlers {
	return &ServiceHandlers{
		k8sClient: k8sClient,
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
