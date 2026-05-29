package machinescoped

import (
	"context"
	"fmt"

	workmachineshared "github.com/kloudlite/kloudlite/controllers/workmachine/shared"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *MachineScopedReconciler) ensureIntegratedHostManager(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	if r.Client == nil {
		return markMachineReady(session, workmachineshared.ConditionHostManagerReady, "host manager runs inside workmachine-manager")
	}
	result, err := r.cleanupHostManagerPod(ctx, session)
	if err != nil || !isZeroMachineResult(result) {
		return result, err
	}
	return markMachineReady(session, workmachineshared.ConditionHostManagerReady, "host manager runs inside workmachine-manager")
}

// ensureHostManagerPod ensures the workmachine-host-manager StatefulSet exists
// ensureHostManagerPod is a no-op - host-manager runs inside WM
func (r *MachineScopedReconciler) ensureHostManagerPod(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	return markMachineReady(session, workmachineshared.ConditionHostManagerReady, "host manager runs inside workmachine-manager")
}

// cleanupHostManagerPod deletes the host-manager StatefulSet and service
// This function is called when the WorkMachine is not in running state
func (r *MachineScopedReconciler) cleanupHostManagerPod(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	namespace := obj.Spec.TargetNamespace
	hostManagerName := "host-manager"

	// Delete StatefulSet if it exists (this will cascade delete pods)
	if err := r.Delete(ctx, &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostManagerName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionHostManagerReady, fmt.Errorf("failed to delete host-manager statefulset: %w", err))
		}
	}

	// Delete service
	if err := r.Delete(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      hostManagerName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionHostManagerReady, fmt.Errorf("failed to delete host-manager service: %w", err))
		}
	}

	return markMachineReady(session, workmachineshared.ConditionHostManagerReady, "host manager is cleaned up")
}
