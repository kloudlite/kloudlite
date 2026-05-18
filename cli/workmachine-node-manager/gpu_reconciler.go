package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	zap2 "go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// GPUStatusReconciler monitors GPU hardware and updates node labels and resources
type GPUStatusReconciler struct {
	client.Client
	Logger          *zap2.Logger
	CmdExec         CommandExecutor
	NodeName        string
	LastGPUDetected bool
}

type GPUInfo struct {
	Count         int
	Product       string
	DriverVersion string
}

type GPUMetrics struct {
	Model             string
	DriverVersion     string
	Count             int
	MemoryTotal       int32
	MemoryUsed        int32
	MemoryFree        int32
	UtilizationGPU    int32
	UtilizationMemory int32
	Temperature       int32
	PowerDraw         float32
	PowerLimit        float32
}

func (r *GPUStatusReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap2.String("node", req.Name),
	)

	// Only reconcile our own node
	if req.Name != r.NodeName {
		return reconcile.Result{}, nil
	}

	logger.Info("Reconciling GPU status for node")

	// Fetch the node
	node := &corev1.Node{}
	if err := r.Get(ctx, req.NamespacedName, node); err != nil {
		logger.Error("Failed to get Node", zap2.Error(err))
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Detect GPU hardware
	gpuDetected := r.detectGPU(logger)

	// If GPU status changed, update node
	if gpuDetected != r.LastGPUDetected {
		logger.Info("GPU detection status changed",
			zap2.Bool("previouslyDetected", r.LastGPUDetected),
			zap2.Bool("currentlyDetected", gpuDetected))
		r.LastGPUDetected = gpuDetected
	}

	if !gpuDetected {
		logger.Info("No GPU detected on this node")

		// If GPU was previously detected, clean up GPU resources from the node
		if r.LastGPUDetected {
			logger.Info("Cleaning up GPU resources from node (machine type changed to non-GPU)")
			if err := r.cleanupNodeGPU(ctx, node, logger); err != nil {
				logger.Error("Failed to cleanup node GPU resources", zap2.Error(err))
				return reconcile.Result{}, err
			}
			logger.Info("Successfully cleaned up GPU resources from node")
		}

		return reconcile.Result{}, nil
	}

	// Ensure NVIDIA drivers are available and container runtime is configured
	setupErr := r.ensureNVIDIASetup(logger)
	if setupErr != nil {
		logger.Error("NVIDIA setup not ready", zap2.Error(setupErr))

		// Retry updating node with latest version on conflict
		for retries := 0; retries < 3; retries++ {
			// Refetch the latest node
			latestNode := &corev1.Node{}
			if err := r.Get(ctx, client.ObjectKey{Name: node.Name}, latestNode); err != nil {
				logger.Error("Failed to refetch node", zap2.Error(err))
				break
			}

			// Update labels with setup status
			if latestNode.Labels == nil {
				latestNode.Labels = make(map[string]string)
			}
			latestNode.Labels["nvidia.com/gpu.driver-status"] = "waiting"
			latestNode.Labels["nvidia.com/gpu.driver-message"] = sanitizeLabelValue(setupErr.Error(), 63)

			// Try to update node with status
			if updateErr := r.Update(ctx, latestNode); updateErr != nil {
				if strings.Contains(updateErr.Error(), "the object has been modified") {
					logger.Warn("Node was modified, retrying update", zap2.Int("retry", retries+1))
					time.Sleep(100 * time.Millisecond)
					continue
				}
				logger.Error("Failed to update node with driver status", zap2.Error(updateErr))
				break
			}

			logger.Info("Successfully updated node with driver status")
			break
		}

		return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// Get GPU information
	gpuInfo, err := r.getGPUInfo(logger)
	if err != nil {
		logger.Error("Failed to get GPU information", zap2.Error(err))

		// Retry updating node with latest version on conflict
		for retries := 0; retries < 3; retries++ {
			// Refetch the latest node
			latestNode := &corev1.Node{}
			if getErr := r.Get(ctx, client.ObjectKey{Name: node.Name}, latestNode); getErr != nil {
				logger.Error("Failed to refetch node", zap2.Error(getErr))
				break
			}

			// Update labels with status
			if latestNode.Labels == nil {
				latestNode.Labels = make(map[string]string)
			}
			latestNode.Labels["nvidia.com/gpu.driver-status"] = "error"
			latestNode.Labels["nvidia.com/gpu.driver-message"] = sanitizeLabelValue(err.Error(), 63)

			if updateErr := r.Update(ctx, latestNode); updateErr != nil {
				if strings.Contains(updateErr.Error(), "the object has been modified") {
					logger.Warn("Node was modified, retrying update", zap2.Int("retry", retries+1))
					time.Sleep(100 * time.Millisecond)
					continue
				}
				logger.Error("Failed to update node with GPU error status", zap2.Error(updateErr))
				break
			}

			logger.Info("Successfully updated node with GPU error status")
			break
		}

		return reconcile.Result{RequeueAfter: 30 * time.Second}, nil
	}

	logger.Info("GPU detected",
		zap2.Int("count", gpuInfo.Count),
		zap2.String("product", gpuInfo.Product),
		zap2.String("driverVersion", gpuInfo.DriverVersion))

	// Update node labels and resources
	if err := r.updateNodeGPU(ctx, node, gpuInfo, logger); err != nil {
		logger.Error("Failed to update node GPU status", zap2.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("Successfully updated node with GPU information")
	return reconcile.Result{}, nil
}

// ensureNVIDIASetup checks if NVIDIA drivers are available
// Note: The Deep Learning AMI comes with drivers and container runtime pre-installed
func (r *GPUStatusReconciler) ensureNVIDIASetup(logger *zap2.Logger) error {
	// Check if nvidia-smi is available (should be pre-installed in Deep Learning AMI)
	// Use full path since nsenter may not inherit full PATH from the host
	checkScript := "/usr/bin/nvidia-smi > /dev/null 2>&1"
	if _, err := r.CmdExec.Execute(checkScript); err != nil {
		logger.Info("NVIDIA drivers not available (nvidia-smi failed)")
		return fmt.Errorf("nvidia-smi not available")
	}

	logger.Info("NVIDIA drivers are available and working")
	return nil
}

func (r *GPUStatusReconciler) detectGPU(logger *zap2.Logger) bool {
	// Check for NVIDIA GPU by reading /sys/bus/pci/devices directly
	// This approach doesn't require lspci to be installed
	// Note: When using nsenter to enter host namespaces, paths like /sys are already host paths
	checkScript := `
		if [ -d /sys/bus/pci/devices ]; then
			for device in /sys/bus/pci/devices/*; do
				if [ -f "$device/vendor" ] && [ -f "$device/device" ]; then
					vendor=$(cat "$device/vendor" 2>/dev/null)
					# 0x10de is NVIDIA's PCI vendor ID
					if [ "$vendor" = "0x10de" ]; then
						exit 0
					fi
				fi
			done
		fi
		exit 1
	`
	_, err := r.CmdExec.Execute(checkScript)
	if err != nil {
		logger.Debug("No NVIDIA GPU detected in /sys/bus/pci/devices")
		return false
	}
	logger.Info("NVIDIA GPU detected via PCI device scan")
	return true
}

func (r *GPUStatusReconciler) getGPUInfo(logger *zap2.Logger) (*GPUInfo, error) {
	// Check if nvidia-smi is available
	checkScript := "nvidia-smi > /dev/null 2>&1"
	if _, err := r.CmdExec.Execute(checkScript); err != nil {
		return nil, fmt.Errorf("nvidia-smi not available or drivers not loaded")
	}

	// Get GPU count
	countScript := "nvidia-smi --query-gpu=name --format=csv,noheader | wc -l"
	countOutput, err := r.CmdExec.Execute(countScript)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU count: %w", err)
	}

	count := 1 // Default to 1
	if parsed, err := fmt.Sscanf(strings.TrimSpace(string(countOutput)), "%d", &count); err == nil && parsed == 1 {
		// Successfully parsed count
	}

	// Get GPU product name (normalized)
	productScript := "nvidia-smi --query-gpu=gpu_name --format=csv,noheader | head -1 | tr ' ' '-' | tr '[:upper:]' '[:lower:]'"
	productOutput, err := r.CmdExec.Execute(productScript)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU product: %w", err)
	}
	product := strings.TrimSpace(string(productOutput))

	// Get driver version
	driverScript := "nvidia-smi --query-gpu=driver_version --format=csv,noheader | head -1"
	driverOutput, err := r.CmdExec.Execute(driverScript)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver version: %w", err)
	}
	driverVersion := strings.TrimSpace(string(driverOutput))

	return &GPUInfo{
		Count:         count,
		Product:       product,
		DriverVersion: driverVersion,
	}, nil
}

func (r *GPUStatusReconciler) getGPUMetrics(logger *zap2.Logger) (*GPUMetrics, error) {
	// Query nvidia-smi for comprehensive metrics
	// Fields: name, driver_version, memory.total, memory.used, memory.free, utilization.gpu, utilization.memory, temperature.gpu, power.draw, power.limit
	metricsScript := "nvidia-smi --query-gpu=name,driver_version,memory.total,memory.used,memory.free,utilization.gpu,utilization.memory,temperature.gpu,power.draw,power.limit --format=csv,noheader,nounits | head -1"

	output, err := r.CmdExec.Execute(metricsScript)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU metrics: %w", err)
	}

	// Parse CSV output
	parts := strings.Split(strings.TrimSpace(string(output)), ",")
	if len(parts) < 10 {
		return nil, fmt.Errorf("unexpected nvidia-smi output format: got %d fields, expected 10", len(parts))
	}

	// Helper function to parse int32
	parseInt32 := func(s string, fieldName string) (int32, error) {
		s = strings.TrimSpace(s)
		var val int32
		if _, err := fmt.Sscanf(s, "%d", &val); err != nil {
			return 0, fmt.Errorf("failed to parse %s: %w", fieldName, err)
		}
		return val, nil
	}

	// Helper function to parse float32
	parseFloat32 := func(s string, fieldName string) (float32, error) {
		s = strings.TrimSpace(s)
		var val float32
		if _, err := fmt.Sscanf(s, "%f", &val); err != nil {
			return 0, fmt.Errorf("failed to parse %s: %w", fieldName, err)
		}
		return val, nil
	}

	memoryTotal, err := parseInt32(parts[2], "memory.total")
	if err != nil {
		return nil, err
	}

	memoryUsed, err := parseInt32(parts[3], "memory.used")
	if err != nil {
		return nil, err
	}

	memoryFree, err := parseInt32(parts[4], "memory.free")
	if err != nil {
		return nil, err
	}

	utilizationGPU, err := parseInt32(parts[5], "utilization.gpu")
	if err != nil {
		return nil, err
	}

	utilizationMemory, err := parseInt32(parts[6], "utilization.memory")
	if err != nil {
		return nil, err
	}

	temperature, err := parseInt32(parts[7], "temperature.gpu")
	if err != nil {
		return nil, err
	}

	powerDraw, err := parseFloat32(parts[8], "power.draw")
	if err != nil {
		return nil, err
	}

	powerLimit, err := parseFloat32(parts[9], "power.limit")
	if err != nil {
		return nil, err
	}

	// Get GPU count
	countScript := "nvidia-smi --query-gpu=name --format=csv,noheader | wc -l"
	countOutput, err := r.CmdExec.Execute(countScript)
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU count: %w", err)
	}

	count := 1 // Default to 1
	if parsed, err := fmt.Sscanf(strings.TrimSpace(string(countOutput)), "%d", &count); err == nil && parsed == 1 {
		// Successfully parsed count
	}

	return &GPUMetrics{
		Model:             strings.TrimSpace(parts[0]),
		DriverVersion:     strings.TrimSpace(parts[1]),
		Count:             count,
		MemoryTotal:       memoryTotal,
		MemoryUsed:        memoryUsed,
		MemoryFree:        memoryFree,
		UtilizationGPU:    utilizationGPU,
		UtilizationMemory: utilizationMemory,
		Temperature:       temperature,
		PowerDraw:         powerDraw,
		PowerLimit:        powerLimit,
	}, nil
}

func (r *GPUStatusReconciler) updateNodeGPU(ctx context.Context, node *corev1.Node, gpuInfo *GPUInfo, logger *zap2.Logger) error {
	// Update node labels
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}

	node.Labels["nvidia.com/gpu"] = "true"
	node.Labels["nvidia.com/gpu.count"] = fmt.Sprintf("%d", gpuInfo.Count)
	node.Labels["nvidia.com/gpu.product"] = gpuInfo.Product
	node.Labels["nvidia.com/gpu.driver-version"] = gpuInfo.DriverVersion
	node.Labels["nvidia.com/gpu.driver-status"] = "ready"
	node.Labels["nvidia.com/gpu.driver-message"] = "nvidia-drivers-operational"

	// Update node to apply labels
	if err := r.Update(ctx, node); err != nil {
		return fmt.Errorf("failed to update node labels: %w", err)
	}

	logger.Info("Updated node labels",
		zap2.String("gpu", "true"),
		zap2.Int("count", gpuInfo.Count),
		zap2.String("product", gpuInfo.Product),
		zap2.String("driverVersion", gpuInfo.DriverVersion))

	// Fetch latest node to update status (capacity and allocatable)
	updatedNode := &corev1.Node{}
	if err := r.Get(ctx, client.ObjectKey{Name: node.Name}, updatedNode); err != nil {
		return fmt.Errorf("failed to get latest node: %w", err)
	}

	// Update capacity and allocatable
	if updatedNode.Status.Capacity == nil {
		updatedNode.Status.Capacity = make(corev1.ResourceList)
	}
	if updatedNode.Status.Allocatable == nil {
		updatedNode.Status.Allocatable = make(corev1.ResourceList)
	}

	gpuQuantity := fmt.Sprintf("%d", gpuInfo.Count)
	updatedNode.Status.Capacity[corev1.ResourceName("nvidia.com/gpu")] = *parseQuantity(gpuQuantity)
	updatedNode.Status.Allocatable[corev1.ResourceName("nvidia.com/gpu")] = *parseQuantity(gpuQuantity)

	// Update node status
	if err := r.Status().Update(ctx, updatedNode); err != nil {
		return fmt.Errorf("failed to update node status: %w", err)
	}

	logger.Info("Updated node capacity and allocatable",
		zap2.String("nvidia.com/gpu", gpuQuantity))

	return nil
}

// cleanupNodeGPU removes GPU labels and resources from the node
// Called when machine type changes from GPU to non-GPU
func (r *GPUStatusReconciler) cleanupNodeGPU(ctx context.Context, node *corev1.Node, logger *zap2.Logger) error {
	// Remove GPU labels
	if node.Labels != nil {
		delete(node.Labels, "nvidia.com/gpu")
		delete(node.Labels, "nvidia.com/gpu.count")
		delete(node.Labels, "nvidia.com/gpu.product")
		delete(node.Labels, "nvidia.com/gpu.driver-version")
		delete(node.Labels, "nvidia.com/gpu.driver-status")
		delete(node.Labels, "nvidia.com/gpu.driver-message")
	}

	// Update node to apply label deletions
	if err := r.Update(ctx, node); err != nil {
		return fmt.Errorf("failed to remove GPU labels from node: %w", err)
	}

	logger.Info("Removed GPU labels from node")

	// Fetch latest node to update status (capacity and allocatable)
	updatedNode := &corev1.Node{}
	if err := r.Get(ctx, client.ObjectKey{Name: node.Name}, updatedNode); err != nil {
		return fmt.Errorf("failed to get latest node: %w", err)
	}

	// Remove GPU resources from capacity and allocatable
	if updatedNode.Status.Capacity != nil {
		delete(updatedNode.Status.Capacity, corev1.ResourceName("nvidia.com/gpu"))
	}
	if updatedNode.Status.Allocatable != nil {
		delete(updatedNode.Status.Allocatable, corev1.ResourceName("nvidia.com/gpu"))
	}

	// Update node status
	if err := r.Status().Update(ctx, updatedNode); err != nil {
		return fmt.Errorf("failed to remove GPU resources from node status: %w", err)
	}

	logger.Info("Removed GPU capacity and allocatable from node")

	return nil
}

func (r *GPUStatusReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		WithEventFilter(predicate.Funcs{
			// Only watch our own node
			CreateFunc: func(e event.CreateEvent) bool {
				return e.Object.GetName() == r.NodeName
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return e.ObjectNew.GetName() == r.NodeName
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false // Don't reconcile on delete
			},
		}).
		Complete(r)
}
