package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kloudlite/kloudlite/api/internal/managers"
	"github.com/kloudlite/kloudlite/api/internal/middleware"
	machinesv1 "github.com/kloudlite/kloudlite/api/pkg/apis/machines/v1"
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
	Name string                     `json:"name" binding:"required"`
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
	_, _, exists := middleware.GetUserFromContext(c)
	if !exists {
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

	// Create the resource
	if err := h.manager.MachineTypeRepository.Create(ctx, machineType); err != nil {
		// Check if this is a webhook validation error
		if strings.Contains(err.Error(), "admission webhook") && strings.Contains(err.Error(), "denied the request") {
			errorMsg := "Machine type validation failed"
			if strings.Contains(err.Error(), "already exists") {
				errorMsg = "A machine type with this name already exists"
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
			return
		}
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
	_, _, exists := middleware.GetUserFromContext(c)
	if !exists {
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
	_ = existing.DeepCopy()
	existing.Spec = req.Spec

	// Apply defaults via webhook
	if err := h.manager.MachineTypeWebhook.Default(ctx, existing); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Update the resource
	if err := h.manager.MachineTypeRepository.Update(ctx, existing); err != nil {
		// Check if this is a webhook validation error
		if strings.Contains(err.Error(), "admission webhook") && strings.Contains(err.Error(), "denied the request") {
			errorMsg := "Machine type validation failed"
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
			return
		}
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
	_, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get existing machine type
	_, err := h.manager.MachineTypeRepository.Get(ctx, name)
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
	_, _, exists := middleware.GetUserFromContext(c)
	if !exists {
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

	// Update the resource
	if err := h.manager.MachineTypeRepository.Update(ctx, machineType); err != nil {
		// Check if this is a webhook validation error
		if strings.Contains(err.Error(), "admission webhook") && strings.Contains(err.Error(), "denied the request") {
			errorMsg := "Machine type validation failed"
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    machineType,
	})
}

// DeactivateMachineType deactivates a machine type
func (h *MachineTypeHandlers) DeactivateMachineType(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	// Only admin users can deactivate machine types
	_, _, exists := middleware.GetUserFromContext(c)
	if !exists {
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

	// Update the resource
	if err := h.manager.MachineTypeRepository.Update(ctx, machineType); err != nil {
		// Check if this is a webhook validation error
		if strings.Contains(err.Error(), "admission webhook") && strings.Contains(err.Error(), "denied the request") {
			errorMsg := "Machine type validation failed"
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    machineType,
	})
}

// ToggleMachineTypeActive toggles the active state of a machine type
func (h *MachineTypeHandlers) ToggleMachineTypeActive(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	// Only admin users can toggle machine types
	_, _, exists := middleware.GetUserFromContext(c)
	if !exists {
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

	// Update the resource
	if err := h.manager.MachineTypeRepository.Update(ctx, machineType); err != nil {
		// Check if this is a webhook validation error
		if strings.Contains(err.Error(), "admission webhook") && strings.Contains(err.Error(), "denied the request") {
			errorMsg := "Machine type validation failed"
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
			return
		}
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
