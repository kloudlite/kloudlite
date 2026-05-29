package platformscoped

import (
	"context"
	"errors"
	"fmt"
	"time"

	workmachineshared "github.com/kloudlite/kloudlite/controllers/workmachine/shared"
	klerrors "github.com/kloudlite/kloudlite/pkg/errors"
	fn "github.com/kloudlite/kloudlite/pkg/operator-toolkit/functions"
	v1 "github.com/kloudlite/kloudlite/types/workmachine/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
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
func (r *PlatformScopedReconciler) setupCloudMachine(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	// Early return: Create machine if it doesn't exist
	if obj.Status.MachineID == "" {
		return r.createNewMachine(ctx, session)
	}

	// Fetch and cache machine status (with IP caching optimization)
	node, nodeExists, nodeReady := r.fetchNodeState(ctx, obj)
	machineInfo, err := r.fetchMachineStatus(ctx, session, node, nodeExists, nodeReady)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Handle state transitions (start/stop)
	if result, err := r.handleStateTransitions(ctx, session, machineInfo, node); err != nil || !isZeroResult(result) {
		return result, err
	}

	// Handle volume size increases
	if result, err := r.handleVolumeIncrease(ctx, session); err != nil || !isZeroResult(result) {
		return result, err
	}

	// Update WorkMachine status with machine info
	r.updateMachineStatus(obj, machineInfo, int64(fn.ValueOf(obj.Spec.VolumeSize)))

	// Update node IP labels for caching
	r.updateNodeIPLabels(ctx, obj, node, nodeExists, machineInfo)

	// Verify node readiness for running machines
	return r.verifyNodeReadiness(ctx, session, machineInfo, node, nodeExists, nodeReady)
}

// createNewMachine creates a new cloud machine via the cloud provider API
func (r *PlatformScopedReconciler) createNewMachine(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	mi, err := r.cloudProviderAPI.CreateMachine(ctx, obj)
	if err != nil {
		return markError(session, workmachineshared.ConditionCloudMachineProvisioned, err)
	}

	obj.Status.StartedAt = &metav1.Time{Time: time.Now()}
	obj.Status.MachineInfo = *mi

	// Set state to starting since node hasn't joined yet
	if obj.Spec.State == v1.MachineStateRunning {
		obj.Status.State = v1.MachineStateStarting
		obj.Status.Message = "Cloud machine created, waiting for node to join"
	}

	ctrl.LoggerFrom(ctx).V(1).Info("using configured cloud machine creation retry interval", "interval", r.Cfg.WorkMachine.CloudMachineCreationRetryInterval)
	session.MarkTrue(workmachineshared.ConditionCloudMachineProvisioned, workmachineshared.ReasonReconciled, "cloud machine created")
	return markBlocked(session, workmachineshared.ConditionCloudMachineRunning, workmachineshared.ReasonWaitingForNodeJoin, "created cloud machine", r.Cfg.WorkMachine.CloudMachineCreationRetryInterval)
}

// fetchNodeState retrieves the Kubernetes node state for the WorkMachine
// Returns the node (or nil if not found), existence flag, and readiness flag
func (r *PlatformScopedReconciler) fetchNodeState(ctx context.Context, obj *v1.WorkMachine) (*corev1.Node, bool, bool) {
	node := &corev1.Node{}
	nodeExists := false
	nodeReady := false

	if err := r.Get(ctx, client.ObjectKey{Name: obj.Name}, node); err == nil {
		nodeExists = true
		nodeReady = r.isNodeReady(node)
	}

	return node, nodeExists, nodeReady
}

// fetchMachineStatus retrieves machine status, using cached IPs from node labels when possible
// This optimization reduces AWS API calls when node is ready and has cached IPs
func (r *PlatformScopedReconciler) fetchMachineStatus(ctx context.Context, session *workmachineshared.StatusSession, node *corev1.Node, nodeExists bool, nodeReady bool) (*v1.MachineInfo, error) {
	obj := session.Object()
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
		ctrl.LoggerFrom(ctx).V(1).Info("using cached IPs from node labels",
			"publicIP", machineInfo.PublicIP,
			"privateIP", machineInfo.PrivateIP)
		return machineInfo, nil
	}

	// Fetch fresh IPs from cloud provider API
	machineInfo, err := r.cloudProviderAPI.GetMachineStatus(ctx, obj.Status.MachineID)
	if err != nil {
		_, _ = markError(session, workmachineshared.ConditionCloudMachineProvisioned, err)
		return nil, err
	}

	ctrl.LoggerFrom(ctx).V(1).Info("fetched fresh IPs from cloud provider API",
		"publicIP", machineInfo.PublicIP,
		"privateIP", machineInfo.PrivateIP)

	return machineInfo, nil
}

// handleStateTransitions handles starting or stopping the machine based on desired state
func (r *PlatformScopedReconciler) handleStateTransitions(ctx context.Context, session *workmachineshared.StatusSession, machineInfo *v1.MachineInfo, node *corev1.Node) (ctrl.Result, error) {
	obj := session.Object()
	desiredState := obj.Spec.State
	currentState := machineInfo.State

	// Start machine if desired state is running but machine is stopped
	if desiredState == v1.MachineStateRunning && currentState == v1.MachineStateStopped {
		return r.startMachine(ctx, session)
	}

	// Stop machine if desired state is stopped but machine is running
	if desiredState == v1.MachineStateStopped && currentState == v1.MachineStateRunning {
		return r.stopMachineGracefully(ctx, session, node)
	}

	// Check if machine is transitioning to desired state
	if currentState != desiredState {
		return markBlocked(session, workmachineshared.ConditionCloudMachineRunning, workmachineshared.ReasonWaitingForCloudMachine, "waiting for machine status to change", r.Cfg.WorkMachine.MachineStatusCheckInterval)
	}

	return ctrl.Result{}, nil
}

// startMachine starts a stopped cloud machine
func (r *PlatformScopedReconciler) startMachine(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	if err := r.cloudProviderAPI.StartMachine(ctx, obj.Status.MachineID); err != nil {
		return markError(session, workmachineshared.ConditionCloudMachineRunning, fmt.Errorf("failed to start machine: %w", err))
	}

	r.usageReporter.ReportEvent(ctx, UsageEvent{
		EventType:    "workmachine.started",
		ResourceID:   obj.Status.MachineID,
		ResourceType: "workmachine." + obj.Spec.MachineType,
		Timestamp:    time.Now(),
	})

	obj.Status.State = v1.MachineStateStarting
	obj.Status.StartedAt = &metav1.Time{Time: time.Now()}
	return markBlocked(session, workmachineshared.ConditionCloudMachineRunning, workmachineshared.ReasonWaitingForCloudMachine, "Starting machine", r.Cfg.WorkMachine.CloudMachineStartRetryInterval)
}

// stopMachineGracefully performs a graceful shutdown by suspending workspaces,
// deactivating environments, cordoning the node, and draining pods
func (r *PlatformScopedReconciler) stopMachineGracefully(ctx context.Context, session *workmachineshared.StatusSession, node *corev1.Node) (ctrl.Result, error) {
	obj := session.Object()
	if !machineScopedWorkloadsReady(obj) {
		return markBlocked(session, workmachineshared.ConditionCloudMachineRunning, workmachineshared.ReasonMachineWorkloadsNotReady, "waiting for machine-scoped workloads before stopping machine", r.Cfg.WorkMachine.MachineStatusCheckInterval)
	}

	// Step 1: Suspend all workspaces
	if err := r.suspendAllWorkspaces(ctx, obj.Name); err != nil {
		return markError(session, workmachineshared.ConditionCloudMachineRunning, fmt.Errorf("failed to suspend workspaces before stopping machine: %w", err))
	}

	// Step 2: Deactivate all environments
	if err := r.deactivateAllEnvironments(ctx, obj.Name); err != nil {
		return markError(session, workmachineshared.ConditionCloudMachineRunning, fmt.Errorf("failed to deactivate environments before stopping machine: %w", err))
	}

	// Step 3: Cordon the node (prevent new pods from being scheduled)
	if result, err := r.cordonNode(ctx, session); err != nil || !isZeroResult(result) {
		return result, err
	}

	// Step 4: Evict all pods on this node (drain)
	if result, err := r.drainNode(ctx, session); err != nil || !isZeroResult(result) {
		return result, err
	}

	// Step 5: Delete the Kubernetes node object to prevent metrics-server timeouts
	if result, err := r.deleteNodeObject(ctx, session); err != nil || !isZeroResult(result) {
		return result, err
	}

	// Step 6: All pods have terminated, now safe to stop the VM
	if err := r.cloudProviderAPI.StopMachine(ctx, obj.Status.MachineID); err != nil {
		return markError(session, workmachineshared.ConditionCloudMachineRunning, fmt.Errorf("failed to stop machine: %w", err))
	}

	r.usageReporter.ReportEvent(ctx, UsageEvent{
		EventType:    "workmachine.stopped",
		ResourceID:   obj.Status.MachineID,
		ResourceType: "workmachine." + obj.Spec.MachineType,
		Timestamp:    time.Now(),
	})

	obj.Status.State = v1.MachineStateStopping
	obj.Status.StoppedAt = &metav1.Time{Time: time.Now()}
	return markBlocked(session, workmachineshared.ConditionCloudMachineRunning, workmachineshared.ReasonWaitingForCloudMachine, "Stopping Machine (all pods terminated gracefully)", r.Cfg.WorkMachine.CloudMachineStopRetryInterval)
}

// cordonNode marks the node as unschedulable to prevent new pods
func (r *PlatformScopedReconciler) cordonNode(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	node := &corev1.Node{}
	if err := r.Get(ctx, client.ObjectKey{Name: obj.Name}, node); err != nil {
		if apiErrors.IsNotFound(err) {
			// Node doesn't exist, skip cordoning
			return ctrl.Result{}, nil
		}
		return markError(session, workmachineshared.ConditionCloudMachineRunning, fmt.Errorf("failed to get node for cordoning: %w", err))
	}

	if !node.Spec.Unschedulable {
		node.Spec.Unschedulable = true
		if err := r.Update(ctx, node); err != nil {
			return markError(session, workmachineshared.ConditionCloudMachineRunning, fmt.Errorf("failed to cordon node: %w", err))
		}
		ctrl.LoggerFrom(ctx).Info("cordoned node", "node", obj.Name)
	}

	return ctrl.Result{}, nil
}

// drainNode evicts all pods from the node
// Returns failure if any pod eviction fails, ensuring graceful shutdown integrity
func (r *PlatformScopedReconciler) drainNode(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	podList := &corev1.PodList{}
	if err := r.List(ctx, podList, client.MatchingFields{"spec.nodeName": obj.Name}); err != nil {
		return markError(session, workmachineshared.ConditionCloudMachineRunning, fmt.Errorf("failed to list pods on node: %w", err))
	}

	if len(podList.Items) == 0 {
		return ctrl.Result{}, nil
	}

	ctrl.LoggerFrom(ctx).Info("draining node", "node", obj.Name, "podCount", len(podList.Items))

	evictedCount := 0
	var evictErrors []error

	for i := range podList.Items {
		pod := &podList.Items[i]
		// Skip pods that are already terminating
		if pod.DeletionTimestamp != nil {
			continue
		}
		// Delete pod to trigger graceful termination
		if err := r.Delete(ctx, pod); err != nil {
			if !apiErrors.IsNotFound(err) {
				evictErrors = append(evictErrors, fmt.Errorf("failed to evict pod %s/%s: %w", pod.Namespace, pod.Name, err))
			}
		} else {
			evictedCount++
		}
	}

	// If any eviction failed, aggregate and return errors
	if len(evictErrors) > 0 {
		return markError(session, workmachineshared.ConditionCloudMachineRunning, fmt.Errorf("failed to evict %d pod(s): %v", len(evictErrors), errors.Join(evictErrors...)))
	}

	if evictedCount > 0 {
		// Requeue to wait for pods to terminate gracefully
		return markBlocked(session, workmachineshared.ConditionCloudMachineRunning, workmachineshared.ReasonWaitingForCloudMachine, "Draining node (waiting for pods to terminate gracefully)", r.Cfg.WorkMachine.NodeDrainRetryInterval)
	}

	return ctrl.Result{}, nil
}

// deleteNodeObject deletes the Kubernetes node object
// This is a critical cleanup step that must succeed to prevent orphaned nodes
func (r *PlatformScopedReconciler) deleteNodeObject(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	node := &corev1.Node{}
	if err := r.Get(ctx, client.ObjectKey{Name: obj.Name}, node); err != nil {
		if apiErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return markError(session, workmachineshared.ConditionCloudMachineRunning, fmt.Errorf("failed to get node for deletion: %w", err))
	}

	if err := r.Delete(ctx, node); err != nil {
		if apiErrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return markError(session, workmachineshared.ConditionCloudMachineRunning, fmt.Errorf("failed to delete node %s: %w", obj.Name, err))
	}

	ctrl.LoggerFrom(ctx).Info("deleted node object", "node", obj.Name)
	return ctrl.Result{}, nil
}

// handleVolumeIncrease handles increasing the storage volume size if requested
func (r *PlatformScopedReconciler) handleVolumeIncrease(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	specVolume := fn.ValueOf(obj.Spec.VolumeSize)

	if specVolume <= obj.Status.StorageVolumeSize {
		return ctrl.Result{}, nil
	}

	ctrl.LoggerFrom(ctx).Info("increasing storage volume size",
		"from", obj.Status.StorageVolumeSize,
		"to", obj.Spec.VolumeSize)

	if err := r.cloudProviderAPI.IncreaseVolumeSize(ctx, obj.Status.MachineID, specVolume); err != nil {
		return markError(session, workmachineshared.ConditionCloudMachineProvisioned, klerrors.Wrap("failed to increase storage volume size", err))
	}

	obj.Status.StorageVolumeSize = specVolume
	if err := r.cloudProviderAPI.RebootMachine(ctx, obj.Status.MachineID); err != nil {
		return markError(session, workmachineshared.ConditionCloudMachineProvisioned, klerrors.Wrap(fmt.Sprintf("failed to reboot machine(ID: %s)", obj.Status.MachineID), err))
	}

	return markBlocked(session, workmachineshared.ConditionCloudMachineRunning, workmachineshared.ReasonWaitingForCloudMachine, "waiting for volume size to be increased", r.Cfg.WorkMachine.VolumeResizeCheckInterval)
}

// updateMachineStatus updates the WorkMachine status with machine info
func (r *PlatformScopedReconciler) updateMachineStatus(obj *v1.WorkMachine, machineInfo *v1.MachineInfo, volumeSize int64) {
	obj.Status.PublicIP = machineInfo.PublicIP
	obj.Status.PrivateIP = machineInfo.PrivateIP
	obj.Status.StorageVolumeSize = int32(volumeSize)
	obj.Status.Message = machineInfo.Message
	obj.Status.AvailabilityZone = machineInfo.AvailabilityZone
	obj.Status.Region = machineInfo.Region
}

// updateNodeIPLabels updates node labels with IP addresses for caching
func (r *PlatformScopedReconciler) updateNodeIPLabels(ctx context.Context, obj *v1.WorkMachine, node *corev1.Node, nodeExists bool, machineInfo *v1.MachineInfo) {
	if !nodeExists || machineInfo.PublicIP == "" || machineInfo.PrivateIP == "" {
		return
	}

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nodeForUpdate := &corev1.Node{}
		if err := r.Get(ctx, client.ObjectKey{Name: node.Name}, nodeForUpdate); err != nil {
			return err
		}

		needsUpdate := false
		if nodeForUpdate.Labels == nil {
			nodeForUpdate.Labels = make(map[string]string)
		}
		if nodeForUpdate.Labels[NodeLabelPublicIP] != machineInfo.PublicIP {
			nodeForUpdate.Labels[NodeLabelPublicIP] = machineInfo.PublicIP
			needsUpdate = true
		}
		if nodeForUpdate.Labels[NodeLabelPrivateIP] != machineInfo.PrivateIP {
			nodeForUpdate.Labels[NodeLabelPrivateIP] = machineInfo.PrivateIP
			needsUpdate = true
		}
		if !needsUpdate {
			return nil
		}
		return r.Update(ctx, nodeForUpdate)
	}); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed to update node IP labels")
	} else {
		ctrl.LoggerFrom(ctx).Info("updated node IP labels",
			"publicIP", machineInfo.PublicIP,
			"privateIP", machineInfo.PrivateIP)
	}
}

// verifyNodeReadiness verifies the node has joined and is ready before marking machine as running
func (r *PlatformScopedReconciler) verifyNodeReadiness(ctx context.Context, session *workmachineshared.StatusSession, machineInfo *v1.MachineInfo, node *corev1.Node, nodeExists bool, nodeReady bool) (ctrl.Result, error) {
	obj := session.Object()
	session.MarkTrue(workmachineshared.ConditionCloudMachineProvisioned, workmachineshared.ReasonReconciled, "cloud machine is provisioned")
	// For non-running states, use cloud provider state directly
	if machineInfo.State != v1.MachineStateRunning {
		obj.Status.State = machineInfo.State
		return markReady(session, workmachineshared.ConditionCloudMachineRunning, "cloud machine reached desired non-running state")
	}

	// Reuse node if already fetched, otherwise fetch it
	if !nodeExists {
		node, nodeExists, nodeReady = r.fetchNodeState(ctx, obj)
	}

	// Node hasn't joined yet
	if !nodeExists {
		obj.Status.State = v1.MachineStateStarting
		obj.Status.Message = "Waiting for node to join cluster"
		return markBlocked(session, workmachineshared.ConditionNodeJoined, workmachineshared.ReasonWaitingForNodeJoin, "waiting for node to join cluster", r.Cfg.WorkMachine.NodeJoinCheckInterval)
	}

	// Verify the found node actually belongs to this workmachine by checking its
	// private IP matches the cloud VM's private IP. This prevents the controller
	// from using a stale or mislabeled control-plane node.
	if machineInfo.PrivateIP != "" {
		nodeIP := r.getNodeInternalIP(node)
		if nodeIP != machineInfo.PrivateIP {
			ctrl.LoggerFrom(ctx).Info("found node with mismatched IP, waiting for correct node to join",
				"node", node.Name,
				"nodeIP", nodeIP,
				"expectedPrivateIP", machineInfo.PrivateIP,
			)
			obj.Status.State = v1.MachineStateStarting
			obj.Status.Message = "Node IP mismatch, waiting for correct node"
			return markBlocked(session, workmachineshared.ConditionNodeJoined, workmachineshared.ReasonWaitingForNodeJoin, "node IP mismatch, waiting for correct node", r.Cfg.WorkMachine.NodeJoinCheckInterval)
		}
	}

	session.MarkTrue(workmachineshared.ConditionNodeJoined, workmachineshared.ReasonReconciled, "node joined cluster")

	// Node joined but not ready yet
	if !nodeReady {
		r.clearNodeIPLabels(ctx, node)
		obj.Status.State = v1.MachineStateStarting
		obj.Status.Message = "Node joined, waiting for node to be ready"
		return markBlocked(session, workmachineshared.ConditionCloudMachineRunning, workmachineshared.ReasonWaitingForNodeJoin, "waiting for node to be ready", r.Cfg.WorkMachine.NodeReadyRetryInterval)
	}

	// Uncordon the node if it was previously cordoned (e.g., during shutdown)
	if node.Spec.Unschedulable {
		if result, err := r.uncordonNode(ctx, session, node); err != nil || !isZeroResult(result) {
			return result, err
		}
	}

	// Node is ready, mark as running
	obj.Status.State = v1.MachineStateRunning
	obj.Status.Message = "Node is ready"
	return markReady(session, workmachineshared.ConditionCloudMachineRunning, "node is ready")
}

// clearNodeIPLabels removes IP labels from a node that is not ready
// This forces a fresh lookup on next reconciliation
func (r *PlatformScopedReconciler) clearNodeIPLabels(ctx context.Context, node *corev1.Node) {
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		nodeForUpdate := &corev1.Node{}
		if err := r.Get(ctx, client.ObjectKey{Name: node.Name}, nodeForUpdate); err != nil {
			return err
		}
		if nodeForUpdate.Labels == nil || (nodeForUpdate.Labels[NodeLabelPublicIP] == "" && nodeForUpdate.Labels[NodeLabelPrivateIP] == "") {
			return nil
		}
		delete(nodeForUpdate.Labels, NodeLabelPublicIP)
		delete(nodeForUpdate.Labels, NodeLabelPrivateIP)
		return r.Update(ctx, nodeForUpdate)
	}); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed to remove IP labels from not-ready node")
	} else {
		ctrl.LoggerFrom(ctx).Info("removed IP labels from not-ready node")
	}
}

// uncordonNode marks a node as schedulable
func (r *PlatformScopedReconciler) uncordonNode(ctx context.Context, session *workmachineshared.StatusSession, node *corev1.Node) (ctrl.Result, error) {
	obj := session.Object()
	node.Spec.Unschedulable = false
	if err := r.Update(ctx, node); err != nil {
		return markError(session, workmachineshared.ConditionCloudMachineRunning, fmt.Errorf("failed to uncordon node: %w", err))
	}
	ctrl.LoggerFrom(ctx).Info("uncordoned node", "node", obj.Name)
	return ctrl.Result{}, nil
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
func (r *PlatformScopedReconciler) cleanupCloudMachine(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	if obj.Status.MachineID == "" {
		return markReady(session, workmachineshared.ConditionCloudMachineDeleted, "cloud machine is already deleted")
	}

	// Step 1: Add NoExecute taint to evict pods
	if result, err := r.addDeletionTaint(ctx, session); err != nil || !isZeroResult(result) {
		return result, err
	}

	// Step 2: Force delete remaining pods
	if result, err := r.forceDeletePods(ctx, session); err != nil || !isZeroResult(result) {
		return result, err
	}

	// Step 3: Delete Kubernetes node object
	if result, err := r.deleteKubernetesNode(ctx, session); err != nil || !isZeroResult(result) {
		return result, err
	}

	// Step 4: Delete cloud machine
	return r.deleteCloudMachine(ctx, session)
}

// addDeletionTaint adds a NoExecute taint to the node to trigger pod eviction
func (r *PlatformScopedReconciler) addDeletionTaint(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	node := &corev1.Node{}

	if err := r.Get(ctx, client.ObjectKey{Name: obj.Name}, node); err != nil {
		if apiErrors.IsNotFound(err) {
			// Node doesn't exist, skip tainting and proceed
			ctrl.LoggerFrom(ctx).Info("node not found during cleanup, proceeding with deletion", "node", obj.Name)
			return ctrl.Result{}, nil
		}
		return markError(session, workmachineshared.ConditionCloudMachineDeleted, fmt.Errorf("failed to get node for cleanup: %w", err))
	}

	// Check if taint already exists
	if r.hasDeletionTaint(node) {
		return ctrl.Result{}, nil
	}

	// Add NoExecute taint
	node.Spec.Taints = append(node.Spec.Taints, corev1.Taint{
		Key:    "kloudlite.io/workmachine-deleting",
		Value:  "true",
		Effect: corev1.TaintEffectNoExecute,
	})

	if err := r.Update(ctx, node); err != nil {
		return markError(session, workmachineshared.ConditionCloudMachineDeleted, fmt.Errorf("failed to add NoExecute taint to node %s: %w", obj.Name, err))
	}

	ctrl.LoggerFrom(ctx).Info("added NoExecute taint to node, waiting for pod eviction")
	return markBlocked(session, workmachineshared.ConditionCloudMachineDeleted, workmachineshared.ReasonDeleting, "Added NoExecute taint to node", r.Cfg.WorkMachine.NodeDeleteRetryInterval)
}

// hasDeletionTaint checks if the node already has the deletion taint
func (r *PlatformScopedReconciler) hasDeletionTaint(node *corev1.Node) bool {
	for _, taint := range node.Spec.Taints {
		if taint.Key == "kloudlite.io/workmachine-deleting" && taint.Effect == corev1.TaintEffectNoExecute {
			return true
		}
	}
	return false
}

// forceDeletePods force deletes all pods on the node
func (r *PlatformScopedReconciler) forceDeletePods(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	podList := &corev1.PodList{}
	if err := r.List(ctx, podList, client.MatchingFields{"spec.nodeName": obj.Name}); err != nil {
		return markError(session, workmachineshared.ConditionCloudMachineDeleted, fmt.Errorf("failed to list pods on node: %w", err))
	}

	if len(podList.Items) == 0 {
		return ctrl.Result{}, nil
	}

	gracePeriod := int64(0)
	deleteOptions := &client.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	}

	var podDeleteErrors []error
	for i := range podList.Items {
		pod := &podList.Items[i]
		ctrl.LoggerFrom(ctx).Info("force deleting pod", "pod", pod.Name, "namespace", pod.Namespace)
		if err := r.Delete(ctx, pod, deleteOptions); err != nil {
			if !apiErrors.IsNotFound(err) {
				podDeleteErrors = append(podDeleteErrors, fmt.Errorf("failed to delete pod %s/%s: %w", pod.Namespace, pod.Name, err))
			}
		}
	}

	// If any pod deletion failed, return aggregated error
	if len(podDeleteErrors) > 0 {
		return markError(session, workmachineshared.ConditionCloudMachineDeleted, fmt.Errorf("failed to delete %d pod(s): %v", len(podDeleteErrors), errors.Join(podDeleteErrors...)))
	}

	return markBlocked(session, workmachineshared.ConditionCloudMachineDeleted, workmachineshared.ReasonDeleting, "Waiting for pods to be deleted", r.Cfg.WorkMachine.NodeDeleteRetryInterval)
}

// deleteKubernetesNode deletes the Kubernetes Node object
func (r *PlatformScopedReconciler) deleteKubernetesNode(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	node := &corev1.Node{}
	if err := r.Get(ctx, client.ObjectKey{Name: obj.Name}, node); err != nil {
		if apiErrors.IsNotFound(err) {
			// Node doesn't exist, nothing to delete
			return ctrl.Result{}, nil
		}
		return markError(session, workmachineshared.ConditionCloudMachineDeleted, fmt.Errorf("failed to get Kubernetes node for deletion: %w", err))
	}

	if err := r.Delete(ctx, node); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markError(session, workmachineshared.ConditionCloudMachineDeleted, fmt.Errorf("failed to delete Kubernetes node %s: %w", obj.Name, err))
		}
	}

	ctrl.LoggerFrom(ctx).Info("deleted node object during cleanup", "node", obj.Name)
	return ctrl.Result{}, nil
}

// deleteCloudMachine deletes the cloud machine via the cloud provider API
func (r *PlatformScopedReconciler) deleteCloudMachine(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	machineID := obj.Status.MachineID

	if err := r.cloudProviderAPI.DeleteMachine(ctx, machineID); err != nil {
		return markError(session, workmachineshared.ConditionCloudMachineDeleted, fmt.Errorf("failed to delete cloud machine %s: %w", machineID, err))
	}

	r.usageReporter.ReportEvent(ctx, UsageEvent{
		EventType:    "workmachine.stopped",
		ResourceID:   machineID,
		ResourceType: "workmachine." + obj.Spec.MachineType,
		Timestamp:    time.Now(),
	})

	obj.Status.MachineInfo = v1.MachineInfo{}
	ctrl.LoggerFrom(ctx).Info("successfully deleted cloud machine", "machineID", machineID)
	return markReady(session, workmachineshared.ConditionCloudMachineDeleted, "cloud machine deleted")
}

// getNodeInternalIP returns the InternalIP address from the node's status addresses.
func (r *PlatformScopedReconciler) getNodeInternalIP(node *corev1.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			return addr.Address
		}
	}
	return ""
}
