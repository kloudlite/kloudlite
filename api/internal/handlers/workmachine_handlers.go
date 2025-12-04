package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/internal/managers"
	"github.com/kloudlite/kloudlite/api/internal/middleware"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkMachineHandlers handles HTTP requests for WorkMachine resources
type WorkMachineHandlers struct {
	manager *managers.Manager
}

// NewWorkMachineHandlers creates a new WorkMachine handler
func NewWorkMachineHandlers(manager *managers.Manager) *WorkMachineHandlers {
	return &WorkMachineHandlers{
		manager: manager,
	}
}

// WorkMachineCreateRequest represents a request to create a WorkMachine
type WorkMachineCreateRequest struct {
	MachineType   string   `json:"machineType,omitempty"` // Optional - uses default if not specified
	SSHPublicKeys []string `json:"sshPublicKeys,omitempty"`
}

// AutoShutdownConfig represents auto-shutdown configuration in API requests
type AutoShutdownConfig struct {
	Enabled              bool  `json:"enabled"`
	IdleThresholdMinutes int32 `json:"idleThresholdMinutes,omitempty"`
}

// WorkMachineUpdateRequest represents a request to update a WorkMachine
type WorkMachineUpdateRequest struct {
	MachineType   string              `json:"machineType,omitempty"`
	SSHPublicKeys []string            `json:"sshPublicKeys,omitempty"`
	AutoShutdown  *AutoShutdownConfig `json:"autoShutdown,omitempty"`
}

// GetMyWorkMachine returns the current user's work machine
func (h *WorkMachineHandlers) GetMyWorkMachine(c *gin.Context) {
	ctx := c.Request.Context()

	// Get current user
	userName, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get user's machine
	machine, err := h.manager.WorkMachineRepository.GetByOwner(ctx, userName)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No work machine found for current user",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, machine)
}

// CreateMyWorkMachine creates a work machine for the current user
func (h *WorkMachineHandlers) CreateMyWorkMachine(c *gin.Context) {
	ctx := c.Request.Context()

	// Get current user
	userName, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req WorkMachineCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check if user already has a machine
	existingMachine, _ := h.manager.WorkMachineRepository.GetByOwner(ctx, userName)
	if existingMachine != nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "User already has a work machine",
		})
		return
	}

	// Determine machine type - use default if not specified
	machineType := req.MachineType
	if machineType == "" {
		defaultType, err := h.manager.MachineTypeRepository.GetDefault(ctx)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "No machine type specified and no default machine type configured",
			})
			return
		}
		machineType = defaultType.Name
	}

	// Generate WorkMachine name from userName
	// Extract username from email (part before @) and sanitize for k8s naming
	username := userName
	if idx := strings.Index(username, "@"); idx > 0 {
		username = username[:idx]
	}
	// Replace dots and special characters with hyphens for valid k8s names
	username = strings.ReplaceAll(username, ".", "-")
	username = strings.ReplaceAll(username, "_", "-")
	username = strings.ToLower(username)
	workMachineName := "wm-" + username

	// Create new machine
	machine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: workMachineName,
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:       userName,
			MachineType:   machineType,
			State:         machinesv1.MachineStateRunning, // Start as running for initial setup
			SSHPublicKeys: req.SSHPublicKeys,
		},
	}

	// Create the resource
	if err := h.manager.WorkMachineRepository.Create(ctx, machine); err != nil {
		// Check if this is a webhook validation error
		if strings.Contains(err.Error(), "admission webhook") && strings.Contains(err.Error(), "denied the request") {
			errorMsg := "Work machine validation failed"
			if strings.Contains(err.Error(), "already exists") {
				errorMsg = "A work machine with this configuration already exists"
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, machine)
}

// UpdateMyWorkMachine updates the current user's work machine
func (h *WorkMachineHandlers) UpdateMyWorkMachine(c *gin.Context) {
	ctx := c.Request.Context()

	// Get current user
	userName, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var req WorkMachineUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Get user's machine
	machine, err := h.manager.WorkMachineRepository.GetByOwner(ctx, userName)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No work machine found for current user",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	_ = machine.DeepCopy()

	// Update fields
	if req.MachineType != "" {
		machine.Spec.MachineType = req.MachineType
	}
	if req.SSHPublicKeys != nil {
		machine.Spec.SSHPublicKeys = req.SSHPublicKeys
	}
	if req.AutoShutdown != nil {
		machine.Spec.AutoShutdown = &machinesv1.AutoShutdownConfig{
			Enabled:              req.AutoShutdown.Enabled,
			IdleThresholdMinutes: req.AutoShutdown.IdleThresholdMinutes,
		}
	}

	// Update the resource
	if err := h.manager.WorkMachineRepository.Update(ctx, machine); err != nil {
		// Check if this is a webhook validation error
		if strings.Contains(err.Error(), "admission webhook") && strings.Contains(err.Error(), "denied the request") {
			errorMsg := "Work machine validation failed"
			c.JSON(http.StatusBadRequest, gin.H{"error": errorMsg})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, machine)
}

// DeleteMyWorkMachine deletes the current user's work machine
func (h *WorkMachineHandlers) DeleteMyWorkMachine(c *gin.Context) {
	ctx := c.Request.Context()

	// Get current user
	userName, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get user's machine
	machine, err := h.manager.WorkMachineRepository.GetByOwner(ctx, userName)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No work machine found for current user",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Delete the resource
	if err := h.manager.WorkMachineRepository.Delete(ctx, machine.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Work machine deleted successfully",
	})
}

// StartMyWorkMachine starts the current user's work machine
func (h *WorkMachineHandlers) StartMyWorkMachine(c *gin.Context) {
	ctx := c.Request.Context()

	// Get current user
	userName, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get user's machine
	machine, err := h.manager.WorkMachineRepository.GetByOwner(ctx, userName)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No work machine found for current user",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Start the machine
	if err := h.manager.WorkMachineRepository.StartMachine(ctx, machine.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Work machine starting",
		"state":   machinesv1.MachineStateStarting,
	})
}

// StopMyWorkMachine stops the current user's work machine
func (h *WorkMachineHandlers) StopMyWorkMachine(c *gin.Context) {
	ctx := c.Request.Context()

	// Get current user
	userName, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Get user's machine
	machine, err := h.manager.WorkMachineRepository.GetByOwner(ctx, userName)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No work machine found for current user",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Stop the machine
	if err := h.manager.WorkMachineRepository.StopMachine(ctx, machine.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Work machine stopping",
		"state":   machinesv1.MachineStateStopping,
	})
}

// ListAllWorkMachines lists all work machines (admin only)
func (h *WorkMachineHandlers) ListAllWorkMachines(c *gin.Context) {
	ctx := c.Request.Context()

	// Only admin users can list all machines
	_, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// TODO: Check if user has admin role

	// Get machine type filter
	machineType := c.Query("machineType")

	var list *machinesv1.WorkMachineList
	var err error

	if machineType != "" {
		list, err = h.manager.WorkMachineRepository.ListByMachineType(ctx, machineType)
	} else {
		list, err = h.manager.WorkMachineRepository.List(ctx)
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

// GetWorkMachine gets a specific work machine by name (admin only)
func (h *WorkMachineHandlers) GetWorkMachine(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Param("name")

	// Only admin users can get any machine
	_, _, _, exists := middleware.GetUserFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// TODO: Check if user has admin role

	machine, err := h.manager.WorkMachineRepository.Get(ctx, name)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Work machine not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, machine)
}

// NodeMetrics represents CPU and memory metrics for a node
type NodeMetrics struct {
	CPU struct {
		Usage       int64 `json:"usage"`       // in millicores
		Capacity    int64 `json:"capacity"`    // in millicores
		Allocatable int64 `json:"allocatable"` // in millicores
	} `json:"cpu"`
	Memory struct {
		Usage       int64 `json:"usage"`       // in bytes
		Capacity    int64 `json:"capacity"`    // in bytes
		Allocatable int64 `json:"allocatable"` // in bytes
	} `json:"memory"`
	Timestamp string `json:"timestamp"`
}

// GPUMetrics represents GPU metrics from workmachine-node-manager
type GPUMetrics struct {
	Detected          bool    `json:"detected"`
	Model             string  `json:"model,omitempty"`
	DriverVersion     string  `json:"driverVersion,omitempty"`
	Count             int     `json:"count,omitempty"`
	MemoryTotal       int32   `json:"memoryTotal,omitempty"`
	MemoryUsed        int32   `json:"memoryUsed,omitempty"`
	MemoryFree        int32   `json:"memoryFree,omitempty"`
	UtilizationGPU    int32   `json:"utilizationGpu,omitempty"`
	UtilizationMemory int32   `json:"utilizationMemory,omitempty"`
	Temperature       int32   `json:"temperature,omitempty"`
	PowerDraw         float32 `json:"powerDraw,omitempty"`
	PowerLimit        float32 `json:"powerLimit,omitempty"`
}

// MetricsStreamEvent represents a combined metrics event for SSE
type MetricsStreamEvent struct {
	NodeMetrics *NodeMetrics `json:"nodeMetrics,omitempty"`
	GPUMetrics  *GPUMetrics  `json:"gpuMetrics,omitempty"`
}

// GetWorkMachineMetricsStream handles GET /api/v1/work-machines/:name/metrics-stream
// This endpoint streams metrics updates via Server-Sent Events (SSE)
func (h *WorkMachineHandlers) GetWorkMachineMetricsStream(c *gin.Context) {
	ctx := c.Request.Context()

	// Get work machine name from URL parameter
	workMachineName := c.Param("name")
	if workMachineName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Work machine name is required",
		})
		return
	}

	// Verify WorkMachine exists
	wm, err := h.manager.WorkMachineRepository.Get(ctx, workMachineName)
	if err != nil {
		if client.IgnoreNotFound(err) == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Work machine not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Create a ticker for 5-second intervals
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	// Helper function to fetch node metrics
	fetchNodeMetrics := func() *NodeMetrics {
		reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		metrics := &NodeMetrics{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		}

		// Get node metrics from Kubernetes metrics API
		var nodeMetrics metricsv1beta1.NodeMetrics
		if err := h.manager.K8sClient.Get(reqCtx, client.ObjectKey{Name: wm.Name}, &nodeMetrics); err != nil {
			return metrics
		}

		// Get node to read capacity and allocatable resources
		var node corev1.Node
		if err := h.manager.K8sClient.Get(reqCtx, client.ObjectKey{Name: wm.Name}, &node); err != nil {
			return metrics
		}

		// CPU metrics
		if cpu, ok := nodeMetrics.Usage[corev1.ResourceCPU]; ok {
			metrics.CPU.Usage = cpu.MilliValue()
		}
		if capacity, ok := node.Status.Capacity[corev1.ResourceCPU]; ok {
			metrics.CPU.Capacity = capacity.MilliValue()
		}
		if allocatable, ok := node.Status.Allocatable[corev1.ResourceCPU]; ok {
			metrics.CPU.Allocatable = allocatable.MilliValue()
		}

		// Memory metrics
		if mem, ok := nodeMetrics.Usage[corev1.ResourceMemory]; ok {
			metrics.Memory.Usage = mem.Value()
		}
		if capacity, ok := node.Status.Capacity[corev1.ResourceMemory]; ok {
			metrics.Memory.Capacity = capacity.Value()
		}
		if allocatable, ok := node.Status.Allocatable[corev1.ResourceMemory]; ok {
			metrics.Memory.Allocatable = allocatable.Value()
		}

		return metrics
	}

	// Helper function to fetch GPU metrics
	fetchGPUMetrics := func() *GPUMetrics {
		reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		metricsURL := "http://hm-" + workMachineName + ".kloudlite-hostmanager:8081/metrics/gpu"
		req, err := http.NewRequestWithContext(reqCtx, "GET", metricsURL, nil)
		if err != nil {
			return nil
		}

		httpClient := &http.Client{
			Timeout: 3 * time.Second,
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil
		}

		var metrics GPUMetrics
		if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
			return nil
		}

		return &metrics
	}

	// Send initial metrics immediately
	event := MetricsStreamEvent{
		NodeMetrics: fetchNodeMetrics(),
		GPUMetrics:  fetchGPUMetrics(),
	}

	eventData, err := json.Marshal(event)
	if err == nil {
		c.Writer.Write([]byte("event: metrics\n"))
		c.Writer.Write([]byte("data: "))
		c.Writer.Write(eventData)
		c.Writer.Write([]byte("\n\n"))
		c.Writer.Flush()
	}

	// Stream metrics at regular intervals
	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			return
		case <-ticker.C:
			event := MetricsStreamEvent{
				NodeMetrics: fetchNodeMetrics(),
				GPUMetrics:  fetchGPUMetrics(),
			}

			eventData, err := json.Marshal(event)
			if err != nil {
				continue
			}

			// Write SSE event
			c.Writer.Write([]byte("event: metrics\n"))
			c.Writer.Write([]byte("data: "))
			c.Writer.Write(eventData)
			c.Writer.Write([]byte("\n\n"))
			c.Writer.Flush()
		}
	}
}
