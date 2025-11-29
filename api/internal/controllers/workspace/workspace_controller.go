package workspace

import (
	"context"
	"fmt"
	"time"

	packagesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/packages/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	workspaceFinalizer        = "workspaces.kloudlite.io/finalizer"
	workspaceCleanupFinalizer = "workspaces.kloudlite.io/directory-cleanup"

	// Default idle timeout if not specified in workspace settings (30 minutes)
	defaultIdleTimeoutMinutes = 30
)

// WorkspaceReconciler reconciles Workspace objects and manages VS Code server pods
type WorkspaceReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Logger    *zap.Logger
	Config    *rest.Config
	Clientset *kubernetes.Clientset
}

// Reconcile handles Workspace events and ensures the workspace pod exists
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("workspace", req.Name),
	)

	logger.Info("Reconciling Workspace")

	// Fetch the Workspace instance (namespaced)
	workspace := &workspacev1.Workspace{}
	err := r.Get(ctx, client.ObjectKey{Name: req.Name, Namespace: req.Namespace}, workspace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Workspace not found, likely deleted")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get Workspace", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Check if workspace is being deleted
	if workspace.DeletionTimestamp != nil {
		logger.Info("Workspace is being deleted, starting cleanup")
		return r.handleDeletion(ctx, workspace, logger)
	}

	// Add finalizers if not present
	finalizersAdded := false
	if !controllerutil.ContainsFinalizer(workspace, workspaceFinalizer) {
		logger.Info("Adding workspace finalizer")
		controllerutil.AddFinalizer(workspace, workspaceFinalizer)
		finalizersAdded = true
	}
	if !controllerutil.ContainsFinalizer(workspace, workspaceCleanupFinalizer) {
		logger.Info("Adding directory cleanup finalizer")
		controllerutil.AddFinalizer(workspace, workspaceCleanupFinalizer)
		finalizersAdded = true
	}
	if finalizersAdded {
		if err := r.Update(ctx, workspace); err != nil {
			logger.Error("Failed to add finalizers", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Set up workspace-specific RBAC
	if err := r.setupWorkspaceRBAC(ctx, workspace, logger); err != nil {
		logger.Error("Failed to setup workspace RBAC", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Set WorkMachine as owner if WorkmachineName is specified and owner reference not yet set
	if workspace.Spec.WorkmachineName != "" {
		needsOwnerUpdate := true
		for _, ownerRef := range workspace.OwnerReferences {
			if ownerRef.Kind == "WorkMachine" && ownerRef.Name == workspace.Spec.WorkmachineName {
				needsOwnerUpdate = false
				break
			}
		}

		if needsOwnerUpdate {
			logger.Info("Setting WorkMachine as owner of Workspace",
				zap.String("workmachine", workspace.Spec.WorkmachineName))

			// Fetch WorkMachine to set as owner
			workmachine, err := r.getWorkMachine(ctx, workspace.Spec.WorkmachineName)
			if err != nil {
				logger.Error("Failed to get WorkMachine for ownership",
					zap.String("workmachine", workspace.Spec.WorkmachineName),
					zap.Error(err))
				// Don't fail reconciliation, just log the error
				// The ownership will be set on next reconciliation
			} else {
				// Set WorkMachine as owner for cascading deletion (without blockOwnerDeletion)
				blockOwnerDeletion := false
				ownerRef := metav1.OwnerReference{
					APIVersion:         workmachine.APIVersion,
					Kind:               workmachine.Kind,
					Name:               workmachine.Name,
					UID:                workmachine.UID,
					BlockOwnerDeletion: &blockOwnerDeletion,
				}
				workspace.SetOwnerReferences([]metav1.OwnerReference{ownerRef})

				if err := r.Update(ctx, workspace); err != nil {
					logger.Error("Failed to update Workspace with owner reference", zap.Error(err))
					return reconcile.Result{}, err
				}
				logger.Info("Successfully set WorkMachine as owner of Workspace")
				return reconcile.Result{Requeue: true}, nil
			}
		}
	}

	// Set default VS Code version if not provided
	if workspace.Spec.VSCodeVersion == "" {
		workspace.Spec.VSCodeVersion = "latest"
	}

	// Handle workspace cloning if copyFrom is set
	// Cloning takes precedence over normal workspace reconciliation
	if workspace.Spec.CopyFrom != "" {
		logger.Info("Workspace has copyFrom set, handling cloning",
			zap.String("sourceWorkspace", workspace.Spec.CopyFrom))
		return r.handleCloning(ctx, workspace, logger)
	}

	// Handle workspace based on its status
	var result reconcile.Result

	switch workspace.Spec.Status {
	case "active":
		result, err = r.handleActiveWorkspace(ctx, workspace, logger)
	case "suspended", "archived":
		result, err = r.handleSuspendedWorkspace(ctx, workspace, logger)
	default:
		// Default to active if status is not set
		workspace.Spec.Status = "active"
		if err := r.Update(ctx, workspace); err != nil {
			logger.Error("Failed to update workspace status", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Requeue after 1 minute to check idle status periodically
	if workspace.Spec.Status == "active" && workspace.Spec.Settings != nil && workspace.Spec.Settings.AutoStop {
		if result.RequeueAfter == 0 && !result.Requeue {
			result.RequeueAfter = 1 * time.Minute
		}
	}

	return result, err
}

// setupWorkspaceRBAC creates workspace-specific ClusterRole and ClusterRoleBinding
// to restrict workspace users to only access their own workspace and intercepts
func (r *WorkspaceReconciler) setupWorkspaceRBAC(ctx context.Context, workspace *workspacev1.Workspace, logger *zap.Logger) error {
	// Get the workspace namespace from WorkMachine
	if workspace.Spec.WorkmachineName == "" {
		return nil // Skip RBAC if no WorkMachine is specified
	}

	workmachine, err := r.getWorkMachine(ctx, workspace.Spec.WorkmachineName)
	if err != nil {
		return fmt.Errorf("failed to get WorkMachine: %w", err)
	}

	namespace := workmachine.Spec.TargetNamespace
	workspaceName := workspace.Name

	// Create ServiceAccount for the workspace (no prefix needed since namespaced)
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workspaceName,
			Namespace: namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, sa, func() error {
		// Set Workspace as owner for cascade deletion
		if err := controllerutil.SetControllerReference(workspace, sa, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on ServiceAccount: %w", err)
		}

		if sa.Labels == nil {
			sa.Labels = make(map[string]string)
		}
		sa.Labels["kloudlite.io/workspace-rbac"] = "true"
		sa.Labels["kloudlite.io/workspace-name"] = workspaceName

		return nil
	}); err != nil {
		return fmt.Errorf("failed to create/update ServiceAccount: %w", err)
	}

	// Create workspace-specific Role with ResourceNames restrictions
	role := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      workspaceName,
			Namespace: namespace,
			Labels: map[string]string{
				"kloudlite.io/workspace-rbac": "true",
				"kloudlite.io/workspace-name": workspaceName,
			},
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, role, func() error {
		// Set Workspace as owner for cascade deletion
		if err := controllerutil.SetControllerReference(workspace, role, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on Role: %w", err)
		}

		role.Rules = []rbacv1.PolicyRule{
			{
				// Allow access only to this specific workspace
				APIGroups:     []string{"workspaces.kloudlite.io"},
				Resources:     []string{"workspaces"},
				ResourceNames: []string{workspaceName},
				Verbs:         []string{"get", "list", "watch", "update", "patch"},
			},
			{
				// Allow access to workspace status
				APIGroups:     []string{"workspaces.kloudlite.io"},
				Resources:     []string{"workspaces/status"},
				ResourceNames: []string{workspaceName},
				Verbs:         []string{"get"},
			},
			{
				// Allow reading environments (needed for env connect command)
				APIGroups: []string{"environments.kloudlite.io"},
				Resources: []string{"environments"},
				Verbs:     []string{"get", "list"},
			},
			{
				// Allow reading services in any namespace (needed for intercept command)
				// The kl CLI will only access services in connected environment namespaces
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"get", "list"},
			},
			{
				// Allow updating Composition intercepts (workspace controller manages composition intercepts)
				APIGroups: []string{"environments.kloudlite.io"},
				Resources: []string{"compositions"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				// Allow reading Composition status for intercept status
				APIGroups: []string{"environments.kloudlite.io"},
				Resources: []string{"compositions/status"},
				Verbs:     []string{"get"},
			},
			{
				// Allow managing PackageRequests (cluster-scoped, cannot use ResourceNames)
				// Will be filtered by workspace ownership in application logic
				APIGroups: []string{"workspaces.kloudlite.io"},
				Resources: []string{"packagerequests"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				// Allow reading PackageRequest status
				APIGroups: []string{"workspaces.kloudlite.io"},
				Resources: []string{"packagerequests/status"},
				Verbs:     []string{"get"},
			},
			{
				// Allow reading pod logs (for streaming nix installation output from host-manager)
				// Note: resourceNames doesn't work with subresources like pods/log
				APIGroups: []string{""},
				Resources: []string{"pods/log"},
				Verbs:     []string{"get"},
			},
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to create/update Role: %w", err)
	}

	// Create RoleBinding for the workspace service account
	roleBinding := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-binding", workspaceName),
			Namespace: namespace,
			Labels: map[string]string{
				"kloudlite.io/workspace-rbac": "true",
				"kloudlite.io/workspace-name": workspaceName,
			},
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, roleBinding, func() error {
		// Set Workspace as owner for cascade deletion
		if err := controllerutil.SetControllerReference(workspace, roleBinding, r.Scheme); err != nil {
			return fmt.Errorf("failed to set owner reference on RoleBinding: %w", err)
		}

		roleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      workspaceName,
				Namespace: namespace,
			},
		}

		roleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     workspaceName,
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to create/update RoleBinding: %w", err)
	}

	logger.Info("Successfully created workspace-specific RBAC",
		zap.String("role", role.Name),
		zap.String("roleBinding", roleBinding.Name))

	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workspacev1.Workspace{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&corev1.Pod{}).
		Owns(&networkingv1.Ingress{}). // Watch Ingress resources owned by Workspaces
		// Watch PackageRequest status changes and trigger Workspace reconciliation
		// This is needed because host-manager updates PackageRequest status when packages are installed
		Watches(
			&packagesv1.PackageRequest{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				pkgReq, ok := obj.(*packagesv1.PackageRequest)
				if !ok {
					return nil
				}
				// Find the owning Workspace from the owner references
				for _, ownerRef := range pkgReq.OwnerReferences {
					if ownerRef.Kind == "Workspace" {
						// PackageRequest is cluster-scoped but owned by a namespaced Workspace
						// The namespace should be the same as the WorkMachine's target namespace
						// We need to find the Workspace by name across all namespaces
						var workspaces workspacev1.WorkspaceList
						if err := mgr.GetClient().List(ctx, &workspaces); err != nil {
							return nil
						}
						for _, ws := range workspaces.Items {
							if ws.Name == ownerRef.Name && ws.UID == ownerRef.UID {
								return []reconcile.Request{
									{NamespacedName: types.NamespacedName{Name: ws.Name, Namespace: ws.Namespace}},
								}
							}
						}
					}
				}
				return nil
			}),
		).
		Complete(r)
}
