package machinescoped

import (
	"context"
	"fmt"

	workmachineshared "github.com/kloudlite/kloudlite/controllers/workmachine/shared"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	tunnelServerName = "tunnel-server"

	// Wildcard TLS certificate containing tls.crt, tls.key, and ca.crt
	// This secret is synced to each workmachine namespace
	kloudliteWildcardCertName = "kloudlite-wildcard-cert-tls"
)

// ensureTunnelServer is a no-op because the tunnel server now runs inside the workmachine-manager process.
// It only cleans up any legacy resources that may still exist from before the integration.
func (r *MachineScopedReconciler) ensureTunnelServer(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	result, err := r.cleanupTunnelServer(ctx, session)
	if err != nil || !isZeroMachineResult(result) {
		return result, err
	}
	return markMachineReady(session, workmachineshared.ConditionTunnelServerReady, "tunnel server runs inside workmachine-manager")
}

// cleanupTunnelServer deletes the tunnel-server StatefulSet, service, and RBAC resources
func (r *MachineScopedReconciler) cleanupTunnelServer(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	namespace := obj.Spec.TargetNamespace
	clusterRoleName := fmt.Sprintf("%s-%s", tunnelServerName, obj.Name)
	clusterRoleBindingName := fmt.Sprintf("%s-%s", tunnelServerName, obj.Name)

	// Delete StatefulSet if it exists
	if err := r.Delete(ctx, &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tunnelServerName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionTunnelServerReady, fmt.Errorf("failed to delete tunnel-server statefulset: %w", err))
		}
	}

	// Delete service
	if err := r.Delete(ctx, &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tunnelServerName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionTunnelServerReady, fmt.Errorf("failed to delete tunnel-server service: %w", err))
		}
	}

	// Delete ClusterRoleBinding
	if err := r.Delete(ctx, &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleBindingName,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionTunnelServerReady, fmt.Errorf("failed to delete tunnel-server cluster role binding: %w", err))
		}
	}

	// Delete ClusterRole
	if err := r.Delete(ctx, &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionTunnelServerReady, fmt.Errorf("failed to delete tunnel-server cluster role: %w", err))
		}
	}

	// Delete ServiceAccount
	if err := r.Delete(ctx, &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tunnelServerName,
			Namespace: namespace,
		},
	}); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionTunnelServerReady, fmt.Errorf("failed to delete tunnel-server service account: %w", err))
		}
	}

	return markMachineReady(session, workmachineshared.ConditionTunnelServerReady, "tunnel server is cleaned up")
}
