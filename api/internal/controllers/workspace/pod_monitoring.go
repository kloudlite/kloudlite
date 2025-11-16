package workspace

import (
	"context"
	"fmt"
	"strings"
	"time"

	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// hasActiveConnections checks if there are active SSH or web connections to the workspace
// by examining active TCP connections in the pod
// Returns: hasConnections bool, connectionCount int, error
func (r *WorkspaceReconciler) hasActiveConnections(ctx context.Context, workspace *workspacev1.Workspace) (bool, int, error) {
	podName := fmt.Sprintf("workspace-%s", workspace.Name)

	// Get the target namespace from WorkMachine
	targetNamespace, err := r.getWorkspaceTargetNamespace(ctx, workspace)
	if err != nil {
		return false, 0, fmt.Errorf("failed to get target namespace: %w", err)
	}

	pod := &corev1.Pod{}
	err = r.Get(ctx, client.ObjectKey{Name: podName, Namespace: targetNamespace}, pod)
	if err != nil {
		return false, 0, fmt.Errorf("failed to get pod: %w", err)
	}

	// Check pod IP and if it's accessible
	if pod.Status.PodIP == "" {
		return false, 0, nil
	}

	// Check if pod is not ready yet (still initializing)
	if pod.Status.Phase != corev1.PodRunning {
		return true, 0, nil // Consider as active while starting
	}

	// If pod was just started (within last 2 minutes), consider it as having connections
	// This gives time for the user to connect after starting the workspace
	if pod.Status.StartTime != nil {
		timeSinceStart := time.Since(pod.Status.StartTime.Time)
		if timeSinceStart < 2*time.Minute {
			return true, 0, nil
		}
	}

	// Check for actual active network connections
	// We check /proc/net/tcp for ESTABLISHED connections (state 01)
	// Important ports: SSH (22=0016), ttyd (7681=1E01), code-server (8080=1F90), vscode-tunnel (8000=1F40)
	// Connection state 01 = ESTABLISHED, 0A = LISTEN

	// Get the main container name (usually the first container)
	if len(pod.Spec.Containers) == 0 {
		return false, 0, nil
	}

	// Count ESTABLISHED connections by checking /proc/net/tcp
	// Format: awk '$4 == "01"' /proc/net/tcp /proc/net/tcp6 2>/dev/null | wc -l
	// This counts all ESTABLISHED TCP connections (excluding LISTEN sockets)
	command := []string{"sh", "-c", "awk '$4 == \"01\"' /proc/net/tcp /proc/net/tcp6 2>/dev/null | wc -l"}

	output, err := r.execInPod(ctx, pod, pod.Spec.Containers[0].Name, command)
	if err != nil {
		// If we can't check connections, assume there might be connections (fail-safe)
		r.Logger.Warn("Failed to check active connections, assuming active",
			zap.String("workspace", workspace.Name),
			zap.Error(err),
		)
		return true, 0, nil
	}

	// Parse the connection count
	connectionCount := 0
	fmt.Sscanf(strings.TrimSpace(output), "%d", &connectionCount)

	// Log the connection count for debugging
	r.Logger.Info("Active connection check",
		zap.String("workspace", workspace.Name),
		zap.Int("connectionCount", connectionCount),
		zap.Bool("hasConnections", connectionCount > 0),
	)

	return connectionCount > 0, connectionCount, nil
}

// isWorkspaceIdle checks if a workspace has been idle by checking for active connections
// A workspace is considered idle ONLY if there are no active connections (SSH, ttyd, vscode, code-server)
// Returns: isIdle bool, connectionCount int, error
func (r *WorkspaceReconciler) isWorkspaceIdle(ctx context.Context, workspace *workspacev1.Workspace) (bool, int, error) {
	// Check for active connections - this is the ONLY factor that matters
	hasConnections, connectionCount, err := r.hasActiveConnections(ctx, workspace)
	if err != nil {
		// If we can't check connections, assume workspace is active (fail-safe)
		r.Logger.Warn("Failed to check active connections, assuming workspace is active",
			zap.String("workspace", workspace.Name),
			zap.Error(err),
		)
		return false, 0, nil
	}

	// Workspace is idle if there are NO active connections
	isIdle := !hasConnections

	// Log activity status for debugging
	r.Logger.Info("Workspace activity check",
		zap.String("workspace", workspace.Name),
		zap.Bool("hasConnections", hasConnections),
		zap.Int("connectionCount", connectionCount),
		zap.Bool("isIdle", isIdle),
	)

	return isIdle, connectionCount, nil
}

// checkAndSuspendIdleWorkspace checks if a workspace should be auto-suspended and suspends it if needed
func (r *WorkspaceReconciler) checkAndSuspendIdleWorkspace(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	// Skip if auto-stop is not enabled
	if workspace.Spec.Settings == nil || !workspace.Spec.Settings.AutoStop {
		return nil
	}

	// Skip if workspace is not active
	if workspace.Spec.Status != "active" {
		return nil
	}

	// Get idle timeout from workspace settings or use default
	idleTimeout := defaultIdleTimeoutMinutes
	if workspace.Spec.Settings.IdleTimeout > 0 {
		idleTimeout = int(workspace.Spec.Settings.IdleTimeout)
	}

	// Check if workspace is idle
	isIdle, connectionCount, err := r.isWorkspaceIdle(ctx, workspace)
	if err != nil {
		logger.Warn("Failed to check workspace idle state", zap.Error(err))
		return nil // Don't fail reconciliation on metrics errors
	}

	// Update active connections count in workspace status
	workspace.Status.ActiveConnections = connectionCount

	if !isIdle {
		// Workspace is active, update last activity time
		now := metav1.Now()
		if workspace.Status.LastActivityTime == nil ||
			time.Since(workspace.Status.LastActivityTime.Time) > 30*time.Second {
			workspace.Status.LastActivityTime = &now
			if err := r.updateStatusPreservingPackages(ctx, workspace, logger); err != nil {
				logger.Warn("Failed to update last activity time", zap.Error(err))
			}
		}
		return nil
	}

	// Workspace is idle, check if idle timeout has been reached
	if workspace.Status.LastActivityTime == nil {
		// No last activity time set, initialize it
		now := metav1.Now()
		workspace.Status.LastActivityTime = &now
		if err := r.updateStatusPreservingPackages(ctx, workspace, logger); err != nil {
			logger.Warn("Failed to initialize last activity time", zap.Error(err))
		}
		return nil
	}

	// Calculate idle duration
	idleDuration := time.Since(workspace.Status.LastActivityTime.Time)
	idleTimeoutDuration := time.Duration(idleTimeout) * time.Minute

	// Log idle duration for debugging
	logger.Info("Checking idle timeout",
		zap.String("workspace", workspace.Name),
		zap.Duration("idleDuration", idleDuration),
		zap.Duration("idleTimeout", idleTimeoutDuration),
		zap.Bool("willSuspend", idleDuration >= idleTimeoutDuration),
	)

	if idleDuration >= idleTimeoutDuration {
		// Idle timeout reached, suspend workspace
		logger.Info("Auto-suspending idle workspace",
			zap.String("workspace", workspace.Name),
			zap.Duration("idleDuration", idleDuration),
			zap.Duration("idleTimeout", idleTimeoutDuration),
		)

		// Fetch the latest version to avoid conflict errors
		latest := &workspacev1.Workspace{}
		if err := r.Get(ctx, client.ObjectKey{Name: workspace.Name}, latest); err != nil {
			return fmt.Errorf("failed to fetch latest workspace: %w", err)
		}

		latest.Spec.Status = "suspended"
		if err := r.Update(ctx, latest); err != nil {
			return fmt.Errorf("failed to suspend idle workspace: %w", err)
		}
	}

	return nil
}
