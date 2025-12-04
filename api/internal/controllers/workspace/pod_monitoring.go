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
	podName := fmt.Sprintf("ws-%s", workspace.Name)

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
	// We use cat to combine IPv4 and IPv6 TCP connections, then filter for ESTABLISHED state (01)
	// This counts all ESTABLISHED TCP connections (excluding LISTEN sockets)
	// Note: We avoid shell redirections like 2>/dev/null as they may be blocked by command validation
	command := []string{"sh", "-c", "cat /proc/net/tcp /proc/net/tcp6 | awk '$4 == \"01\"' | wc -l"}

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

// checkAndSuspendIdleWorkspace checks workspace idle state and auto-suspends if enabled and idle timeout reached
// This always tracks idle state for UI display, but only auto-suspends when auto-stop is enabled
func (r *WorkspaceReconciler) checkAndSuspendIdleWorkspace(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	// Skip if workspace is not active
	if workspace.Spec.Status != "active" {
		return nil
	}

	// Check if workspace is idle
	isIdle, connectionCount, err := r.isWorkspaceIdle(ctx, workspace)
	if err != nil {
		logger.Warn("Failed to check workspace idle state", zap.Error(err))
		return nil // Don't fail reconciliation on metrics errors
	}

	// Track idle state transition instead of continuous timestamp updates
	// This prevents continuous status writes every 30 seconds
	now := metav1.Now()
	needsStatusUpdate := false

	// Update active connections count if it changed
	if workspace.Status.ActiveConnections != connectionCount {
		workspace.Status.ActiveConnections = connectionCount
		needsStatusUpdate = true
	}

	if !isIdle {
		// Workspace is active - only update status if state changed from idle to active
		if workspace.Status.IdleState == "idle" || workspace.Status.IdleState == "" {
			workspace.Status.IdleState = "active"
			workspace.Status.IdleSince = nil
			workspace.Status.LastActivityTime = &now
			needsStatusUpdate = true
			logger.Info("Workspace state changed to active", zap.String("workspace", workspace.Name))
		}

		// Only update status if something changed
		if needsStatusUpdate {
			if err := r.updateStatusPreservingPackages(ctx, workspace, logger); err != nil {
				logger.Warn("Failed to update workspace activity state", zap.Error(err))
			}
		}
		return nil
	}

	// Workspace is idle - only update status if state changed from active to idle
	if workspace.Status.IdleState != "idle" {
		workspace.Status.IdleState = "idle"
		workspace.Status.IdleSince = &now
		// Also update LastActivityTime to current time when transitioning to idle
		workspace.Status.LastActivityTime = &now
		needsStatusUpdate = true
		logger.Info("Workspace state changed to idle", zap.String("workspace", workspace.Name))
	}

	// Only update status if something changed
	if needsStatusUpdate {
		if err := r.updateStatusPreservingPackages(ctx, workspace, logger); err != nil {
			logger.Warn("Failed to update idle state", zap.Error(err))
		}
	}

	// Auto-suspend logic - only runs if auto-stop is enabled
	if workspace.Spec.Settings == nil || !workspace.Spec.Settings.AutoStop {
		return nil // Skip auto-suspend if not enabled, but idle tracking above still happens
	}

	// Need idleSince to be set to calculate duration
	if workspace.Status.IdleSince == nil {
		return nil
	}

	// Get idle timeout from workspace settings or use default
	idleTimeout := defaultIdleTimeoutMinutes
	if workspace.Spec.Settings.IdleTimeout > 0 {
		idleTimeout = int(workspace.Spec.Settings.IdleTimeout)
	}

	// Calculate idle duration
	idleDuration := time.Since(workspace.Status.IdleSince.Time)
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
		if err := r.Get(ctx, client.ObjectKey{Name: workspace.Name, Namespace: workspace.Namespace}, latest); err != nil {
			return fmt.Errorf("failed to fetch latest workspace: %w", err)
		}

		latest.Spec.Status = "suspended"
		if err := r.Update(ctx, latest); err != nil {
			return fmt.Errorf("failed to suspend idle workspace: %w", err)
		}
	}

	return nil
}
