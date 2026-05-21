package machinescoped

import (
	"context"
	"fmt"
	"time"

	workmachineshared "github.com/kloudlite/kloudlite/controllers/workmachine/shared"
	snapshotv1 "github.com/kloudlite/kloudlite/types/snapshot/v1"
	v1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/types/workspace/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// WorkspaceActivitySummary contains aggregated workspace activity information
type WorkspaceActivitySummary struct {
	TotalWorkspaces      int
	ActiveWorkspaceCount int32
	AllWorkspacesIdle    bool
	LastActivityTime     *metav1.Time
}

// aggregateWorkspaceStates lists all workspaces for a WorkMachine and aggregates their activity states
func (r *MachineScopedReconciler) aggregateWorkspaceStates(ctx context.Context, targetNamespace string) (*WorkspaceActivitySummary, error) {
	workspaceList := &workspacev1.WorkspaceList{}
	if err := r.List(ctx, workspaceList); err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}

	summary := &WorkspaceActivitySummary{
		AllWorkspacesIdle: true,
	}

	for _, ws := range workspaceList.Items {
		// Filter by namespace (workspaces live in the WorkMachine's targetNamespace)
		if ws.Namespace != targetNamespace {
			continue
		}
		summary.TotalWorkspaces++

		// Skip suspended/archived workspaces - they don't contribute to activity
		if ws.Spec.Status == "suspended" || ws.Spec.Status == "archived" {
			continue
		}

		// A workspace is considered active if:
		// - spec.status == "active" AND
		// - status.IdleState != "idle" OR has ActiveConnections > 0
		isWorkspaceActive := ws.Status.IdleState != "idle" || ws.Status.ActiveConnections > 0
		if isWorkspaceActive {
			summary.ActiveWorkspaceCount++
			summary.AllWorkspacesIdle = false
		}

		// Track latest activity time across all workspaces
		if ws.Status.LastActivityTime != nil {
			if summary.LastActivityTime == nil || ws.Status.LastActivityTime.After(summary.LastActivityTime.Time) {
				summary.LastActivityTime = ws.Status.LastActivityTime
			}
		}
	}

	return summary, nil
}

// hasInProgressSnapshots checks if there are any snapshot operations in progress
func (r *MachineScopedReconciler) hasInProgressSnapshots(ctx context.Context, namespace string) (bool, error) {
	// Check for in-progress Snapshot resources (state != Ready && state != Failed)
	snapshots := &snapshotv1.SnapshotList{}
	if err := r.List(ctx, snapshots); err != nil {
		return false, err
	}

	for _, snapshot := range snapshots.Items {
		// In-progress if state is not Ready or Failed
		if snapshot.Status.State != snapshotv1.SnapshotStateReady &&
			snapshot.Status.State != snapshotv1.SnapshotStateFailed {
			return true, nil
		}
	}

	// Also check for in-progress SnapshotRestore resources
	restores := &snapshotv1.SnapshotRestoreList{}
	if err := r.List(ctx, restores, client.InNamespace(namespace)); err != nil {
		return false, err
	}

	for _, restore := range restores.Items {
		if restore.Status.State != snapshotv1.SnapshotRestoreStateCompleted &&
			restore.Status.State != snapshotv1.SnapshotRestoreStateFailed {
			return true, nil
		}
	}

	return false, nil
}

// checkAutoShutdown is a reconciliation step that handles automatic WorkMachine shutdown
// based on workspace idle states
func (r *MachineScopedReconciler) checkAutoShutdown(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()

	// Skip auto-shutdown check if machine is not in running state
	if obj.Status.State != v1.MachineStateRunning {
		return markMachineReady(session, workmachineshared.ConditionAutoShutdownChecked, "auto-shutdown check is not required")
	}

	// Check for stale AllIdleSince from previous session
	// If allIdleSince is from before the machine started, it's stale and should be cleared
	if obj.Status.AllIdleSince != nil && obj.Status.StartedAt != nil {
		if obj.Status.AllIdleSince.Before(obj.Status.StartedAt) {
			ctrl.LoggerFrom(ctx).Info("clearing stale AllIdleSince timestamp from previous session",
				"allIdleSince", obj.Status.AllIdleSince,
				"startedAt", obj.Status.StartedAt)
			obj.Status.AllIdleSince = nil
			obj.Status.IsAutoStopped = false
			// Continue with fresh idle tracking below
		}
	}

	// Aggregate workspace states
	summary, err := r.aggregateWorkspaceStates(ctx, obj.Spec.TargetNamespace)
	if err != nil {
		return markMachineError(session, workmachineshared.ConditionAutoShutdownChecked, fmt.Errorf("failed to aggregate workspace states: %w", err))
	}

	// Update status fields
	obj.Status.ActiveWorkspaceCount = summary.ActiveWorkspaceCount
	if summary.LastActivityTime != nil {
		obj.Status.LastWorkspaceActivity = summary.LastActivityTime
	}

	// Check for in-progress snapshot operations
	hasSnapshots, err := r.hasInProgressSnapshots(ctx, obj.Spec.TargetNamespace)
	if err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed to check in-progress snapshots")
	}
	if hasSnapshots {
		ctrl.LoggerFrom(ctx).Info("in-progress snapshots detected, skipping auto-shutdown check")
		// Reset idle timer since snapshot operations are active
		if obj.Status.AllIdleSince != nil {
			obj.Status.AllIdleSince = nil
		}
		// Get check interval for requeuing
		var checkInterval int32 = 5
		if obj.Spec.AutoShutdown != nil && obj.Spec.AutoShutdown.CheckIntervalMinutes > 0 {
			checkInterval = obj.Spec.AutoShutdown.CheckIntervalMinutes
		}
		return markMachineWaiting(session, workmachineshared.ConditionAutoShutdownChecked, "in-progress snapshots detected, skipping auto-shutdown check", time.Duration(checkInterval)*time.Minute)
	}

	now := metav1.Now()

	// Case 1: At least one workspace is active
	if !summary.AllWorkspacesIdle {
		// Reset idle tracking
		if obj.Status.AllIdleSince != nil {
			ctrl.LoggerFrom(ctx).Info("workspace became active, resetting auto-shutdown timer",
				"activeWorkspaceCount", summary.ActiveWorkspaceCount)
			obj.Status.AllIdleSince = nil
		}
		// Clear IsAutoStopped flag if machine was restarted and is now active
		if obj.Status.IsAutoStopped {
			obj.Status.IsAutoStopped = false
		}
		return markMachineReady(session, workmachineshared.ConditionAutoShutdownChecked, "auto-shutdown check completed")
	}

	// Case 2: All workspaces are idle (or no workspaces exist)
	var idleThresholdMinutes int32 = 30 // Default to 30 minutes
	if obj.Spec.AutoShutdown != nil && obj.Spec.AutoShutdown.IdleThresholdMinutes > 0 {
		idleThresholdMinutes = obj.Spec.AutoShutdown.IdleThresholdMinutes
	}
	idleThreshold := time.Duration(idleThresholdMinutes) * time.Minute

	// Get check interval for requeuing
	var checkInterval int32 = 5 // Default to 5 minutes
	if obj.Spec.AutoShutdown != nil && obj.Spec.AutoShutdown.CheckIntervalMinutes > 0 {
		checkInterval = obj.Spec.AutoShutdown.CheckIntervalMinutes
	}

	// Start idle timer if not already started
	if obj.Status.AllIdleSince == nil {
		ctrl.LoggerFrom(ctx).Info("all workspaces are idle, starting auto-shutdown timer",
			"totalWorkspaces", summary.TotalWorkspaces,
			"idleThresholdMinutes", idleThresholdMinutes)
		obj.Status.AllIdleSince = &now

		// Requeue for monitoring (use Requeue instead of Passed to ensure requeue is honored)
		return markMachineWaiting(session, workmachineshared.ConditionAutoShutdownChecked, "all workspaces are idle, waiting for auto-shutdown threshold", time.Duration(checkInterval)*time.Minute)
	}

	// Calculate idle duration
	idleDuration := time.Since(obj.Status.AllIdleSince.Time)

	// Case 3: Idle threshold not yet reached
	if idleDuration < idleThreshold {
		remaining := idleThreshold - idleDuration
		ctrl.LoggerFrom(ctx).Info("waiting for idle threshold",
			"idleDuration", idleDuration.Round(time.Second),
			"idleThreshold", idleThreshold,
			"remaining", remaining.Round(time.Second))
		// Requeue for next check (use Requeue instead of Passed to ensure requeue is honored)
		return markMachineWaiting(session, workmachineshared.ConditionAutoShutdownChecked, "waiting for idle threshold", time.Duration(checkInterval)*time.Minute)
	}

	// Case 4: Idle threshold reached - trigger auto-stop
	ctrl.LoggerFrom(ctx).Info("idle threshold reached, triggering auto-shutdown",
		"idleDuration", idleDuration.Round(time.Second),
		"idleThreshold", idleThreshold,
		"totalWorkspaces", summary.TotalWorkspaces)

	obj.Status.IsAutoStopped = true
	obj.Status.AllIdleSince = nil // Clear since we're stopping

	// Update only the main resource spec. Status changes remain on the session object
	// and are written by the controller's final status patch.
	specObj := obj.DeepCopy()
	specObj.Status = *obj.Status.DeepCopy()
	specObj.Spec.State = v1.MachineStateStopped
	if err := r.Patch(ctx, specObj, client.MergeFrom(obj.DeepCopy())); err != nil {
		return markMachineError(session, workmachineshared.ConditionAutoShutdownChecked, fmt.Errorf("failed to update WorkMachine spec for auto-shutdown: %w", err))
	}
	obj.Spec.State = v1.MachineStateStopped

	ctrl.LoggerFrom(ctx).V(1).Info("Using configured auto-shutdown trigger retry interval", "interval", r.Cfg.WorkMachine.AutoShutdownTriggerRetryInterval)
	return markMachineWaiting(session, workmachineshared.ConditionAutoShutdownChecked, "Auto-shutdown triggered due to idle timeout", r.Cfg.WorkMachine.AutoShutdownTriggerRetryInterval)
}
