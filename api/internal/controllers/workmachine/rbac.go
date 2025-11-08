package workmachine

import (
	"fmt"

	v1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	fn "github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/functions"
	"github.com/kloudlite/kloudlite/api/pkg/operator-toolkit/reconciler"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// createHostManagerRBAC creates RBAC resources for the workmachine-node-manager (host manager pod)
// This service account runs in the kloudlite-hostmanager namespace and needs access to:
// - PackageRequests (cluster-wide) - to install Nix packages
// - Workspaces (cluster-wide) - to manage SSH configuration
// - Nodes (cluster-wide) - to update GPU status
// - Secrets (in hostmanager namespace) - to manage SSH keys
func (r *WorkMachineReconciler) createHostManagerRBAC(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	const hostManagerNamespace = "kloudlite-hostmanager"
	const serviceAccountName = "workmachine-node-manager"

	// Create ClusterRole for host manager
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("hm-%s", obj.Name),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, clusterRole, func() error {
		clusterRole.SetLabels(fn.MapMerge(clusterRole.GetLabels(), map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": "true",
		}))

		clusterRole.Rules = []rbacv1.PolicyRule{
			// PackageRequests - for Nix package management
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
			// Workspaces - for SSH configuration management
			{
				APIGroups: []string{"workspaces.kloudlite.io"},
				Resources: []string{"workspaces"},
				Verbs:     []string{"get", "list", "watch"},
			},
			// Nodes - for GPU status updates
			{
				APIGroups: []string{""},
				Resources: []string{"nodes"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
		}

		if !fn.IsOwner(clusterRole, obj) {
			clusterRole.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	// Create ClusterRoleBinding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("hm-%s", obj.Name),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, clusterRoleBinding, func() error {
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
				Namespace: hostManagerNamespace,
			},
		}

		if !fn.IsOwner(clusterRoleBinding, obj) {
			clusterRoleBinding.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	// Create Role in hostmanager namespace for Secrets access
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("hm-%s", obj.Name),
			Namespace: hostManagerNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, role, func() error {
		role.SetLabels(fn.MapMerge(role.GetLabels(), map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": "true",
		}))

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
		return check.Failed(err)
	}

	// Create RoleBinding in hostmanager namespace
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("hm-%s", obj.Name),
			Namespace: hostManagerNamespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, roleBinding, func() error {
		roleBinding.SetLabels(fn.MapMerge(roleBinding.GetLabels(), map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": "true",
		}))

		roleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     role.Name,
		}

		roleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: hostManagerNamespace,
			},
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// createRBACInNamespace creates Role and RoleBinding in the target namespace
func (r *WorkMachineReconciler) createRBACInNamespace(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespaceName := obj.Spec.TargetNamespace
	serviceAccountName := obj.Name

	// Create ServiceAccount
	sa := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: serviceAccountName, Namespace: namespaceName}}
	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, sa, func() error {
		sa.SetLabels(fn.MapMerge(sa.GetLabels(), map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": "true",
		}))
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	// Create Role
	role := &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: namespaceName}}
	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, role, func() error {
		role.SetLabels(fn.MapMerge(role.GetLabels(), map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": "true",
		}))

		role.Rules = []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"pods", "pods/log", "pods/exec"},
				Verbs:     []string{"get", "list", "watch", "create", "delete"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"secrets", "configmaps"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	// Create RoleBinding
	roleBinding := &rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: obj.Name, Namespace: namespaceName}}
	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, roleBinding, func() error {
		roleBinding.SetLabels(fn.MapMerge(roleBinding.GetLabels(), map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workmachine": "true",
		}))

		roleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     role.Name,
		}

		roleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: namespaceName,
			},
		}
		return nil
	}); err != nil {
		return check.Failed(err)
	}

	return check.Passed()
}

// createClusterRBAC creates cluster-level RBAC resources
func (r *WorkMachineReconciler) createClusterRBAC(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine) reconciler.StepResult {
	namespaceName := obj.Spec.TargetNamespace
	serviceAccountName := obj.Name

	// Create ClusterRole
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("workmachine-%s", obj.Name),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, clusterRole, func() error {
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
		return check.Failed(err)
	}

	// Create ClusterRoleBinding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("workmachine-%s", obj.Name),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, clusterRoleBinding, func() error {
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
		return check.Failed(err)
	}

	return check.Passed()
}

// createWorkspaceRBACForWorkspace creates workspace-specific RBAC resources
// This is a helper function that can be called by other controllers (like Workspace controller)
func (r *WorkMachineReconciler) createWorkspaceRBACForWorkspace(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine, workspaceName string) error {
	namespaceName := obj.Spec.TargetNamespace
	serviceAccountName := fmt.Sprintf("workspace-%s", workspaceName)

	// Create ServiceAccount
	sa := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: serviceAccountName, Namespace: namespaceName}}
	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, sa, func() error {
		sa.SetLabels(fn.MapMerge(sa.GetLabels(), map[string]string{
			"kloudlite.io/managed":     "true",
			"kloudlite.io/workspace":   workspaceName,
			"kloudlite.io/workmachine": obj.Name,
		}))
		return nil
	}); err != nil {
		return err
	}

	// Create ClusterRole for workspace
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("ws-%s", workspaceName),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, clusterRole, func() error {
		clusterRole.SetLabels(fn.MapMerge(clusterRole.GetLabels(), map[string]string{
			"kloudlite.io/managed":   "true",
			"kloudlite.io/workspace": workspaceName,
		}))

		clusterRole.Rules = []rbacv1.PolicyRule{
			{
				APIGroups:     []string{"workspaces.kloudlite.io"},
				Resources:     []string{"workspaces"},
				ResourceNames: []string{workspaceName},
				Verbs:         []string{"get", "list", "watch", "update", "patch"},
			},
			{
				APIGroups: []string{"environments.kloudlite.io"},
				Resources: []string{"environments"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"get", "list"},
			},
			{
				APIGroups: []string{"intercepts.kloudlite.io"},
				Resources: []string{"serviceintercepts"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"intercepts.kloudlite.io"},
				Resources: []string{"serviceintercepts/status"},
				Verbs:     []string{"get"},
			},
			{
				APIGroups: []string{"packages.kloudlite.io"},
				Resources: []string{"packagerequests"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				APIGroups: []string{"packages.kloudlite.io"},
				Resources: []string{"packagerequests/status"},
				Verbs:     []string{"get"},
			},
		}

		if !fn.IsOwner(clusterRole, obj) {
			clusterRole.SetOwnerReferences([]metav1.OwnerReference{fn.AsOwner(obj, true)})
		}
		return nil
	}); err != nil {
		return err
	}

	// Create ClusterRoleBinding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("ws-%s", workspaceName),
		},
	}

	if _, err := controllerutil.CreateOrUpdate(check.Context(), r.Client, clusterRoleBinding, func() error {
		clusterRoleBinding.SetLabels(fn.MapMerge(clusterRoleBinding.GetLabels(), map[string]string{
			"kloudlite.io/managed":   "true",
			"kloudlite.io/workspace": workspaceName,
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
		return err
	}

	return nil
}

// deleteWorkspaceRBACForWorkspace deletes workspace-specific RBAC resources
// This is a helper function that can be called by other controllers (like Workspace controller)
func (r *WorkMachineReconciler) deleteWorkspaceRBACForWorkspace(check *reconciler.Check[*v1.WorkMachine], obj *v1.WorkMachine, workspaceName string) error {
	namespaceName := obj.Spec.TargetNamespace

	// Delete ServiceAccount
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("workspace-%s", workspaceName),
			Namespace: namespaceName,
		},
	}
	if err := r.Delete(check.Context(), sa); err != nil && !apiErrors.IsNotFound(err) {
		return err
	}

	// Delete ClusterRole
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("ws-%s", workspaceName),
		},
	}
	if err := r.Delete(check.Context(), clusterRole); err != nil && !apiErrors.IsNotFound(err) {
		return err
	}

	// Delete ClusterRoleBinding
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("ws-%s", workspaceName),
		},
	}
	if err := r.Delete(check.Context(), clusterRoleBinding); err != nil && !apiErrors.IsNotFound(err) {
		return err
	}

	return nil
}
