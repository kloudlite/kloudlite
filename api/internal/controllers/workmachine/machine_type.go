package workmachine

import (
	"context"
	"fmt"
	"time"

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
func (r *WorkMachineReconciler) handleMachineTypeChange(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// Only applicable for cloud machines
	if obj.Status.MachineID == "" {
		return check.Passed()
	}

	// Initialize CurrentMachineType if not set
	if obj.Status.CurrentMachineType == "" {
		obj.Status.CurrentMachineType = obj.Spec.MachineType
		return check.Passed()
	}

	// Check if machine type has changed
	if obj.Spec.MachineType == obj.Status.CurrentMachineType {
		// No change, clear any change-in-progress state
		if obj.Status.MachineTypeChanging {
			obj.Status.MachineTypeChanging = false
			obj.Status.MachineTypeChangeMessage = ""
		}
		return check.Passed()
	}

	// Machine type change requested
	if !obj.Status.MachineTypeChanging {
		// Start the change process
		obj.Status.MachineTypeChanging = true
		obj.Status.MachineTypeChangeMessage = fmt.Sprintf("Starting machine type change from %s to %s", obj.Status.CurrentMachineType, obj.Spec.MachineType)
		check.Logger().Info("Machine type change initiated", "from", obj.Status.CurrentMachineType, "to", obj.Spec.MachineType)
		return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(2 * time.Second)
	}

	// Step 1: Suspend all workspaces
	if err := r.suspendAllWorkspaces(check.Context(), obj.Name); err != nil {
		obj.Status.MachineTypeChangeMessage = fmt.Sprintf("Failed to suspend workspaces: %v", err)
		return check.Failed(err)
	}
	obj.Status.MachineTypeChangeMessage = "Workspaces suspended"
	check.Logger().Info("All workspaces suspended")

	// Step 2: Deactivate all environments
	if err := r.deactivateAllEnvironments(check.Context(), obj.Name); err != nil {
		obj.Status.MachineTypeChangeMessage = fmt.Sprintf("Failed to deactivate environments: %v", err)
		return check.Failed(err)
	}
	obj.Status.MachineTypeChangeMessage = "Environments deactivated, workspaces suspended"
	check.Logger().Info("All environments deactivated")

	// Step 3: Verify all workspaces are suspended
	workspaceList := &workspacev1.WorkspaceList{}
	if err := r.List(check.Context(), workspaceList); err != nil {
		return check.Failed(fmt.Errorf("failed to list workspaces: %w", err))
	}

	activeWorkspaces := 0
	for _, ws := range workspaceList.Items {
		if ws.Spec.WorkmachineName != obj.Name {
			continue
		}
		if ws.Spec.Status != "suspended" && ws.Spec.Status != "archived" {
			activeWorkspaces++
		}
	}

	if activeWorkspaces > 0 {
		obj.Status.MachineTypeChangeMessage = fmt.Sprintf("Waiting for %d workspaces to suspend", activeWorkspaces)
		return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(5 * time.Second)
	}

	// Step 4: Verify all environments are deactivated
	envList := &environmentV1.EnvironmentList{}
	if err := r.List(check.Context(), envList); err != nil {
		return check.Failed(fmt.Errorf("failed to list environments: %w", err))
	}

	activeEnvironments := 0
	for _, env := range envList.Items {
		if env.Spec.WorkMachineName != obj.Name {
			continue
		}
		if env.Spec.Activated {
			activeEnvironments++
		}
	}

	if activeEnvironments > 0 {
		obj.Status.MachineTypeChangeMessage = fmt.Sprintf("Waiting for %d environments to deactivate", activeEnvironments)
		return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(5 * time.Second)
	}

	// Step 5: Stop the machine
	machineInfo, err := r.cloudProviderAPI.GetMachineStatus(check.Context(), obj.Status.MachineID)
	if err != nil {
		return check.Failed(fmt.Errorf("failed to get machine status: %w", err))
	}

	if machineInfo.State == v1.MachineStateRunning {
		if err := r.cloudProviderAPI.StopMachine(check.Context(), obj.Status.MachineID); err != nil {
			return check.Failed(fmt.Errorf("failed to stop machine: %w", err))
		}
		obj.Status.State = v1.MachineStateStopping
		obj.Status.MachineTypeChangeMessage = "Stopping machine"
		return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(10 * time.Second)
	}

	if machineInfo.State == v1.MachineStateStopping {
		obj.Status.MachineTypeChangeMessage = "Waiting for machine to stop"
		return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(10 * time.Second)
	}

	// Step 6: Change the machine type
	if machineInfo.State == v1.MachineStateStopped {
		if err := r.cloudProviderAPI.ChangeMachine(check.Context(), obj.Status.MachineID, obj.Spec.MachineType); err != nil {
			return check.Failed(fmt.Errorf("failed to change machine type: %w", err))
		}
		obj.Status.MachineTypeChangeMessage = fmt.Sprintf("Machine type changed to %s", obj.Spec.MachineType)
		check.Logger().Info("Machine type changed successfully", "newType", obj.Spec.MachineType)

		// Update current machine type
		obj.Status.CurrentMachineType = obj.Spec.MachineType

		// Step 7: Start the machine
		if err := r.cloudProviderAPI.StartMachine(check.Context(), obj.Status.MachineID); err != nil {
			return check.Failed(fmt.Errorf("failed to start machine after type change: %w", err))
		}
		obj.Status.State = v1.MachineStateStarting
		obj.Status.MachineTypeChangeMessage = "Starting machine with new type"
		return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(10 * time.Second)
	}

	// Step 8: Wait for node to rejoin cluster
	if obj.Status.State == v1.MachineStateStarting {
		node := &corev1.Node{}
		if err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, node); err != nil {
			if apiErrors.IsNotFound(err) {
				obj.Status.MachineTypeChangeMessage = "Waiting for node to rejoin cluster"
				return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(10 * time.Second)
			}
			return check.Failed(fmt.Errorf("failed to get node: %w", err))
		}

		// Check if node is ready
		nodeReady := false
		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady && condition.Status == corev1.ConditionTrue {
				nodeReady = true
				break
			}
		}

		if !nodeReady {
			obj.Status.MachineTypeChangeMessage = "Node joined, waiting for node to be ready"
			return check.UpdateMsg(obj.Status.MachineTypeChangeMessage).RequeueAfter(5 * time.Second)
		}

		// Machine type change complete
		obj.Status.State = v1.MachineStateRunning
		obj.Status.MachineTypeChanging = false
		obj.Status.MachineTypeChangeMessage = fmt.Sprintf("Machine type change complete: %s → %s", obj.Status.CurrentMachineType, obj.Spec.MachineType)
		obj.Status.StartedAt = &metav1.Time{Time: time.Now()}
		check.Logger().Info("Machine type change completed successfully", "newType", obj.Spec.MachineType)
		return check.Passed()
	}

	return check.Passed()
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
			return fmt.Errorf("failed to deactivate environment %s: %w", env.Name, err)
		}
	}
	return nil
}
