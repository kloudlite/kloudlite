package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	machinesv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/machines/v1"
	"github.com/kloudlite/kloudlite/v2/api/internal/managers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MachineTypeHandlers handles HTTP requests for MachineType resources
type MachineTypeHandlers struct {
	manager *managers.Manager
}

// NewMachineTypeHandlers creates a new MachineType handler
func NewMachineTypeHandlers(manager *managers.Manager) *MachineTypeHandlers {
	return &MachineTypeHandlers{
		manager: manager,
	}
}

// MachineTypeCreateRequest represents a request to create a MachineType
type MachineTypeCreateRequest struct {
	Name string                  `json:"name" binding:"required"`
	Spec machinesv1.MachineTypeSpec `json:"spec" binding:"required"`
}

// MachineTypeUpdateRequest represents a request to update a MachineType
type MachineTypeUpdateRequest struct {
	Spec machinesv1.MachineTypeSpec `json:"spec" binding:"required"`
}

// ListMachineTypes returns all machine types
func (h *MachineTypeHandlers) ListMachineTypes(c *gin.Context) {
	ctx := c.Request.Context()

	// Check for query parameters
	activeOnly := c.Query("active") == "true"
	category := c.Query("category")

	var list *machinesv1.MachineTypeList
	var err error

	if activeOnly {
		list, err = h.manager.MachineTypeRepository.ListActive(ctx)
	} else if category != "" {
		list, err = h.manager.MachineTypeRepository.GetByCategory(ctx, category)
	} else {
		list, err = h.manager.MachineTypeRepository.List(ctx)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": list.Items,
		"count": len(list.Items),
	})
}

// GetMachineType returns a specific machine type
func (h *MachineTypeHandlers) GetMachineType(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	machineType, err := h.manager.MachineTypeRepository.Get(ctx, name)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Machine type not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, machineType)
}

// CreateMachineType creates a new machine type
func (h *MachineTypeHandlers) CreateMachineType(c *gin.Context) {
	ctx := c.Request.Context()

	// Only admin users can create machine types
	userName := c.GetHeader("X-User-Name")
	if userName == "" {
		userName = c.GetHeader("X-User-Email")
	}
	if userName == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// TODO: Check if user has admin role

	var req MachineTypeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	machineType := &machinesv1.MachineType{
		ObjectMeta: metav1.ObjectMeta{
			Name: req.Name,
		},
		Spec: req.Spec,
	}

	// Apply defaults via webhook
	if err := h.manager.MachineTypeWebhook.Default(ctx, machineType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate via webhook
	if err := h.manager.MachineTypeWebhook.ValidateCreate(ctx, machineType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Create the resource
	if err := h.manager.MachineTypeRepository.Create(ctx, machineType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, machineType)
}

// UpdateMachineType updates an existing machine type
func (h *MachineTypeHandlers) UpdateMachineType(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	// Only admin users can update machine types
	userName := c.GetHeader("X-User-Name")
	if userName == "" {
		userName = c.GetHeader("X-User-Email")
	}
	if userName == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req MachineTypeUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get existing machine type
	existing, err := h.manager.MachineTypeRepository.Get(ctx, name)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Machine type not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Update spec
	oldMachineType := existing.DeepCopy()
	existing.Spec = req.Spec

	// Apply defaults via webhook
	if err := h.manager.MachineTypeWebhook.Default(ctx, existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate via webhook
	if err := h.manager.MachineTypeWebhook.ValidateUpdate(ctx, oldMachineType, existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Update the resource
	if err := h.manager.MachineTypeRepository.Update(ctx, existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, existing)
}

// DeleteMachineType deletes a machine type
func (h *MachineTypeHandlers) DeleteMachineType(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	// Only admin users can delete machine types
	userName := c.GetHeader("X-User-Name")
	if userName == "" {
		userName = c.GetHeader("X-User-Email")
	}
	if userName == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get existing machine type
	existing, err := h.manager.MachineTypeRepository.Get(ctx, name)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Machine type not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate via webhook
	if err := h.manager.MachineTypeWebhook.ValidateDelete(ctx, existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Delete the resource
	if err := h.manager.MachineTypeRepository.Delete(ctx, name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Machine type deleted successfully",
	})
}

// ActivateMachineType activates a machine type
func (h *MachineTypeHandlers) ActivateMachineType(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	// Only admin users can activate machine types
	userName := c.GetHeader("X-User-Name")
	if userName == "" {
		userName = c.GetHeader("X-User-Email")
	}
	if userName == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get existing machine type
	machineType, err := h.manager.MachineTypeRepository.Get(ctx, name)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Machine type not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set active to true
	machineType.Spec.Active = true

	// Apply defaults via webhook
	if err := h.manager.MachineTypeWebhook.Default(ctx, machineType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Update the resource
	if err := h.manager.MachineTypeRepository.Update(ctx, machineType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": machineType,
	})
}

// DeactivateMachineType deactivates a machine type
func (h *MachineTypeHandlers) DeactivateMachineType(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	// Only admin users can deactivate machine types
	userName := c.GetHeader("X-User-Name")
	if userName == "" {
		userName = c.GetHeader("X-User-Email")
	}
	if userName == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get existing machine type
	machineType, err := h.manager.MachineTypeRepository.Get(ctx, name)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Machine type not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set active to false
	machineType.Spec.Active = false

	// Apply defaults via webhook
	if err := h.manager.MachineTypeWebhook.Default(ctx, machineType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Update the resource
	if err := h.manager.MachineTypeRepository.Update(ctx, machineType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": machineType,
	})
}

// ToggleMachineTypeActive toggles the active state of a machine type
func (h *MachineTypeHandlers) ToggleMachineTypeActive(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	// Only admin users can toggle machine types
	userName := c.GetHeader("X-User-Name")
	if userName == "" {
		userName = c.GetHeader("X-User-Email")
	}
	if userName == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get existing machine type
	machineType, err := h.manager.MachineTypeRepository.Get(ctx, name)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Machine type not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Toggle active state
	machineType.Spec.Active = !machineType.Spec.Active

	// Apply defaults via webhook
	if err := h.manager.MachineTypeWebhook.Default(ctx, machineType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Update the resource
	if err := h.manager.MachineTypeRepository.Update(ctx, machineType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Machine type active state toggled successfully",
		"active":  machineType.Spec.Active,
	})
}