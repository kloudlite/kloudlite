package workmachine

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WorkspaceActivitySummary contains aggregated workspace activity information
type WorkspaceActivitySummary struct {
	TotalWorkspaces      int
	ActiveWorkspaceCount int32
	AllWorkspacesIdle    bool
	LastActivityTime     *metav1.Time
}

// aggregateWorkspaceStates lists all workspaces for a WorkMachine and aggregates their activity states
func (r *WorkMachineReconciler) aggregateWorkspaceStates(ctx context.Context, workMachineName string) (*WorkspaceActivitySummary, error) {
	workspaceList := &workspacev1.WorkspaceList{}
	if err := r.List(ctx, workspaceList); err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}

	summary := &WorkspaceActivitySummary{
		AllWorkspacesIdle: true,
	}

	for _, ws := range workspaceList.Items {
		// Filter by WorkMachine name
		if ws.Spec.WorkmachineName != workMachineName {
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

// checkAutoShutdown is a reconciliation step that handles automatic WorkMachine shutdown
// based on workspace idle states
func (r *WorkMachineReconciler) checkAutoShutdown(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	ctx := check.Context()

	// Aggregate workspace states
	summary, err := r.aggregateWorkspaceStates(ctx, obj.Name)
	if err != nil {
		return check.Errored(fmt.Errorf("failed to aggregate workspace states: %w", err))
	}

	// Update status fields
	obj.Status.ActiveWorkspaceCount = summary.ActiveWorkspaceCount
	if summary.LastActivityTime != nil {
		obj.Status.LastWorkspaceActivity = summary.LastActivityTime
	}

	now := metav1.Now()

	// Case 1: At least one workspace is active
	if !summary.AllWorkspacesIdle {
		// Reset idle tracking
		if obj.Status.AllIdleSince != nil {
			check.Logger().Info("workspace became active, resetting auto-shutdown timer",
				"activeWorkspaceCount", summary.ActiveWorkspaceCount)
			obj.Status.AllIdleSince = nil
		}
		// Clear IsAutoStopped flag if machine was restarted and is now active
		if obj.Status.IsAutoStopped {
			obj.Status.IsAutoStopped = false
		}
		return check.Passed()
	}

	// Case 2: All workspaces are idle (or no workspaces exist)
	idleThresholdMinutes := obj.Spec.AutoShutdown.IdleThresholdMinutes
	if idleThresholdMinutes <= 0 {
		idleThresholdMinutes = 30 // Default to 30 minutes
	}
	idleThreshold := time.Duration(idleThresholdMinutes) * time.Minute

	// Start idle timer if not already started
	if obj.Status.AllIdleSince == nil {
		check.Logger().Info("all workspaces are idle, starting auto-shutdown timer",
			"totalWorkspaces", summary.TotalWorkspaces,
			"idleThresholdMinutes", idleThresholdMinutes)
		obj.Status.AllIdleSince = &now

		// Requeue after idle threshold
		return check.UpdateMsg(fmt.Sprintf("All workspaces idle, auto-shutdown in %d minutes", idleThresholdMinutes)).RequeueAfter(idleThreshold)
	}

	// Calculate idle duration
	idleDuration := time.Since(obj.Status.AllIdleSince.Time)

	// Case 3: Idle threshold not yet reached
	if idleDuration < idleThreshold {
		remaining := idleThreshold - idleDuration
		check.Logger().Info("waiting for idle threshold",
			"idleDuration", idleDuration.Round(time.Second),
			"idleThreshold", idleThreshold,
			"remaining", remaining.Round(time.Second))
		return check.UpdateMsg(fmt.Sprintf("Auto-shutdown in %.0f minutes", remaining.Minutes())).RequeueAfter(remaining)
	}

	// Case 4: Idle threshold reached - trigger auto-stop
	check.Logger().Info("idle threshold reached, triggering auto-shutdown",
		"idleDuration", idleDuration.Round(time.Second),
		"idleThreshold", idleThreshold,
		"totalWorkspaces", summary.TotalWorkspaces)

	// Set spec.state to trigger the stop flow in setupCloudMachine
	obj.Spec.State = v1.MachineStateStopped
	obj.Status.IsAutoStopped = true
	obj.Status.AllIdleSince = nil // Clear since we're stopping

	// Update the spec (triggers the existing stop logic)
	if err := r.Update(ctx, obj); err != nil {
		return check.Errored(fmt.Errorf("failed to update WorkMachine spec for auto-shutdown: %w", err))
	}

	return check.UpdateMsg("Auto-shutdown triggered due to idle timeout").RequeueAfter(5 * time.Second)
}
