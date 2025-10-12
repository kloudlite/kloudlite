package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	environmentsv1 "github.com/kloudlite/kloudlite/api/pkg/apis/environments/v1"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// HelmChartHandlers handles HTTP requests for HelmChart resources
type HelmChartHandlers struct {
	helmChartRepo repository.HelmChartRepository
	k8sClient     client.Client
	logger        *zap.Logger
}

// NewHelmChartHandlers creates a new HelmChartHandlers
func NewHelmChartHandlers(helmChartRepo repository.HelmChartRepository, k8sClient client.Client, logger *zap.Logger) *HelmChartHandlers {
	return &HelmChartHandlers{
		helmChartRepo: helmChartRepo,
		k8sClient:     k8sClient,
		logger:        logger,
	}
}

// CreateHelmChart handles POST /api/v1/namespaces/:namespace/helmcharts
func (h *HelmChartHandlers) CreateHelmChart(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace is required",
		})
		return
	}

	var req struct {
		Name string                       `json:"name" binding:"required"`
		Spec environmentsv1.HelmChartSpec `json:"spec" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse create helm chart request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Create HelmChart object
	helmChart := &environmentsv1.HelmChart{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
		},
		Spec: req.Spec,
	}

	// Create the helm chart
	if err := h.helmChartRepo.Create(c.Request.Context(), helmChart); err != nil {
		h.logger.Error("Failed to create helm chart",
			zap.String("name", req.Name),
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create helm chart",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("HelmChart created successfully",
		zap.String("name", req.Name),
		zap.String("namespace", namespace))

	c.JSON(http.StatusCreated, gin.H{
		"message":   "HelmChart created successfully",
		"helmChart": helmChart,
	})
}

// GetHelmChart handles GET /api/v1/namespaces/:namespace/helmcharts/:name
func (h *HelmChartHandlers) GetHelmChart(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and helm chart name are required",
		})
		return
	}

	helmChart, err := h.helmChartRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		h.logger.Error("Failed to get helm chart",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("helmchart %s not found in namespace %s", name, namespace) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get helm chart",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, helmChart)
}

// ListHelmCharts handles GET /api/v1/namespaces/:namespace/helmcharts
func (h *HelmChartHandlers) ListHelmCharts(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace is required",
		})
		return
	}

	chartList, err := h.helmChartRepo.List(c.Request.Context(), namespace)
	if err != nil {
		h.logger.Error("Failed to list helm charts",
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list helm charts",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"helmCharts": chartList.Items,
		"count":      len(chartList.Items),
	})
}

// UpdateHelmChart handles PUT /api/v1/namespaces/:namespace/helmcharts/:name
func (h *HelmChartHandlers) UpdateHelmChart(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and helm chart name are required",
		})
		return
	}

	var req struct {
		Spec environmentsv1.HelmChartSpec `json:"spec" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update helm chart request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get existing helm chart
	helmChart, err := h.helmChartRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		h.logger.Error("Failed to get helm chart for update",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("helmchart %s not found in namespace %s", name, namespace) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get helm chart",
			"details": err.Error(),
		})
		return
	}

	// Update the spec
	helmChart.Spec = req.Spec

	// Update the helm chart
	if err := h.helmChartRepo.Update(c.Request.Context(), helmChart); err != nil {
		h.logger.Error("Failed to update helm chart",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update helm chart",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("HelmChart updated successfully",
		zap.String("name", name),
		zap.String("namespace", namespace))

	c.JSON(http.StatusOK, gin.H{
		"message":   "HelmChart updated successfully",
		"helmChart": helmChart,
	})
}

// DeleteHelmChart handles DELETE /api/v1/namespaces/:namespace/helmcharts/:name
func (h *HelmChartHandlers) DeleteHelmChart(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and helm chart name are required",
		})
		return
	}

	// Delete the helm chart
	if err := h.helmChartRepo.Delete(c.Request.Context(), namespace, name); err != nil {
		h.logger.Error("Failed to delete helm chart",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete helm chart",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("HelmChart deleted successfully",
		zap.String("name", name),
		zap.String("namespace", namespace))

	c.JSON(http.StatusOK, gin.H{
		"message":   "HelmChart deleted successfully",
		"name":      name,
		"namespace": namespace,
	})
}

// GetHelmChartStatus handles GET /api/v1/namespaces/:namespace/helmcharts/:name/status
func (h *HelmChartHandlers) GetHelmChartStatus(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and helm chart name are required",
		})
		return
	}

	helmChart, err := h.helmChartRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		h.logger.Error("Failed to get helm chart",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("helmchart %s not found in namespace %s", name, namespace) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get helm chart",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":                 helmChart.Name,
		"namespace":            helmChart.Namespace,
		"isReady":              helmChart.Status.IsReady,
		"checks":               helmChart.Status.Checks,
		"lastReconcileTime":    helmChart.Status.LastReconcileTime,
		"lastReadyGeneration":  helmChart.Status.LastReadyGeneration,
	})
}
