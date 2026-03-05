package workmachine

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	environmentV1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// handleMachineTypeChange handles the machine type change process
// This is only applicable for cloud provider machines (AWS, GCP, Azure)
//
// The machine type change process follows these steps:
// 1. Suspend all workspaces
// 2. Deactivate all environments
// 3. Verify workspaces are suspended
// 4. Verify environments are deactivated
// 5. Stop the machine
// 6. Change the machine type (cloud provider API)
// 7. Start the machine
// 8. Wait for node to rejoin cluster and be ready
func (r *WorkMachineReconciler) handleMachineTypeChange(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// Early return: Only applicable for cloud machines
	if obj.Status.MachineID == "" {
		return check.Passed()
	}

	// Early return: Initialize current machine type if not set
	if obj.Status.CurrentMachineType == "" {
		obj.Status.CurrentMachineType = obj.Spec.MachineType
		return check.Passed()
	}

	// Early return: No change requested
	if !r.hasMachineTypeChanged(obj) {
		return check.Passed()
	}

	// Early return: Initialize the change process
	if !obj.Status.MachineTypeChanging {
		return r.initiateMachineTypeChange(check, obj)
	}

	// Execute the machine type change state machine
	return r.executeMachineTypeChange(check, obj)
}

// hasMachineTypeChanged checks if the machine type has changed from current state
func (r *WorkMachineReconciler) hasMachineTypeChanged(obj *v1.WorkMachine) bool {
	// If types match, clear any lingering change-in-progress state
	if obj.Spec.MachineType == obj.Status.CurrentMachineType {
		if obj.Status.MachineTypeChanging {
			obj.Status.MachineTypeChanging = false
			obj.Status.MachineTypeChangeMessage = ""
		}
		return false
	}
	return true
}

// initiateMachineTypeChange initializes the machine type change process
func (r *WorkMachineReconciler) initiateMachineTypeChange(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	oldType := obj.Status.CurrentMachineType
	newType := obj.Spec.MachineType

	obj.Status.MachineTypeChanging = true
	obj.Status.MachineTypeChangeMessage = fmt.Sprintf("Starting machine type change from %s to %s", oldType, newType)

	check.Logger().Info("Machine type change initiated",
		"from", oldType,
		"to", newType,
		"workMachine", obj.Name)

	check.Logger().Debug("Using configured machine type change retry interval", zap.Duration("interval", r.Cfg.WorkMachine.MachineTypeChangeRetryInterval))
	return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(r.Cfg.WorkMachine.MachineTypeChangeRetryInterval)
}

// executeMachineTypeChange executes the state machine for machine type change
// Each step is a separate function that returns early when not ready to proceed
func (r *WorkMachineReconciler) executeMachineTypeChange(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// Step 1-2: Suspend workspaces and deactivate environments
	if result := r.ensureWorkspacesSuspended(check, obj); !result.ShouldProceed() {
		return result
	}

	if result := r.ensureEnvironmentsDeactivated(check, obj); !result.ShouldProceed() {
		return result
	}

	// Step 3-4: Verify all workspaces and environments are in desired state
	if result := r.verifyWorkspacesSuspended(check, obj); !result.ShouldProceed() {
		return result
	}

	if result := r.verifyEnvironmentsDeactivated(check, obj); !result.ShouldProceed() {
		return result
	}

	// Step 5: Stop the machine
	if result := r.ensureMachineStopped(check, obj); !result.ShouldProceed() {
		return result
	}

	// Step 6-7: Change machine type and start machine
	if result := r.changeMachineType(check, obj); !result.ShouldProceed() {
		return result
	}

	// Step 8: Wait for node to be ready
	return r.waitForNodeReady(check, obj)
}

// ensureWorkspacesSuspended ensures all workspaces are suspended
// Returns early if suspension is in progress
func (r *WorkMachineReconciler) ensureWorkspacesSuspended(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	if err := r.suspendAllWorkspaces(check.Context(), obj.Name); err != nil {
		obj.Status.MachineTypeChangeMessage = fmt.Sprintf("Failed to suspend workspaces: %v", err)
		return check.Failed(err)
	}

	obj.Status.MachineTypeChangeMessage = "Workspaces suspended"
	check.Logger().Info("All workspaces suspended", "workMachine", obj.Name)
	return check.Passed()
}

// ensureEnvironmentsDeactivated ensures all environments are deactivated
// Returns early if deactivation is in progress
func (r *WorkMachineReconciler) ensureEnvironmentsDeactivated(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	if err := r.deactivateAllEnvironments(check.Context(), obj.Name); err != nil {
		obj.Status.MachineTypeChangeMessage = fmt.Sprintf("Failed to deactivate environments: %v", err)
		return check.Failed(err)
	}

	obj.Status.MachineTypeChangeMessage = "Environments deactivated, workspaces suspended"
	check.Logger().Info("All environments deactivated", "workMachine", obj.Name)
	return check.Passed()
}

// verifyWorkspacesSuspended checks if all workspaces are suspended
// Returns with requeue if any workspaces are still active
func (r *WorkMachineReconciler) verifyWorkspacesSuspended(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	workspaceList := &workspacev1.WorkspaceList{}
	if err := r.List(check.Context(), workspaceList); err != nil {
		return check.Failed(fmt.Errorf("failed to list workspaces: %w", err))
	}

	activeWorkspaceCount := r.countActiveWorkspaces(workspaceList, obj.Name)
	if activeWorkspaceCount > 0 {
		obj.Status.MachineTypeChangeMessage = fmt.Sprintf("Waiting for %d workspaces to suspend", activeWorkspaceCount)
		return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(r.Cfg.WorkMachine.MachineTypeChangeRetryInterval)
	}

	return check.Passed()
}

// verifyEnvironmentsDeactivated checks if all environments are deactivated
// Returns with requeue if any environments are still active
func (r *WorkMachineReconciler) verifyEnvironmentsDeactivated(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	envList := &environmentV1.EnvironmentList{}
	if err := r.List(check.Context(), envList); err != nil {
		return check.Failed(fmt.Errorf("failed to list environments: %w", err))
	}

	activeEnvironmentCount := r.countActiveEnvironments(envList, obj.Name)
	if activeEnvironmentCount > 0 {
		obj.Status.MachineTypeChangeMessage = fmt.Sprintf("Waiting for %d environments to deactivate", activeEnvironmentCount)
		return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(r.Cfg.WorkMachine.MachineTypeChangeRetryInterval)
	}

	return check.Passed()
}

// ensureMachineStopped ensures the cloud machine is stopped before changing type
// Returns with requeue if machine is still stopping
func (r *WorkMachineReconciler) ensureMachineStopped(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	machineInfo, err := r.cloudProviderAPI.GetMachineStatus(check.Context(), obj.Status.MachineID)
	if err != nil {
		return check.Failed(fmt.Errorf("failed to get machine status: %w", err))
	}

	// If machine is running, stop it
	if machineInfo.State == v1.MachineStateRunning {
		if err := r.cloudProviderAPI.StopMachine(check.Context(), obj.Status.MachineID); err != nil {
			return check.Failed(fmt.Errorf("failed to stop machine: %w", err))
		}
		obj.Status.State = v1.MachineStateStopping
		obj.Status.MachineTypeChangeMessage = "Stopping machine"
		return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(r.Cfg.WorkMachine.CloudMachineStopRetryInterval)
	}

	// If machine is stopping, wait for it to complete
	if machineInfo.State == v1.MachineStateStopping {
		obj.Status.MachineTypeChangeMessage = "Waiting for machine to stop"
		return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(r.Cfg.WorkMachine.CloudMachineStopRetryInterval)
	}

	// Machine is stopped, proceed to next step
	return check.Passed()
}

// changeMachineType changes the machine type via cloud provider API and starts the machine
// Returns with requeue after initiating the start
func (r *WorkMachineReconciler) changeMachineType(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// Only proceed if machine is stopped
	machineInfo, err := r.cloudProviderAPI.GetMachineStatus(check.Context(), obj.Status.MachineID)
	if err != nil {
		return check.Failed(fmt.Errorf("failed to get machine status: %w", err))
	}

	if machineInfo.State != v1.MachineStateStopped {
		// Machine not stopped yet, wait for ensureMachineStopped step
		return check.UpdateMsg("Waiting for machine to stop before changing type").RequeueAfter(r.Cfg.WorkMachine.MachineTypeChangeRetryInterval)
	}

	// Change the machine type via cloud provider API
	if err := r.cloudProviderAPI.ChangeMachine(check.Context(), obj.Status.MachineID, obj.Spec.MachineType); err != nil {
		return check.Failed(fmt.Errorf("failed to change machine type: %w", err))
	}

	// Update current machine type to reflect the change
	oldType := obj.Status.CurrentMachineType
	newType := obj.Spec.MachineType
	obj.Status.CurrentMachineType = newType
	obj.Status.MachineTypeChangeMessage = fmt.Sprintf("Machine type changed to %s", newType)

	check.Logger().Info("Machine type changed successfully",
		"oldType", oldType,
		"newType", newType,
		"workMachine", obj.Name)

	// Start the machine with the new type
	if err := r.cloudProviderAPI.StartMachine(check.Context(), obj.Status.MachineID); err != nil {
		return check.Failed(fmt.Errorf("failed to start machine after type change: %w", err))
	}

	obj.Status.State = v1.MachineStateStarting
	obj.Status.MachineTypeChangeMessage = "Starting machine with new type"
	return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(10 * time.Second)
}

// waitForNodeReady waits for the Kubernetes node to rejoin and become ready
// Returns passed when node is ready, otherwise requeues
func (r *WorkMachineReconciler) waitForNodeReady(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// Only proceed if we're in starting state
	if obj.Status.State != v1.MachineStateStarting {
		return check.Passed()
	}

	// Get the node
	node := &corev1.Node{}
	if err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, node); err != nil {
		if apiErrors.IsNotFound(err) {
			obj.Status.MachineTypeChangeMessage = "Waiting for node to rejoin cluster"
			return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(10 * time.Second)
		}
		return check.Failed(fmt.Errorf("failed to get node: %w", err))
	}

	// Check if node is ready
	if !r.isNodeReady(node) {
		obj.Status.MachineTypeChangeMessage = "Node joined, waiting for node to be ready"
		return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(5 * time.Second)
	}

	// Machine type change complete - mark as running
	r.markMachineTypeChangeComplete(obj, node)
	check.Logger().Info("Machine type change completed successfully",
		"newType", obj.Spec.MachineType,
		"workMachine", obj.Name)

	return check.Passed()
}

// countActiveWorkspaces counts the number of active workspaces for a given WorkMachine
func (r *WorkMachineReconciler) countActiveWorkspaces(list *workspacev1.WorkspaceList, workMachineName string) int {
	count := 0
	for _, ws := range list.Items {
		if ws.Spec.WorkmachineName != workMachineName {
			continue
		}
		// Consider suspended and archived as inactive
		if ws.Spec.Status != "suspended" && ws.Spec.Status != "archived" {
			count++
		}
	}
	return count
}

// countActiveEnvironments counts the number of active environments for a given WorkMachine
func (r *WorkMachineReconciler) countActiveEnvironments(list *environmentV1.EnvironmentList, workMachineName string) int {
	count := 0
	for _, env := range list.Items {
		if env.Spec.WorkMachineName != workMachineName {
			continue
		}
		if env.Spec.Activated {
			count++
		}
	}
	return count
}

// isNodeReady checks if a Kubernetes node is in Ready condition
func (r *WorkMachineReconciler) isNodeReady(node *corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// markMachineTypeChangeComplete marks the machine type change as complete
func (r *WorkMachineReconciler) markMachineTypeChangeComplete(obj *v1.WorkMachine, node *corev1.Node) {
	obj.Status.State = v1.MachineStateRunning
	obj.Status.MachineTypeChanging = false
	obj.Status.MachineTypeChangeMessage = fmt.Sprintf("Machine type change complete: %s → %s",
		obj.Status.CurrentMachineType,
		obj.Spec.MachineType)
	obj.Status.StartedAt = &metav1.Time{Time: time.Now()}
}

// suspendAllWorkspaces suspends all active workspaces on the WorkMachine
func (r *WorkMachineReconciler) suspendAllWorkspaces(ctx context.Context, workMachineName string) error {
	workspaceList := &workspacev1.WorkspaceList{}
	if err := r.List(ctx, workspaceList); err != nil {
		return fmt.Errorf("failed to list workspaces: %w", err)
	}

	for _, workspace := range workspaceList.Items {
		if workspace.Spec.WorkmachineName != workMachineName {
			continue
		}
		if workspace.Spec.Status == "suspended" || workspace.Spec.Status == "archived" {
			continue
		}
		workspace.Spec.Status = "suspended"
		if err := r.Update(ctx, &workspace); err != nil {
			if apiErrors.IsConflict(err) {
				// Resource was modified, will be retried on next reconciliation
				continue
			}
			return fmt.Errorf("failed to suspend workspace %s: %w", workspace.Name, err)
		}
	}
	return nil
}

// deactivateAllEnvironments deactivates all active environments on the WorkMachine
func (r *WorkMachineReconciler) deactivateAllEnvironments(ctx context.Context, workMachineName string) error {
	envList := &environmentV1.EnvironmentList{}
	if err := r.List(ctx, envList); err != nil {
		return fmt.Errorf("failed to list environments: %w", err)
	}

	for _, env := range envList.Items {
		if env.Spec.WorkMachineName != workMachineName {
			continue
		}
		if !env.Spec.Activated {
			continue
		}
		env.Spec.Activated = false
		if err := r.Update(ctx, &env); err != nil {
			if apiErrors.IsConflict(err) {
				// Resource was modified, will be retried on next reconciliation
				continue
			}
			return fmt.Errorf("failed to deactivate environment %s: %w", env.Name, err)
		}
	}
	return nil
}
