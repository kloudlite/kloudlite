package workspace

import (
	"context"
	"time"

	interceptsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/serviceintercept/v1"
	workspacev1 "github.com/kloudlite/kloudlite/api/internal/controllers/workspace/v1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	workspaceFinalizer = "workspaces.kloudlite.io/finalizer"

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

	// Fetch the Workspace instance (cluster-scoped, no namespace)
	workspace := &workspacev1.Workspace{}
	err := r.Get(ctx, client.ObjectKey{Name: req.Name}, workspace)
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

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(workspace, workspaceFinalizer) {
		logger.Info("Adding finalizer to workspace")
		controllerutil.AddFinalizer(workspace, workspaceFinalizer)
		if err := r.Update(ctx, workspace); err != nil {
			logger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
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
				// Set WorkMachine as controller owner for cascading deletion
				if err := controllerutil.SetControllerReference(workmachine, workspace, r.Scheme); err != nil {
					logger.Error("Failed to set WorkMachine as owner",
						zap.String("workmachine", workspace.Spec.WorkmachineName),
						zap.Error(err))
					// Don't fail reconciliation
				} else {
					if err := r.Update(ctx, workspace); err != nil {
						logger.Error("Failed to update Workspace with owner reference", zap.Error(err))
						return reconcile.Result{}, err
					}
					logger.Info("Successfully set WorkMachine as owner of Workspace")
					return reconcile.Result{Requeue: true}, nil
				}
			}
		}
	}

	// Set default workspace path if not provided
	if workspace.Spec.WorkspacePath == "" {
		workspace.Spec.WorkspacePath = "/workspace"
	}

	// Set default VS Code version if not provided
	if workspace.Spec.VSCodeVersion == "" {
		workspace.Spec.VSCodeVersion = "latest"
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

// SetupWithManager sets up the controller with the Manager
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workspacev1.Workspace{}).
		Owns(&corev1.Pod{}).
		Owns(&workspacev1.PackageRequest{}).
		Watches(
			&interceptsv1.ServiceIntercept{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				// Extract workspace name from labels (cluster-scoped, no namespace needed)
				labels := obj.GetLabels()
				workspaceName := labels["workspaces.kloudlite.io/workspace-name"]

				if workspaceName == "" {
					return nil
				}

				// Trigger reconciliation for the workspace (cluster-scoped)
				return []reconcile.Request{
					{
						NamespacedName: client.ObjectKey{
							Name: workspaceName,
						},
					},
				}
			}),
		).
		Complete(r)
}
