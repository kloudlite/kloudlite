package machinescoped

import (
	"context"
	"fmt"

	workmachineshared "github.com/kloudlite/kloudlite/controllers/workmachine/shared"
	fn "github.com/kloudlite/kloudlite/pkg/operator-toolkit/functions"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// createHostManagerRBAC creates RBAC resources for the workmachine-node-manager (host manager pod)
// This service account runs in the workmachine namespace and needs access to:
// - PackageRequests (namespace-scoped, but needs cluster-wide access) - to install Nix packages
// - Snapshots, SnapshotRestores (namespaced) - for snapshot operations
// - Workspaces (namespace-scoped, but needs cluster-wide access) - to manage SSH configuration
// - Nodes (cluster-wide) - to update GPU status
// - Environments (cluster-wide) - for garbage collection of orphaned storage
// - Secrets (in workmachine namespace) - to manage SSH keys
func (r *MachineScopedReconciler) createHostManagerRBAC(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	serviceAccountName := fmt.Sprintf("workmachine-manager-%s", obj.Name)
	serviceAccountNamespace := obj.Spec.TargetNamespace

	// The integrated host-manager runtime now runs in the platform-scoped
	// workmachine-manager pod. Remove the legacy target-namespace service account
	// when present, but don't block reconciliation if it is already gone.
	legacyServiceAccount := &corev1.ServiceAccount{}
	if err := r.Get(ctx, client.ObjectKey{Name: "host-manager", Namespace: obj.Spec.TargetNamespace}, legacyServiceAccount); err != nil {
		if !apiErrors.IsNotFound(err) {
			return markMachineFailed(session, workmachineshared.ConditionHostManagerRBACReady, err)
		}
	} else if fn.IsOwner(legacyServiceAccount, obj) {
		if err := r.Delete(ctx, legacyServiceAccount); err != nil {
			return markMachineFailed(session, workmachineshared.ConditionHostManagerRBACReady, err)
		}
	}

	// Create ClusterRole for host manager
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("hm-%s", obj.Name),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, clusterRole, func() error {
		clusterRole.SetLabels(fn.MapMerge(clusterRole.GetLabels(), map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": "true",
		}))

		clusterRole.Rules = []rbacv1.PolicyRule{
			// PackageRequests (namespace-scoped) - for Nix package management across all namespaces
			{
				APIGroups: []string{"packages.kloudlite.io"},
				Resources: []string{"packagerequests"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{"packages.kloudlite.io"},
				Resources: []string{"packagerequests/status"},
				Verbs:     []string{"get", "update", "patch"},
			},
			// Snapshots - for snapshot operations (cluster-scoped)
			{
				APIGroups: []string{"snapshots.kloudlite.io"},
				Resources: []string{"snapshots", "snapshotstores"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch"},
			},
			{
				APIGroups: []string{"snapshots.kloudlite.io"},
				Resources: []string{"snapshots/status", "snapshotstores/status"},
				Verbs:     []string{"get", "update", "patch"},
			},
			// SnapshotRequests - for creating snapshots
			{
				APIGroups: []string{"snapshots.kloudlite.io"},
				Resources: []string{"snapshotrequests"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{"snapshots.kloudlite.io"},
				Resources: []string{"snapshotrequests/status"},
				Verbs:     []string{"get", "update", "patch"},
			},
			// SnapshotRestores - for restoring snapshots
			{
				APIGroups: []string{"snapshots.kloudlite.io"},
				Resources: []string{"snapshotrestores"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{"snapshots.kloudlite.io"},
				Resources: []string{"snapshotrestores/status"},
				Verbs:     []string{"get", "update", "patch"},
			},
			// Workspaces - for SSH configuration management and directory cleanup
			{
				APIGroups: []string{"workspaces.kloudlite.io"},
				Resources: []string{"workspaces"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			// Nodes - for GPU status updates
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			// Environments - for garbage collection of orphaned storage
			{
				APIGroups: []string{"environments.kloudlite.io"},
				Resources: []string{"environments"},
				Verbs:     []string{"get", "list", "watch"},
			},
		}

		if !fn.IsOwner(clusterRole, obj) {
			clusterRole.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		return nil
	}); err != nil {
		return markMachineFailed(session, workmachineshared.ConditionHostManagerRBACReady, err)
	}

	// Create ClusterRoleBinding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("hm-%s", obj.Name),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, clusterRoleBinding, func() error {
		clusterRoleBinding.SetLabels(fn.MapMerge(clusterRoleBinding.GetLabels(), map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": "true",
		}))

		clusterRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRole.Name,
		}

		clusterRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: serviceAccountNamespace,
			},
		}

		if !fn.IsOwner(clusterRoleBinding, obj) {
			clusterRoleBinding.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		return nil
	}); err != nil {
		return markMachineFailed(session, workmachineshared.ConditionHostManagerRBACReady, err)
	}

	// Create Role in target namespace for Secrets access
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "host-manager",
			Namespace: obj.Spec.TargetNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, role, func() error {
		role.SetLabels(fn.MapMerge(role.GetLabels(), map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": "true",
		}))

		if !fn.IsOwner(role, obj) {
			role.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		role.Rules = []rbacv1.PolicyRule{
			// Secrets - for SSH key management
			{
				APIGroups: []string{""},
				Resources: []string{"secrets"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch"},
			},
		}
		return nil
	}); err != nil {
		return markMachineFailed(session, workmachineshared.ConditionHostManagerRBACReady, err)
	}

	// Create RoleBinding in target namespace
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "host-manager",
			Namespace: obj.Spec.TargetNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, roleBinding, func() error {
		roleBinding.SetLabels(fn.MapMerge(roleBinding.GetLabels(), map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": "true",
		}))

		if !fn.IsOwner(roleBinding, obj) {
			roleBinding.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}

		roleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     role.Name,
		}

		roleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: serviceAccountNamespace,
			},
		}
		return nil
	}); err != nil {
		return markMachineFailed(session, workmachineshared.ConditionHostManagerRBACReady, err)
	}

	return markMachineReady(session, workmachineshared.ConditionHostManagerRBACReady, "host manager RBAC is ready")
}

// createClusterRBAC creates cluster-level RBAC resources
func (r *MachineScopedReconciler) createClusterRBAC(ctx context.Context, session *workmachineshared.StatusSession) (ctrl.Result, error) {
	obj := session.Object()
	namespaceName := obj.Spec.TargetNamespace
	serviceAccountName := obj.Name

	// Create ClusterRole
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("workmachine-%s", obj.Name),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, clusterRole, func() error {
		clusterRole.SetLabels(fn.MapMerge(clusterRole.GetLabels(), map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": "true",
		}))

		clusterRole.Rules = []rbacv1.PolicyRule{
			{
				APIGroups: []string{"workspaces.kloudlite.io"},
				Resources: []string{"workspaces"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"workspaces.kloudlite.io"},
				Resources: []string{"workspaces/status"},
				Verbs:     []string{"get", "update", "patch"},
			},
			{
				APIGroups: []string{"environments.kloudlite.io"},
				Resources: []string{"environments"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"environments.kloudlite.io"},
				Resources: []string{"environments/status"},
				Verbs:     []string{"get", "update", "patch"},
			},
		}

		if !fn.IsOwner(clusterRole, obj) {
			clusterRole.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		return nil
	}); err != nil {
		return markMachineFailed(session, workmachineshared.ConditionHostManagerRBACReady, err)
	}

	// Create ClusterRoleBinding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("workmachine-%s", obj.Name),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, clusterRoleBinding, func() error {
		clusterRoleBinding.SetLabels(fn.MapMerge(clusterRoleBinding.GetLabels(), map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": "true",
		}))

		clusterRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRole.Name,
		}

		clusterRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: namespaceName,
			},
		}

		if !fn.IsOwner(clusterRoleBinding, obj) {
			clusterRoleBinding.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		return nil
	}); err != nil {
		return markMachineFailed(session, workmachineshared.ConditionHostManagerRBACReady, err)
	}

	return markMachineReady(session, workmachineshared.ConditionHostManagerRBACReady, "cluster RBAC is ready")
}
