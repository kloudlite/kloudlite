package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace is required",
		})
		return
	}

	// List all services in the namespace
	serviceList := &corev1.ServiceList{}
	if err := h.k8sClient.List(context.Background(), serviceList, client.InNamespace(namespace)); err != nil {
		h.logger.Error("Failed to list services",
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list services",
			"details": err.Error(),
		})
		return
	}

	// Transform the services to a simpler format
	type ServicePort struct {
		Name       string `json:"name"`
		Protocol   string `json:"protocol"`
		Port       int32  `json:"port"`
		TargetPort string `json:"targetPort"`
	}

	type ServiceInfo struct {
		Name      string        `json:"name"`
		Namespace string        `json:"namespace"`
		Type      string        `json:"type"`
		ClusterIP string        `json:"clusterIP"`
		Ports     []ServicePort `json:"ports"`
		Selector  map[string]string `json:"selector,omitempty"`
	}

	services := make([]ServiceInfo, 0, len(serviceList.Items))
	for _, svc := range serviceList.Items {
		ports := make([]ServicePort, 0, len(svc.Spec.Ports))
		for _, port := range svc.Spec.Ports {
			ports = append(ports, ServicePort{
				Name:       port.Name,
				Protocol:   string(port.Protocol),
				Port:       port.Port,
				TargetPort: port.TargetPort.String(),
			})
		}

		services = append(services, ServiceInfo{
			Name:      svc.Name,
			Namespace: svc.Namespace,
			Type:      string(svc.Spec.Type),
			ClusterIP: svc.Spec.ClusterIP,
			Ports:     ports,
			Selector:  svc.Spec.Selector,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"count":    len(services),
	})
}
