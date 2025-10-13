package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CompositionHandlers handles HTTP requests for Composition resources
type CompositionHandlers struct {
	compRepo  repository.CompositionRepository
	k8sClient client.Client
	logger    *zap.Logger
}

// NewCompositionHandlers creates a new CompositionHandlers
func NewCompositionHandlers(compRepo repository.CompositionRepository, k8sClient client.Client, logger *zap.Logger) *CompositionHandlers {
	return &CompositionHandlers{
		compRepo:  compRepo,
		k8sClient: k8sClient,
		logger:    logger,
	}
}

// CreateComposition handles POST /api/v1/namespaces/:namespace/compositions
func (h *CompositionHandlers) CreateComposition(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace is required",
		})
		return
	}

	var req struct {
		Name string                         `json:"name" binding:"required"`
		Spec environmentsv1.CompositionSpec `json:"spec" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse create composition request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Create Composition object
	composition := &environmentsv1.Composition{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
		},
		Spec: req.Spec,
	}

	// Create the composition
	if err := h.compRepo.Create(c.Request.Context(), composition); err != nil {
		h.logger.Error("Failed to create composition",
			zap.String("name", req.Name),
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create composition",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Composition created successfully",
		zap.String("name", req.Name),
		zap.String("namespace", namespace))

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Composition created successfully",
		"composition": composition,
	})
}

// GetComposition handles GET /api/v1/namespaces/:namespace/compositions/:name
func (h *CompositionHandlers) GetComposition(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and composition name are required",
		})
		return
	}

	composition, err := h.compRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		h.logger.Error("Failed to get composition",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("composition %s not found in namespace %s", name, namespace) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get composition",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, composition)
}

// ListCompositions handles GET /api/v1/namespaces/:namespace/compositions
func (h *CompositionHandlers) ListCompositions(c *gin.Context) {
	namespace := c.Param("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace is required",
		})
		return
	}

	// Parse query parameters for filtering
	state := c.Query("state")

	var compList *environmentsv1.CompositionList
	var err error

	// Handle state-based filtering
	if state != "" {
		compList, err = h.compRepo.ListByState(c.Request.Context(), namespace, environmentsv1.CompositionState(state))
	} else {
		compList, err = h.compRepo.List(c.Request.Context(), namespace)
	}

	if err != nil {
		h.logger.Error("Failed to list compositions",
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to list compositions",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"compositions": compList.Items,
		"count":        len(compList.Items),
	})
}

// UpdateComposition handles PUT /api/v1/namespaces/:namespace/compositions/:name
func (h *CompositionHandlers) UpdateComposition(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and composition name are required",
		})
		return
	}

	var req struct {
		Spec environmentsv1.CompositionSpec `json:"spec" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to parse update composition request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get existing composition
	composition, err := h.compRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		h.logger.Error("Failed to get composition for update",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("composition %s not found in namespace %s", name, namespace) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get composition",
			"details": err.Error(),
		})
		return
	}

	// Update the spec
	composition.Spec = req.Spec

	// Update the composition
	if err := h.compRepo.Update(c.Request.Context(), composition); err != nil {
		h.logger.Error("Failed to update composition",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update composition",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Composition updated successfully",
		zap.String("name", name),
		zap.String("namespace", namespace))

	c.JSON(http.StatusOK, gin.H{
		"message":     "Composition updated successfully",
		"composition": composition,
	})
}

// DeleteComposition handles DELETE /api/v1/namespaces/:namespace/compositions/:name
func (h *CompositionHandlers) DeleteComposition(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and composition name are required",
		})
		return
	}

	// Delete the composition
	if err := h.compRepo.Delete(c.Request.Context(), namespace, name); err != nil {
		h.logger.Error("Failed to delete composition",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete composition",
			"details": err.Error(),
		})
		return
	}

	h.logger.Info("Composition deleted successfully",
		zap.String("name", name),
		zap.String("namespace", namespace))

	c.JSON(http.StatusOK, gin.H{
		"message":   "Composition deleted successfully",
		"name":      name,
		"namespace": namespace,
	})
}

// GetCompositionStatus handles GET /api/v1/namespaces/:namespace/compositions/:name/status
func (h *CompositionHandlers) GetCompositionStatus(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	if namespace == "" || name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Namespace and composition name are required",
		})
		return
	}

	composition, err := h.compRepo.Get(c.Request.Context(), namespace, name)
	if err != nil {
		h.logger.Error("Failed to get composition",
			zap.String("name", name),
			zap.String("namespace", namespace),
			zap.Error(err))

		statusCode := http.StatusInternalServerError
		if err.Error() == fmt.Sprintf("composition %s not found in namespace %s", name, namespace) {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, gin.H{
			"error":   "Failed to get composition",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":             composition.Name,
		"namespace":        composition.Namespace,
		"state":            composition.Status.State,
		"message":          composition.Status.Message,
		"servicesCount":    composition.Status.ServicesCount,
		"runningCount":     composition.Status.RunningCount,
		"services":         composition.Status.Services,
		"endpoints":        composition.Status.Endpoints,
		"lastDeployedTime": composition.Status.LastDeployedTime,
	})
}
