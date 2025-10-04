package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/v2/api/internal/dto"
	"go.uber.org/zap"
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
func (h *ServiceHandlers) ListServices(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: "Namespace is required",
		})
		return
	}

	// List all services in the namespace
	serviceList := &corev1.ServiceList{}
	if err := h.k8sClient.List(c.Request.Context(), serviceList, client.InNamespace(namespace)); err != nil {
		h.logger.Error("Failed to list services",
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "Failed to list services",
			Details: err.Error(),
		})
		return
	}

	// Transform the services to a simpler format
	services := make([]dto.ServiceInfo, 0, len(serviceList.Items))
	for _, svc := range serviceList.Items {
		ports := make([]dto.ServicePort, 0, len(svc.Spec.Ports))
		for _, port := range svc.Spec.Ports {
			ports = append(ports, dto.ServicePort{
				Name:       port.Name,
				Protocol:   string(port.Protocol),
				Port:       port.Port,
				TargetPort: port.TargetPort.String(),
			})
		}

		services = append(services, dto.ServiceInfo{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Type:      string(svc.Spec.Type),
			ClusterIP: svc.Spec.ClusterIP,
			Ports:     ports,
			Selector:  svc.Spec.Selector,
		})
	}

	c.JSON(http.StatusOK, dto.ServiceListResponse{
		Services: services,
		Count:    len(services),
	})
}
