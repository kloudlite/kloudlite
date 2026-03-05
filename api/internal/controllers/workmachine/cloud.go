package workmachine

import (
	"errors"
	"fmt"
	"time"

	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	klerrors "github.com/kloudlite/kloudlite/api/internal/pkg/errors"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// setupCloudMachine creates, starts, stops, or manages the cloud machine
//
// This function orchestrates the cloud machine lifecycle including:
// - Machine creation (if machine ID doesn't exist)
// - Starting/stopping based on desired state
// - Graceful shutdown with pod drainage
// - Node readiness verification
// - IP address caching via node labels
// - Volume size increases
func (r *WorkMachineReconciler) setupCloudMachine(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	// Early return: Create machine if it doesn't exist
	if obj.Status.MachineID == "" {
		return r.createNewMachine(check, obj)
	}

	// Fetch and cache machine status (with IP caching optimization)
	node, nodeExists, nodeReady := r.fetchNodeState(check, obj)
	machineInfo := r.fetchMachineStatus(check, obj, node, nodeExists, nodeReady)

	// Handle state transitions (start/stop)
	if result := r.handleStateTransitions(check, obj, machineInfo, node); !result.ShouldProceed() {
		return result
	}

	// Handle volume size increases
	if result := r.handleVolumeIncrease(check, obj); !result.ShouldProceed() {
		return result
	}

	// Update WorkMachine status with machine info
	r.updateMachineStatus(obj, machineInfo, int64(fn.ValueOf(obj.Spec.VolumeSize)))

	// Update node IP labels for caching
	r.updateNodeIPLabels(check, obj, node, nodeExists, machineInfo)

	// Verify node readiness for running machines
	return r.verifyNodeReadiness(check, obj, machineInfo, node, nodeExists, nodeReady)
}

// createNewMachine creates a new cloud machine via the cloud provider API
func (r *WorkMachineReconciler) createNewMachine(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
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

	check.Logger().Debug("Using configured cloud machine creation retry interval", "interval", r.Cfg.WorkMachine.CloudMachineCreationRetryInterval)
	return check.UpdateMsg("created cloud machine").RequeueAfter(r.Cfg.WorkMachine.CloudMachineCreationRetryInterval)
}

// fetchNodeState retrieves the Kubernetes node state for the WorkMachine
// Returns the node (or nil if not found), existence flag, and readiness flag
func (r *WorkMachineReconciler) fetchNodeState(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) (*corev1.Node, bool, bool) {
	node := &corev1.Node{}
	nodeExists := false
	nodeReady := false

	if err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, node); err == nil {
		nodeExists = true
		nodeReady = r.isNodeReady(node)
	}

	return node, nodeExists, nodeReady
}

// fetchMachineStatus retrieves machine status, using cached IPs from node labels when possible
// This optimization reduces AWS API calls when node is ready and has cached IPs
func (r *WorkMachineReconciler) fetchMachineStatus(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine, node *corev1.Node, nodeExists bool, nodeReady bool) *v1.MachineInfo {
	// Check if we can use cached IPs from node labels
	canUseCache := nodeExists && nodeReady && node.Labels != nil &&
		node.Labels[NodeLabelPublicIP] != "" &&
		node.Labels[NodeLabelPrivateIP] != ""

	if canUseCache {
		machineInfo := &v1.MachineInfo{
			MachineID:        obj.Status.MachineID,
			State:            v1.MachineStateRunning,
			PublicIP:         node.Labels[NodeLabelPublicIP],
			PrivateIP:        node.Labels[NodeLabelPrivateIP],
			AvailabilityZone: obj.Status.AvailabilityZone,
			Region:           obj.Status.Region,
			Message:          "Node is ready (using cached IPs)",
		}
		check.Logger().Debug("using cached IPs from node labels",
			"publicIP", machineInfo.PublicIP,
			"privateIP", machineInfo.PrivateIP)
		return machineInfo
	}

	// Fetch fresh IPs from cloud provider API
	machineInfo, err := r.cloudProviderAPI.GetMachineStatus(check.Context(), obj.Status.MachineID)
	if err != nil {
		check.Failed(err)
		return nil
	}

	check.Logger().Debug("fetched fresh IPs from cloud provider API",
		"publicIP", machineInfo.PublicIP,
		"privateIP", machineInfo.PrivateIP)

	return machineInfo
}

// handleStateTransitions handles starting or stopping the machine based on desired state
func (r *WorkMachineReconciler) handleStateTransitions(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine, machineInfo *v1.MachineInfo, node *corev1.Node) reconciler.StepResult {
	desiredState := obj.Spec.State
	currentState := machineInfo.State

	// Start machine if desired state is running but machine is stopped
	if desiredState == v1.MachineStateRunning && currentState == v1.MachineStateStopped {
		return r.startMachine(check, obj)
	}

	// Stop machine if desired state is stopped but machine is running
	if desiredState == v1.MachineStateStopped && currentState == v1.MachineStateRunning {
		return r.stopMachineGracefully(check, obj, node)
	}

	// Check if machine is transitioning to desired state
	if currentState != desiredState {
		return check.UpdateMsg("waiting for machine status to change").RequeueAfter(r.Cfg.WorkMachine.MachineStatusCheckInterval)
	}

	return check.Passed()
}

// startMachine starts a stopped cloud machine
func (r *WorkMachineReconciler) startMachine(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	if err := r.cloudProviderAPI.StartMachine(check.Context(), obj.Status.MachineID); err != nil {
		return check.Failed(fmt.Errorf("failed to start machine: %w", err))
	}

	obj.Status.State = v1.MachineStateStarting
	obj.Status.StartedAt = &metav1.Time{Time: time.Now()}
	return check.UpdateMsg("Starting machine").RequeueAfter(r.Cfg.WorkMachine.CloudMachineStartRetryInterval)
}

// stopMachineGracefully performs a graceful shutdown by suspending workspaces,
// deactivating environments, cordoning the node, and draining pods
func (r *WorkMachineReconciler) stopMachineGracefully(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine, node *corev1.Node) reconciler.StepResult {
	// Step 1: Suspend all workspaces
	if err := r.suspendAllWorkspaces(check.Context(), obj.Name); err != nil {
		return check.Failed(fmt.Errorf("failed to suspend workspaces before stopping machine: %w", err))
	}

	// Step 2: Deactivate all environments
	if err := r.deactivateAllEnvironments(check.Context(), obj.Name); err != nil {
		return check.Failed(fmt.Errorf("failed to deactivate environments before stopping machine: %w", err))
	}

	// Step 3: Cordon the node (prevent new pods from being scheduled)
	if err := r.cordonNode(check, obj); err != nil {
		return err
	}

	// Step 4: Evict all pods on this node (drain)
	if result := r.drainNode(check, obj); !result.ShouldProceed() {
		return result
	}

	// Step 5: Delete the Kubernetes node object to prevent metrics-server timeouts
	if err := r.deleteNodeObject(check, obj); err != nil {
		return err
	}

	// Step 6: All pods have terminated, now safe to stop the VM
	if err := r.cloudProviderAPI.StopMachine(check.Context(), obj.Status.MachineID); err != nil {
		return check.Failed(fmt.Errorf("failed to stop machine: %w", err))
	}

	obj.Status.State = v1.MachineStateStopping
	obj.Status.StoppedAt = &metav1.Time{Time: time.Now()}
	return check.UpdateMsg("Stopping Machine (all pods terminated gracefully)").RequeueAfter(r.Cfg.WorkMachine.CloudMachineStopRetryInterval)
}

// cordonNode marks the node as unschedulable to prevent new pods
func (r *WorkMachineReconciler) cordonNode(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	node := &corev1.Node{}
	if err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, node); err != nil {
		if apiErrors.IsNotFound(err) {
			// Node doesn't exist, skip cordoning
			return check.Passed()
		}
		return check.Failed(fmt.Errorf("failed to get node for cordoning: %w", err))
	}

	if !node.Spec.Unschedulable {
		node.Spec.Unschedulable = true
		if err := r.Update(check.Context(), node); err != nil {
			return check.Failed(fmt.Errorf("failed to cordon node: %w", err))
		}
		check.Logger().Info("cordoned node", "node", obj.Name)
	}

	return check.Passed()
}

// drainNode evicts all pods from the node
// Returns failure if any pod eviction fails, ensuring graceful shutdown integrity
func (r *WorkMachineReconciler) drainNode(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	podList := &corev1.PodList{}
	if err := r.List(check.Context(), podList, client.MatchingFields{"spec.nodeName": obj.Name}); err != nil {
		return check.Failed(fmt.Errorf("failed to list pods on node: %w", err))
	}

	if len(podList.Items) == 0 {
		return check.Passed()
	}

	check.Logger().Info("draining node", "node", obj.Name, "podCount", len(podList.Items))

	evictedCount := 0
	var evictErrors []error

	for i := range podList.Items {
		pod := &podList.Items[i]
		// Skip pods that are already terminating
		if pod.DeletionTimestamp != nil {
			continue
		}
		// Delete pod to trigger graceful termination
		if err := r.Delete(check.Context(), pod); err != nil {
			if !apiErrors.IsNotFound(err) {
				evictErrors = append(evictErrors, fmt.Errorf("failed to evict pod %s/%s: %w", pod.Namespace, pod.Name, err))
			}
		} else {
			evictedCount++
		}
	}

	// If any eviction failed, aggregate and return errors
	if len(evictErrors) > 0 {
		return check.Failed(fmt.Errorf("failed to evict %d pod(s): %v", len(evictErrors), errors.Join(evictErrors...)))
	}

	if evictedCount > 0 {
		// Requeue to wait for pods to terminate gracefully
		return check.UpdateMsg("Draining node (waiting for pods to terminate gracefully)").RequeueAfter(r.Cfg.WorkMachine.NodeDrainRetryInterval)
	}

	return check.Passed()
}

// deleteNodeObject deletes the Kubernetes node object
// This is a critical cleanup step that must succeed to prevent orphaned nodes
func (r *WorkMachineReconciler) deleteNodeObject(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	node := &corev1.Node{}
	if err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, node); err != nil {
		if apiErrors.IsNotFound(err) {
			return check.Passed()
		}
		return check.Failed(fmt.Errorf("failed to get node for deletion: %w", err))
	}

	if err := r.Delete(check.Context(), node); err != nil {
		if apiErrors.IsNotFound(err) {
			return check.Passed()
		}
		return check.Failed(fmt.Errorf("failed to delete node %s: %w", obj.Name, err))
	}

	check.Logger().Info("deleted node object", "node", obj.Name)
	return check.Passed()
}

// handleVolumeIncrease handles increasing the storage volume size if requested
func (r *WorkMachineReconciler) handleVolumeIncrease(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	specVolume := fn.ValueOf(obj.Spec.VolumeSize)

	if specVolume <= obj.Status.StorageVolumeSize {
		return check.Passed()
	}

	check.Logger().Info("increasing storage volume size",
		"from", obj.Status.StorageVolumeSize,
		"to", obj.Spec.VolumeSize)

	if err := r.cloudProviderAPI.IncreaseVolumeSize(check.Context(), obj.Status.MachineID, specVolume); err != nil {
		return check.Failed(klerrors.Wrap("failed to increase storage volume size", err))
	}

	obj.Status.StorageVolumeSize = specVolume
	if err := r.cloudProviderAPI.RebootMachine(check.Context(), obj.Status.MachineID); err != nil {
		return check.Errored(klerrors.Wrap(fmt.Sprintf("failed to reboot machine(ID: %s)", obj.Status.MachineID), err))
	}

	return check.UpdateMsg("waiting for volume size to be increased").RequeueAfter(r.Cfg.WorkMachine.VolumeResizeCheckInterval)
}

// updateMachineStatus updates the WorkMachine status with machine info
func (r *WorkMachineReconciler) updateMachineStatus(obj *v1.WorkMachine, machineInfo *v1.MachineInfo, volumeSize int64) {
	obj.Status.PublicIP = machineInfo.PublicIP
	obj.Status.PrivateIP = machineInfo.PrivateIP
	obj.Status.StorageVolumeSize = int32(volumeSize)
	obj.Status.Message = machineInfo.Message
	obj.Status.AvailabilityZone = machineInfo.AvailabilityZone
	obj.Status.Region = machineInfo.Region
}

// updateNodeIPLabels updates node labels with IP addresses for caching
func (r *WorkMachineReconciler) updateNodeIPLabels(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine, node *corev1.Node, nodeExists bool, machineInfo *v1.MachineInfo) {
	if !nodeExists || machineInfo.PublicIP == "" || machineInfo.PrivateIP == "" {
		return
	}

	needsUpdate := false
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}

	// Update labels only if changed
	if node.Labels[NodeLabelPublicIP] != machineInfo.PublicIP {
		node.Labels[NodeLabelPublicIP] = machineInfo.PublicIP
		needsUpdate = true
	}
	if node.Labels[NodeLabelPrivateIP] != machineInfo.PrivateIP {
		node.Labels[NodeLabelPrivateIP] = machineInfo.PrivateIP
		needsUpdate = true
	}

	if needsUpdate {
		if err := r.Update(check.Context(), node); err != nil {
			check.Logger().Warn("failed to update node IP labels", "error", err)
			// Don't fail reconciliation for label updates
		} else {
			check.Logger().Info("updated node IP labels",
				"publicIP", machineInfo.PublicIP,
				"privateIP", machineInfo.PrivateIP)
		}
	}
}

// verifyNodeReadiness verifies the node has joined and is ready before marking machine as running
func (r *WorkMachineReconciler) verifyNodeReadiness(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine, machineInfo *v1.MachineInfo, node *corev1.Node, nodeExists bool, nodeReady bool) reconciler.StepResult {
	// For non-running states, use cloud provider state directly
	if machineInfo.State != v1.MachineStateRunning {
		obj.Status.State = machineInfo.State
		return check.Passed()
	}

	// Reuse node if already fetched, otherwise fetch it
	if !nodeExists {
		var err error
		node, nodeExists, nodeReady = r.fetchNodeState(check, obj)
		if err != nil {
			return check.Failed(fmt.Errorf("failed to get node: %w", err))
		}
	}

	// Node hasn't joined yet
	if !nodeExists {
		obj.Status.State = v1.MachineStateStarting
		obj.Status.Message = "Waiting for node to join cluster"
		return check.UpdateMsg("waiting for node to join cluster").RequeueAfter(r.Cfg.WorkMachine.NodeJoinCheckInterval)
	}

	// Node joined but not ready yet
	if !nodeReady {
		r.clearNodeIPLabels(check, node)
		obj.Status.State = v1.MachineStateStarting
		obj.Status.Message = "Node joined, waiting for node to be ready"
		return check.UpdateMsg("waiting for node to be ready").RequeueAfter(r.Cfg.WorkMachine.NodeReadyRetryInterval)
	}

	// Uncordon the node if it was previously cordoned (e.g., during shutdown)
	if node.Spec.Unschedulable {
		if err := r.uncordonNode(check, obj, node); err != nil {
			return err
		}
	}

	// Node is ready, mark as running
	obj.Status.State = v1.MachineStateRunning
	obj.Status.Message = "Node is ready"
	return check.Passed()
}

// clearNodeIPLabels removes IP labels from a node that is not ready
// This forces a fresh lookup on next reconciliation
func (r *WorkMachineReconciler) clearNodeIPLabels(check *reconciler.Check[*v1.WorkMachine], node *corev1.Node) {
	if node.Labels == nil || (node.Labels[NodeLabelPublicIP] == "" && node.Labels[NodeLabelPrivateIP] == "") {
		return
	}

	delete(node.Labels, NodeLabelPublicIP)
	delete(node.Labels, NodeLabelPrivateIP)

	if err := r.Update(check.Context(), node); err != nil {
		check.Logger().Warn("failed to remove IP labels from not-ready node", "error", err)
	} else {
		check.Logger().Info("removed IP labels from not-ready node")
	}
}

// uncordonNode marks a node as schedulable
func (r *WorkMachineReconciler) uncordonNode(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine, node *corev1.Node) reconciler.StepResult {
	node.Spec.Unschedulable = false
	if err := r.Update(check.Context(), node); err != nil {
		return check.Failed(fmt.Errorf("failed to uncordon node: %w", err))
	}
	check.Logger().Info("uncordoned node", "node", obj.Name)
	return check.Passed()
}

// cleanupCloudMachine handles cloud machine deletion
//
// The cleanup process follows these steps:
// 1. Add NoExecute taint to node to evict pods
// 2. Force delete any remaining pods on node
// 3. Delete Kubernetes Node object
// 4. Delete cloud machine via cloud provider API
//
// This is a critical cleanup operation that must succeed to prevent resource leaks
func (r *WorkMachineReconciler) cleanupCloudMachine(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	if obj.Status.MachineID == "" {
		return check.Passed()
	}

	// Step 1: Add NoExecute taint to evict pods
	if result := r.addDeletionTaint(check, obj); !result.ShouldProceed() {
		return result
	}

	// Step 2: Force delete remaining pods
	if result := r.forceDeletePods(check, obj); !result.ShouldProceed() {
		return result
	}

	// Step 3: Delete Kubernetes node object
	if result := r.deleteKubernetesNode(check, obj); !result.ShouldProceed() {
		return result
	}

	// Step 4: Delete cloud machine
	return r.deleteCloudMachine(check, obj)
}

// addDeletionTaint adds a NoExecute taint to the node to trigger pod eviction
func (r *WorkMachineReconciler) addDeletionTaint(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	node := &corev1.Node{}

	if err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, node); err != nil {
		if apiErrors.IsNotFound(err) {
			// Node doesn't exist, skip tainting and proceed
			check.Logger().Info("node not found during cleanup, proceeding with deletion", "node", obj.Name)
			return check.Passed()
		}
		return check.Failed(fmt.Errorf("failed to get node for cleanup: %w", err))
	}

	// Check if taint already exists
	if r.hasDeletionTaint(node) {
		return check.Passed()
	}

	// Add NoExecute taint
	node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{
		Key:    "kloudlite.io/workmachine-deleting",
		Value:  "true",
		Effect: corev1.TaintEffectNoExecute,
	})

	if err := r.Update(check.Context(), node); err != nil {
		return check.Failed(fmt.Errorf("failed to add NoExecute taint to node %s: %w", obj.Name, err))
	}

	check.Logger().Info("added NoExecute taint to node, waiting for pod eviction")
	return check.UpdateMsg("Added NoExecute taint to node").RequeueAfter(r.Cfg.WorkMachine.NodeDeleteRetryInterval)
}

// hasDeletionTaint checks if the node already has the deletion taint
func (r *WorkMachineReconciler) hasDeletionTaint(node *corev1.Node) bool {
	for _, taint := range node.Spec.Taints {
		if taint.Key == "kloudlite.io/workmachine-deleting" && taint.Effect == corev1.TaintEffectNoExecute {
			return true
		}
	}
	return false
}

// forceDeletePods force deletes all pods on the node
func (r *WorkMachineReconciler) forceDeletePods(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	podList := &corev1.PodList{}
	if err := r.List(check.Context(), podList, client.MatchingFields{"spec.nodeName": obj.Name}); err != nil {
		return check.Failed(fmt.Errorf("failed to list pods on node: %w", err))
	}

	if len(podList.Items) == 0 {
		return check.Passed()
	}

	gracePeriod := int64(0)
	deleteOptions := &client.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	}

	var podDeleteErrors []error
	for i := range podList.Items {
		pod := &podList.Items[i]
		check.Logger().Info("force deleting pod", "pod", pod.Name, "namespace", pod.Namespace)
		if err := r.Delete(check.Context(), pod, deleteOptions); err != nil {
			if !apiErrors.IsNotFound(err) {
				podDeleteErrors = append(podDeleteErrors, fmt.Errorf("failed to delete pod %s/%s: %w", pod.Namespace, pod.Name, err))
			}
		}
	}

	// If any pod deletion failed, return aggregated error
	if len(podDeleteErrors) > 0 {
		return check.Failed(fmt.Errorf("failed to delete %d pod(s): %v", len(podDeleteErrors), errors.Join(podDeleteErrors...)))
	}

	return check.UpdateMsg("Waiting for pods to be deleted").RequeueAfter(r.Cfg.WorkMachine.NodeDeleteRetryInterval)
}

// deleteKubernetesNode deletes the Kubernetes Node object
func (r *WorkMachineReconciler) deleteKubernetesNode(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	node := &corev1.Node{}
	if err := r.Get(check.Context(), client.ObjectKey{Name: obj.Name}, node); err != nil {
		if apiErrors.IsNotFound(err) {
			// Node doesn't exist, nothing to delete
			return check.Passed()
		}
		return check.Failed(fmt.Errorf("failed to get Kubernetes node for deletion: %w", err))
	}

	if err := r.Delete(check.Context(), node); err != nil {
		if !apiErrors.IsNotFound(err) {
			return check.Failed(fmt.Errorf("failed to delete Kubernetes node %s: %w", obj.Name, err))
		}
	}

	check.Logger().Info("deleted node object during cleanup", "node", obj.Name)
	return check.Passed()
}

// deleteCloudMachine deletes the cloud machine via the cloud provider API
func (r *WorkMachineReconciler) deleteCloudMachine(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	if err := r.cloudProviderAPI.DeleteMachine(check.Context(), obj.Status.MachineID); err != nil {
		return check.Failed(fmt.Errorf("failed to delete cloud machine %s: %w", obj.Status.MachineID, err))
	}

	obj.Status.MachineInfo = v1.MachineInfo{}
	check.Logger().Info("successfully deleted cloud machine", "machineID", obj.Status.MachineID)
	return check.Passed()
}
