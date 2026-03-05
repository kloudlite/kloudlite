package workspace

import (
	"context"
	"fmt"
	"time"

	"github.com/kloudlite/kloudlite/api/internal/controllers/shared"
	environmentv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/pagination"
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
)

// ControllerConfig holds workspace controller configuration values
// This is a local copy to avoid import cycles with the controllers package
type ControllerConfig struct {
	Workspace struct {
		DefaultIdleTimeoutMinutes    int
		RequeueIntervalMinutes      int
		RBACCleanupIntervalMinutes  int
		KubectlImage               string
		GitImage                   string
		AlpineImage                string
		CleanupPodTTLSeconds       int64
		VSCodeVersion              string
	}
	Environment struct {
		LifecycleRetryInterval time.Duration
	}
}

// Controller configuration - initialized during controller setup
var cfg *ControllerConfig

// Global pod deletion tracker to prevent race conditions between workspace and workmachine controllers
var podDeletionTracker *shared.PodDeletionTracker

// shouldRunRBACCleanup determines if RBAC garbage collection should run
// Uses a simple time-based check to run cleanup periodically
func shouldRunRBACCleanup() bool {
	// Use current minute to determine if we should run cleanup
	// Run every rbacCleanupIntervalMinutes minutes
	currentMinute := time.Now().Minute()
	return currentMinute%cfg.Workspace.RBACCleanupIntervalMinutes == 0
}

// WorkspaceReconciler reconciles Workspace objects and manages VS Code server pods
type WorkspaceReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Logger    *zap.Logger
	Config    *rest.Config
	Clientset *kubernetes.Clientset
	JWTSecret string // JWT secret (kept for compatibility, no longer used for registry)
	Cfg       *ControllerConfig // Controller configuration
}

// Reconcile handles Workspace events and ensures the workspace pod exists
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("workspace", req.Name),
	)

	logger.Info("Reconciling Workspace")

	// Periodically run orphaned RBAC cleanup (every rbacCleanupIntervalMinutes)
	// We use a simple counter approach by checking if this is the first workspace in the list
	// This avoids running cleanup on every reconciliation
	if shouldRunRBACCleanup() {
		go func() {
			r.Logger.Info("Running periodic orphaned RBAC cleanup")
			deletedCount, errors := r.cleanupOrphanedRBACResources(ctx, r.Logger)
			if len(errors) > 0 {
				r.Logger.Warn("Periodic RBAC cleanup encountered errors",
					zap.Int("deletedCount", deletedCount),
					zap.Int("errorCount", len(errors)),
					zap.String("errors", fmt.Sprintf("%v", errors)))
			}
		}()
	}

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

	// Add finalizers and hash label if not present
	finalizersAdded := false
	labelsUpdated := false

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

	// Ensure hash label is set on the workspace resource for efficient lookups
	wsHash := generateHash(fmt.Sprintf("%s-%s", workspace.Spec.OwnedBy, workspace.Name))
	labels := workspace.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	if labels["kloudlite.io/hash"] != wsHash {
		labels["kloudlite.io/hash"] = wsHash
		workspace.Labels = labels
		labelsUpdated = true
		logger.Info("Adding hash label to workspace", zap.String("hash", wsHash))
	}

	if finalizersAdded || labelsUpdated {
		if err := r.Update(ctx, workspace); err != nil {
			logger.Error("Failed to update workspace", zap.Error(err))
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
		workspace.Spec.VSCodeVersion = cfg.Workspace.VSCodeVersion
	}

	// Handle workspace creation from snapshot if fromSnapshot is set
	// Snapshot restore takes precedence over normal workspace reconciliation
	if workspace.Spec.FromSnapshot != nil {
		logger.Info("Workspace has fromSnapshot set, handling snapshot restore",
			zap.String("snapshotName", workspace.Spec.FromSnapshot.SnapshotName))
		return r.handleSnapshotRestore(ctx, workspace, logger)
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

	// Requeue after configured interval to check idle status periodically for all active workspaces
	// This is needed for idle state tracking (displayed in UI) even when auto-stop is not enabled
	if workspace.Spec.Status == "active" {
		if result.RequeueAfter == 0 && !result.Requeue {
			result.RequeueAfter = time.Duration(cfg.Workspace.RequeueIntervalMinutes) * time.Minute
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
				// Allow reading services in any namespace (needed for intercept command)
				// The kl CLI will only access services in connected environment namespaces
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"get", "list"},
			},
			{
				// Allow updating Environment intercepts (workspace controller manages intercepts in Environment.Spec.Compose)
				APIGroups: []string{"environments.kloudlite.io"},
				Resources: []string{"environments"},
				Verbs:     []string{"get", "list", "watch", "update", "patch"},
			},
			{
				// Allow reading Environment status for intercept status
				APIGroups: []string{"environments.kloudlite.io"},
				Resources: []string{"environments/status"},
				Verbs:     []string{"get"},
			},
			// Note: PackageRequests are cluster-scoped, so they are granted in the ClusterRole below
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

	// Create ClusterRole for cluster-scoped resources (environments)
	// Note: ClusterRoles cannot have owner references to namespaced resources
	clusterRoleName := fmt.Sprintf("workspace-%s-%s", namespace, workspaceName)
	clusterRole := &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
			Labels: map[string]string{
				"kloudlite.io/workspace-rbac":      "true",
				"kloudlite.io/workspace-name":      workspaceName,
				"kloudlite.io/workspace-namespace": namespace,
			},
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, clusterRole, func() error {
		clusterRole.Rules = []rbacv1.PolicyRule{
			{
				// Allow reading environments (cluster-scoped resource)
				// Needed for env connect command and VPN hosts listing
				APIGroups: []string{"environments.kloudlite.io"},
				Resources: []string{"environments"},
				Verbs:     []string{"get", "list"},
			},
			{
				// Allow reading and updating environments
				// Needed for kl intercept commands to manage service intercepts in Environment.Spec.Compose
				APIGroups: []string{"environments.kloudlite.io"},
				Resources: []string{"environments"},
				Verbs:     []string{"get", "list", "update", "patch"},
			},
			{
				// Allow reading services in any namespace
				// Needed for kl svc list command to show services in connected environment
				APIGroups: []string{""},
				Resources: []string{"services"},
				Verbs:     []string{"get", "list"},
			},
			{
				// Allow managing PackageRequests (cluster-scoped resource)
				// Will be filtered by workspace ownership in application logic
				APIGroups: []string{"packages.kloudlite.io"},
				Resources: []string{"packagerequests"},
				Verbs:     []string{"get", "list", "watch", "create", "update", "patch", "delete"},
			},
			{
				// Allow reading PackageRequest status
				APIGroups: []string{"packages.kloudlite.io"},
				Resources: []string{"packagerequests/status"},
				Verbs:     []string{"get"},
			},
		}
		return nil
	}); err != nil {
		return fmt.Errorf("failed to create/update ClusterRole: %w", err)
	}

	// Create ClusterRoleBinding to bind the ClusterRole to the workspace ServiceAccount
	clusterRoleBinding := &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: clusterRoleName,
			Labels: map[string]string{
				"kloudlite.io/workspace-rbac":      "true",
				"kloudlite.io/workspace-name":      workspaceName,
				"kloudlite.io/workspace-namespace": namespace,
			},
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, clusterRoleBinding, func() error {
		clusterRoleBinding.Subjects = []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      workspaceName,
				Namespace: namespace,
			},
		}

		clusterRoleBinding.RoleRef = rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     clusterRoleName,
		}

		return nil
	}); err != nil {
		return fmt.Errorf("failed to create/update ClusterRoleBinding: %w", err)
	}

	// Note: CA certificate is now mounted from local namespace secret (kloudlite-wildcard-cert-tls)
	// No cross-namespace RBAC needed for kloudlite namespace

	logger.Info("Successfully created workspace-specific RBAC",
		zap.String("role", role.Name),
		zap.String("roleBinding", roleBinding.Name),
		zap.String("clusterRole", clusterRoleName),
		zap.String("clusterRoleBinding", clusterRoleName))

	return nil
}

// SetupWithManager sets up the controller with the Manager
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize the global pod deletion tracker if not already initialized
	if podDeletionTracker == nil {
		podDeletionTracker = shared.NewPodDeletionTracker(r.Logger)
		r.Logger.Info("Initialized pod deletion tracker for race condition prevention")
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&workspacev1.Workspace{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&corev1.Pod{}).
		Owns(&networkingv1.Ingress{}). // Watch Ingress resources owned by Workspaces
		Watches(
			&environmentv1.Environment{},
			handler.EnqueueRequestsFromMapFunc(r.findWorkspacesForEnvironment),
		).
		Watches(
			&rbacv1.ClusterRole{},
			handler.EnqueueRequestsFromMapFunc(r.enqueueForRBACCleanup),
		).
		Watches(
			&rbacv1.ClusterRoleBinding{},
			handler.EnqueueRequestsFromMapFunc(r.enqueueForRBACCleanup),
		).
		Complete(r)
}

// enqueueForRBACCleanup triggers orphaned RBAC cleanup when ClusterRole or ClusterRoleBinding changes
// This ensures leaked resources are cleaned up promptly
func (r *WorkspaceReconciler) enqueueForRBACCleanup(ctx context.Context, obj client.Object) []reconcile.Request {
	// Run orphaned RBAC cleanup in the background
	go func() {
		r.Logger.Info("Triggering orphaned RBAC cleanup")
		deletedCount, errors := r.cleanupOrphanedRBACResources(ctx, r.Logger)
		if len(errors) > 0 {
			r.Logger.Warn("Triggered RBAC cleanup encountered errors",
				zap.Int("deletedCount", deletedCount),
				zap.Int("errorCount", len(errors)),
				zap.String("errors", fmt.Sprintf("%v", errors)))
		}
	}()

	// Return empty request list since we're just triggering cleanup
	return nil
}

// findWorkspacesForEnvironment finds all workspaces connected to a deactivated environment.
//
// Uses pagination with a default page size of 100 items to efficiently find workspaces
// in clusters with many resources. This prevents API server overload when environments
// are deactivated and all connected workspaces need to be reconciled.
func (r *WorkspaceReconciler) findWorkspacesForEnvironment(ctx context.Context, obj client.Object) []reconcile.Request {
	env, ok := obj.(*environmentv1.Environment)
	if !ok {
		r.Logger.Info("findWorkspacesForEnvironment: object is not an Environment")
		return nil
	}

	r.Logger.Info("findWorkspacesForEnvironment called",
		zap.String("environment", env.Name),
		zap.Bool("activated", env.Spec.Activated))

	// Only trigger when environment is deactivated
	if env.Spec.Activated {
		return nil
	}

	// Find all workspaces connected to this environment using pagination
	// Default page size of 100 prevents API overload in large clusters
	var workspaces workspacev1.WorkspaceList
	if err := pagination.ListAll(ctx, r.Client, &workspaces); err != nil {
		r.Logger.Error("findWorkspacesForEnvironment: failed to list workspaces", zap.Error(err))
		return nil
	}

	r.Logger.Info("findWorkspacesForEnvironment: listed all workspaces using pagination", zap.Int("count", len(workspaces.Items)))

	var requests []reconcile.Request
	for _, ws := range workspaces.Items {
		if ws.Spec.EnvironmentConnection != nil &&
			ws.Spec.EnvironmentConnection.EnvironmentRef.Name == env.Name {
			r.Logger.Info("findWorkspacesForEnvironment: workspace connected to deactivated environment",
				zap.String("workspace", ws.Name),
				zap.String("namespace", ws.Namespace))
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: ws.Name, Namespace: ws.Namespace},
			})
		}
	}

	r.Logger.Info("findWorkspacesForEnvironment: returning requests", zap.Int("count", len(requests)))
	return requests
}
