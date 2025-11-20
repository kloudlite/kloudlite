package workmachine

import (
	"fmt"
	"time"

	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	errors "github.com/kloudlite/kloudlite/api/internal/pkg/errors"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// setupCloudMachine creates, starts, stops, or manages the cloud machine
func (r *WorkMachineReconciler) setupCloudMachine(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	if obj.Status.MachineID == "" {
		mi, err := r.cloudProviderAPI.CreateMachine(check.Context(), obj)
		if err != nil {
			return check.Failed(err)
		}

		obj.Status.StartedAt = &metav1.Time{Time: time.Now()}
		obj.Status.MachineInfo = *mi
		// Set state to starting since node hasn't joined yet
		if obj.Spec.State == v1.MachineStateRunning {
			obj.Status.State = v1.MachineStateStarting
			obj.Status.Message = "Cloud machine created, waiting for node to join"
		}
		return check.UpdateMsg("created cloud machine").RequeueAfter(2 * time.Second)
	}

	machineInfo, err := r.cloudProviderAPI.GetMachineStatus(check.Context(), obj.Status.MachineID)
	if err != nil {
		return check.Failed(err)
	}

	// Handle desired state transitions
	currentState := machineInfo.State

	// Start machine if desired state is running but machine is stopped
	if obj.Spec.State == v1.MachineStateRunning && currentState == v1.MachineStateStopped {
		if err := r.cloudProviderAPI.StartMachine(check.Context(), obj.Status.MachineID); err != nil {
			return check.Failed(fmt.Errorf("failed to start machine: %w", err))
		}
		obj.Status.State = v1.MachineStateStarting
		obj.Status.StartedAt = &metav1.Time{Time: time.Now()}
		return check.UpdateMsg("Starting machine").RequeueAfter(10 * time.Second)
	}

	// Stop machine if desired state is stopped but machine is running
	if obj.Spec.State == v1.MachineStateStopped && currentState == v1.MachineStateRunning {
		// Step 1: Suspend all workspaces
		if err := r.suspendAllWorkspaces(check.Context(), obj.Name); err != nil {
			return check.Failed(fmt.Errorf("failed to suspend workspaces before stopping machine: %w", err))
		}

		// Step 2: Deactivate all environments
		if err := r.deactivateAllEnvironments(check.Context(), obj.Name); err != nil {
			return check.Failed(fmt.Errorf("failed to deactivate environments before stopping machine: %w", err))
		}

		// Step 3: Cordon the node (prevent new pods from being scheduled)
		node := &corev1.Node{}
		if err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, node); err != nil {
			if !apiErrors.IsNotFound(err) {
				return check.Failed(fmt.Errorf("failed to get node for cordoning: %w", err))
			}
			// Node doesn't exist, skip cordoning
		} else {
			if !node.Spec.Unschedulable {
				node.Spec.Unschedulable = true
				if err := r.Update(check.Context(), node); err != nil {
					return check.Failed(fmt.Errorf("failed to cordon node: %w", err))
				}
				check.Logger().Info("cordoned node", "node", obj.Name)
			}
		}

		// Step 4: Evict all pods on this node (drain)
		podList := &corev1.PodList{}
		if err := r.List(check.Context(), podList, client.MatchingFields{"spec.nodeName": obj.Name}); err != nil {
			return check.Failed(fmt.Errorf("failed to list pods on node: %w", err))
		}

		if len(podList.Items) > 0 {
			check.Logger().Info("draining node", "node", obj.Name, "podCount", len(podList.Items))
			for i := range podList.Items {
				pod := &podList.Items[i]
				// Skip pods that are already terminating
				if pod.DeletionTimestamp != nil {
					continue
				}
				// Delete pod to trigger graceful termination
				if err := r.Delete(check.Context(), pod); err != nil && !apiErrors.IsNotFound(err) {
					check.Logger().Warn("failed to evict pod", "pod", pod.Name, "namespace", pod.Namespace, "error", err)
				}
			}
			// Requeue to wait for pods to terminate gracefully
			return check.UpdateMsg("Draining node (waiting for pods to terminate gracefully)").RequeueAfter(5 * time.Second)
		}

		// Step 5: All pods have terminated, now safe to stop the VM
		if err := r.cloudProviderAPI.StopMachine(check.Context(), obj.Status.MachineID); err != nil {
			return check.Failed(fmt.Errorf("failed to stop machine: %w", err))
		}

		obj.Status.State = v1.MachineStateStopping
		obj.Status.StoppedAt = &metav1.Time{Time: time.Now()}
		return check.UpdateMsg("Stopping Machine (all pods terminated gracefully)").RequeueAfter(10 * time.Second)
	}

	// Check if machine state matches desired state (but we'll verify node readiness below)
	if currentState != obj.Spec.State {
		// Machine is transitioning
		return check.UpdateMsg("waiting for machine status to change").RequeueAfter(5 * time.Second)
	}

	specVolume := fn.ValueOf(obj.Spec.VolumeSize)

	if specVolume > obj.Status.RootVolumeSize {
		check.Logger().Info("increasing volume size", "from", obj.Status.RootVolumeSize, "to", obj.Spec.VolumeSize)

		if err := r.cloudProviderAPI.IncreaseVolumeSize(check.Context(), obj.Status.MachineID, specVolume); err != nil {
			return check.Failed(errors.Wrap("failed to increase volume size", err))
		}

		obj.Status.RootVolumeSize = specVolume
		if err := r.cloudProviderAPI.RebootMachine(check.Context(), obj.Status.MachineID); err != nil {
			return check.Errored(errors.Wrap(fmt.Sprintf("failed to reboot machine(ID: %s)", obj.Status.MachineID), err))
		}

		return check.UpdateMsg("waiting for volume size to be increased").RequeueAfter(10 * time.Second)
	}

	// Update status with current machine info
	obj.Status.PublicIP = machineInfo.PublicIP
	obj.Status.PrivateIP = machineInfo.PrivateIP
	obj.Status.RootVolumeSize = specVolume
	obj.Status.Message = machineInfo.Message

	// Check if node has joined the cluster before marking as running
	if machineInfo.State == v1.MachineStateRunning {
		node := &corev1.Node{}
		if err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, node); err != nil {
			if apiErrors.IsNotFound(err) {
				// Node hasn't joined yet, keep state as "starting"
				obj.Status.State = v1.MachineStateStarting
				obj.Status.Message = "Waiting for node to join cluster"
				return check.UpdateMsg("waiting for node to join cluster").RequeueAfter(10 * time.Second)
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
			obj.Status.State = v1.MachineStateStarting
			obj.Status.Message = "Node joined, waiting for node to be ready"
			return check.UpdateMsg("waiting for node to be ready").RequeueAfter(5 * time.Second)
		}

		// Uncordon the node if it was previously cordoned (e.g., during shutdown)
		if node.Spec.Unschedulable {
			node.Spec.Unschedulable = false
			if err := r.Update(check.Context(), node); err != nil {
				return check.Failed(fmt.Errorf("failed to uncordon node: %w", err))
			}
			check.Logger().Info("uncordoned node", "node", obj.Name)
		}

		// Node is ready, mark as running
		obj.Status.State = v1.MachineStateRunning
		obj.Status.Message = "Node is ready"
	} else {
		// For other states (stopped, stopping, etc.), use the cloud provider state directly
		obj.Status.State = machineInfo.State
	}

	return check.Passed()
}

// cleanupCloudMachine handles cloud machine deletion
func (r *WorkMachineReconciler) cleanupCloudMachine(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	if obj.Status.MachineID == "" {
		return check.Passed()
	}

	// Step 1: Add NoExecute taint to the node to evict pods
	node := &corev1.Node{}
	nodeExists := false
	if err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, node); err != nil {
		if !apiErrors.IsNotFound(err) {
			check.Logger().Warn("failed to get node for tainting", "error", err)
		}
		// Node doesn't exist or failed to get, skip tainting
	} else {
		nodeExists = true
		// Add NoExecute taint if not already present
		taintExists := false
		for _, taint := range node.Spec.Taints {
			if taint.Key == "kloudlite.io/workmachine-deleting" && taint.Effect == corev1.TaintEffectNoExecute {
				taintExists = true
				break
			}
		}

		if !taintExists {
			node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{
				Key:    "kloudlite.io/workmachine-deleting",
				Value:  "true",
				Effect: corev1.TaintEffectNoExecute,
			})
			if err := r.Update(check.Context(), node); err != nil {
				check.Logger().Warn("failed to add NoExecute taint to node", "error", err)
			} else {
				check.Logger().Info("added NoExecute taint to node, waiting for pod eviction")
				return check.UpdateMsg("Added NoExecute taint to node").RequeueAfter(2 * time.Second)
			}
		}
	}

	// Step 2: Force delete any remaining pods on this node
	podList := &corev1.PodList{}
	if err := r.List(check.Context(), podList, client.MatchingFields{"spec.nodeName": obj.Name}); err != nil {
		return check.Failed(fmt.Errorf("failed to list pods on node: %w", err))
	}

	if len(podList.Items) > 0 {
		gracePeriod := int64(0)
		deleteOptions := &client.DeleteOptions{
			GracePeriodSeconds: &gracePeriod,
		}
		for i := range podList.Items {
			pod := &podList.Items[i]
			check.Logger().Info("force deleting pod", "pod", pod.Name, "namespace", pod.Namespace)
			if err := r.Delete(check.Context(), pod, deleteOptions); err != nil && !apiErrors.IsNotFound(err) {
				check.Logger().Warn("failed to force delete pod", "pod", pod.Name, "namespace", pod.Namespace, "error", err)
			}
		}
		// Wait for pods to be deleted
		return check.UpdateMsg("Waiting for pods to be deleted").RequeueAfter(2 * time.Second)
	}

	// Step 3: Delete the Kubernetes Node object (only if it exists)
	if nodeExists {
		if err := r.Delete(check.Context(), node); err != nil {
			if !apiErrors.IsNotFound(err) {
				return check.Failed(fmt.Errorf("failed to delete Kubernetes node: %w", err))
			}
		}
	}

	// Step 4: Delete the cloud machine (EC2 instance)
	if err := r.cloudProviderAPI.DeleteMachine(check.Context(), obj.Status.MachineID); err != nil {
		return check.Failed(fmt.Errorf("failed to delete AWS machine: %w", err))
	}

	obj.Status.MachineInfo = v1.MachineInfo{}
	return check.Passed()
}
