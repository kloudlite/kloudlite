package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceInterceptHandlers handles HTTP requests for ServiceIntercept resources
type ServiceInterceptHandlers struct {
	k8sClient client.Client
	logger    *zap.Logger
}

// NewServiceInterceptHandlers creates a new ServiceInterceptHandlers
func NewServiceInterceptHandlers(k8sClient client.Client, logger *zap.Logger) *ServiceInterceptHandlers {
	return &ServiceInterceptHandlers{
		k8sClient: k8sClient,
		logger:    logger,
	}
}

// CreateServiceIntercept handles POST /api/v1/namespaces/:namespace/service-intercepts
func (h *ServiceInterceptHandlers) CreateServiceIntercept(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace is required",
		})
		return
	}

	var req struct {
		Name string                           `json:"name" binding:"required"`
		Spec interceptsv1.ServiceInterceptSpec `json:"spec" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse create service intercept request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Create ServiceIntercept object
	intercept := &interceptsv1.ServiceIntercept{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
		},
		Spec: req.Spec,
	}

	// Create the service intercept
	if err := h.k8sClient.Create(c.Request.Context(), intercept); err != nil {
		h.logger.Error("Failed to create service intercept",
			zap.String("name", req.Name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if apierrors.IsAlreadyExists(err) {
			statusCode = http.StatusConflict
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to create service intercept",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("ServiceIntercept created successfully",
		zap.String("name", req.Name),
		zap.String("namespace", namespace))

	c.JSON(http.StatusCreated, gin.H{
		"message":          "ServiceIntercept created successfully",
		"serviceIntercept": intercept,
	})
}

// GetServiceIntercept handles GET /api/v1/namespaces/:namespace/service-intercepts/:name
func (h *ServiceInterceptHandlers) GetServiceIntercept(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and service intercept name are required",
		})
		return
	}

	intercept := &interceptsv1.ServiceIntercept{}
	err := h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, intercept)

	if err != nil {
		h.logger.Error("Failed to get service intercept",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if apierrors.IsNotFound(err) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get service intercept",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, intercept)
}

// ListServiceIntercepts handles GET /api/v1/namespaces/:namespace/service-intercepts
func (h *ServiceInterceptHandlers) ListServiceIntercepts(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace is required",
		})
		return
	}

	interceptList := &interceptsv1.ServiceInterceptList{}
	err := h.k8sClient.List(c.Request.Context(), interceptList, client.InNamespace(namespace))

	if err != nil {
		h.logger.Error("Failed to list service intercepts",
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list service intercepts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"serviceIntercepts": interceptList.Items,
		"count":             len(interceptList.Items),
	})
}

// UpdateServiceIntercept handles PUT /api/v1/namespaces/:namespace/service-intercepts/:name
func (h *ServiceInterceptHandlers) UpdateServiceIntercept(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and service intercept name are required",
		})
		return
	}

	var req struct {
		Spec interceptsv1.ServiceInterceptSpec `json:"spec" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update service intercept request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get existing service intercept
	intercept := &interceptsv1.ServiceIntercept{}
	err := h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, intercept)

	if err != nil {
		h.logger.Error("Failed to get service intercept for update",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if apierrors.IsNotFound(err) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get service intercept",
			"details": err.Error(),
		})
		return
	}

	// Update the spec
	intercept.Spec = req.Spec

	// Update the service intercept
	if err := h.k8sClient.Update(c.Request.Context(), intercept); err != nil {
		h.logger.Error("Failed to update service intercept",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update service intercept",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("ServiceIntercept updated successfully",
		zap.String("name", name),
		zap.String("namespace", namespace))

	c.JSON(http.StatusOK, gin.H{
		"message":          "ServiceIntercept updated successfully",
		"serviceIntercept": intercept,
	})
}

// DeleteServiceIntercept handles DELETE /api/v1/namespaces/:namespace/service-intercepts/:name
func (h *ServiceInterceptHandlers) DeleteServiceIntercept(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and service intercept name are required",
		})
		return
	}

	intercept := &interceptsv1.ServiceIntercept{}
	err := h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, intercept)

	if err != nil {
		if apierrors.IsNotFound(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("ServiceIntercept '%s' not found", name),
			})
			return
		}

		h.logger.Error("Failed to get service intercept for deletion",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get service intercept",
			"details": err.Error(),
		})
		return
	}

	// Delete the service intercept
	if err := h.k8sClient.Delete(c.Request.Context(), intercept); err != nil {
		h.logger.Error("Failed to delete service intercept",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete service intercept",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("ServiceIntercept deleted successfully",
		zap.String("name", name),
		zap.String("namespace", namespace))

	c.JSON(http.StatusOK, gin.H{
		"message":   "ServiceIntercept deleted successfully",
		"name":      name,
		"namespace": namespace,
	})
}

// ActivateServiceIntercept handles POST /api/v1/namespaces/:namespace/service-intercepts/:name/activate
func (h *ServiceInterceptHandlers) ActivateServiceIntercept(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and service intercept name are required",
		})
		return
	}

	intercept := &interceptsv1.ServiceIntercept{}
	err := h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, intercept)

	if err != nil {
		h.logger.Error("Failed to get service intercept",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if apierrors.IsNotFound(err) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get service intercept",
			"details": err.Error(),
		})
		return
	}

	// Set status to active
	intercept.Spec.Status = "active"

	if err := h.k8sClient.Update(context.TODO(), intercept); err != nil {
		h.logger.Error("Failed to activate service intercept",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to activate service intercept",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("ServiceIntercept activated successfully",
		zap.String("name", name),
		zap.String("namespace", namespace))

	c.JSON(http.StatusOK, gin.H{
		"message":          "ServiceIntercept activated successfully",
		"serviceIntercept": intercept,
	})
}

// DeactivateServiceIntercept handles POST /api/v1/namespaces/:namespace/service-intercepts/:name/deactivate
func (h *ServiceInterceptHandlers) DeactivateServiceIntercept(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and service intercept name are required",
		})
		return
	}

	intercept := &interceptsv1.ServiceIntercept{}
	err := h.k8sClient.Get(c.Request.Context(), client.ObjectKey{
		Name:      name,
		Namespace: namespace,
	}, intercept)

	if err != nil {
		h.logger.Error("Failed to get service intercept",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if apierrors.IsNotFound(err) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get service intercept",
			"details": err.Error(),
		})
		return
	}

	// Set status to inactive
	intercept.Spec.Status = "inactive"

	if err := h.k8sClient.Update(context.TODO(), intercept); err != nil {
		h.logger.Error("Failed to deactivate service intercept",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to deactivate service intercept",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("ServiceIntercept deactivated successfully",
		zap.String("name", name),
		zap.String("namespace", namespace))

	c.JSON(http.StatusOK, gin.H{
		"message":          "ServiceIntercept deactivated successfully",
		"serviceIntercept": intercept,
	})
}
