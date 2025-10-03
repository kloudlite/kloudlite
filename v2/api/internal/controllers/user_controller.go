package controllers

import (
	"context"
	"fmt"
	"strings"
	"time"

	machinesv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/machines/v1"
	platformv1alpha1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/platform/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// UserReconciler reconciles User objects and creates WorkMachines
type UserReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Logger *zap.Logger
}

const UserFinalizerName = "user.platform.kloudlite.io/cleanup"

// Reconcile handles User events and ensures each user has a WorkMachine
func (r *UserReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("user", req.Name),
		zap.String("namespace", req.Namespace),
	)

	logger.Info("Reconciling User")

	// Fetch the User instance
	user := &platformv1alpha1.User{}
	err := r.Get(ctx, req.NamespacedName, user)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// User has been deleted, nothing to do
			logger.Info("User not found, likely deleted")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to get User", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Check if user is being deleted
	if user.DeletionTimestamp != nil {
		logger.Info("User is being deleted, starting cleanup")
		return r.handleUserDeletion(ctx, user, logger)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(user, UserFinalizerName) {
		logger.Info("Adding finalizer to user")
		controllerutil.AddFinalizer(user, UserFinalizerName)
		if err := r.Update(ctx, user); err != nil {
			logger.Error("Failed to add finalizer", zap.Error(err))
			return reconcile.Result{}, err
		}
		return reconcile.Result{Requeue: true}, nil
	}

	// Check if WorkMachine already exists for this user
	workMachineName := r.generateWorkMachineName(user)
	existingWorkMachine := &machinesv1.WorkMachine{}
	err = r.Get(ctx, client.ObjectKey{Name: workMachineName}, existingWorkMachine)

	if err == nil {
		// WorkMachine already exists
		logger.Info("WorkMachine already exists for user",
			zap.String("workMachine", workMachineName))

		// Check if the WorkMachine is owned by this user
		if existingWorkMachine.Spec.OwnedBy != user.Spec.Email {
			logger.Warn("WorkMachine exists but owned by different user",
				zap.String("workMachine", workMachineName),
				zap.String("currentOwner", existingWorkMachine.Spec.OwnedBy),
				zap.String("expectedOwner", user.Spec.Email))
		}

		// Handle user activation/deactivation - update WorkMachine state
		updated := false
		isUserActive := user.Spec.Active != nil && *user.Spec.Active

		if isUserActive {
			// User is active - ensure WorkMachine is not disabled
			if existingWorkMachine.Spec.DesiredState == machinesv1.MachineStateDisabled {
				logger.Info("User activated - enabling WorkMachine", zap.String("workMachine", workMachineName))
				existingWorkMachine.Spec.DesiredState = machinesv1.MachineStateStopped
				updated = true
			}
		} else {
			// User is inactive - disable WorkMachine
			if existingWorkMachine.Spec.DesiredState != machinesv1.MachineStateDisabled {
				logger.Info("User deactivated - disabling WorkMachine", zap.String("workMachine", workMachineName))
				existingWorkMachine.Spec.DesiredState = machinesv1.MachineStateDisabled
				updated = true
			}
		}

		// Update WorkMachine if needed
		if updated {
			if err := r.Update(ctx, existingWorkMachine); err != nil {
				logger.Error("Failed to update WorkMachine state", zap.Error(err))
				return reconcile.Result{}, err
			}
			logger.Info("Updated WorkMachine state based on user activation status",
				zap.String("workMachine", workMachineName),
				zap.Bool("userActive", isUserActive),
				zap.String("desiredState", string(existingWorkMachine.Spec.DesiredState)))
		}

		return reconcile.Result{}, nil
	}

	if !apierrors.IsNotFound(err) {
		logger.Error("Failed to check existing WorkMachine", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Create WorkMachine for the user
	logger.Info("Creating WorkMachine for user", zap.String("workMachine", workMachineName))

	// Create namespace for the WorkMachine first
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: workMachineName,
			Labels: map[string]string{
				"kloudlite.io/workmachine": workMachineName,
				"kloudlite.io/user":        user.Name,
			},
			Annotations: map[string]string{
				"kloudlite.io/owner-email": user.Spec.Email,
			},
		},
	}

	// Create the namespace if it doesn't exist
	if err := r.Create(ctx, namespace); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			logger.Error("Failed to create namespace for WorkMachine", zap.Error(err))
			return reconcile.Result{}, err
		}
		logger.Info("Namespace already exists", zap.String("namespace", workMachineName))
	} else {
		logger.Info("Created namespace for WorkMachine", zap.String("namespace", workMachineName))
	}

	// Note: RBAC for workspace management is handled by ClusterRoleBinding
	// No need to create namespace-specific RBAC resources

	workMachine, err := r.buildWorkMachineForUser(ctx, user, workMachineName)
	if err != nil {
		logger.Error("Failed to build WorkMachine", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Set User as the owner of the WorkMachine using OwnerReferences
	// This ensures the WorkMachine is garbage collected if the User is deleted
	if err := controllerutil.SetControllerReference(user, workMachine, r.Scheme); err != nil {
		logger.Error("Failed to set owner reference", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Create the WorkMachine
	if err := r.Create(ctx, workMachine); err != nil {
		if apierrors.IsAlreadyExists(err) {
			// Another reconciliation might have created it
			logger.Info("WorkMachine already exists (race condition)")
			return reconcile.Result{}, nil
		}
		logger.Error("Failed to create WorkMachine", zap.Error(err))
		// Retry after a delay
		return reconcile.Result{RequeueAfter: 30 * time.Second}, err
	}

	logger.Info("Successfully created WorkMachine for user",
		zap.String("workMachine", workMachineName),
		zap.String("targetNamespace", workMachine.Spec.TargetNamespace))

	// Update user metadata to indicate WorkMachine has been created
	if user.Spec.Metadata == nil {
		user.Spec.Metadata = make(map[string]string)
	}
	user.Spec.Metadata["workmachine-name"] = workMachineName
	user.Spec.Metadata["workmachine-created"] = "true"

	if err := r.Update(ctx, user); err != nil {
		logger.Warn("Failed to update user metadata", zap.Error(err))
		// Don't fail the reconciliation for metadata update failures
	}

	return reconcile.Result{}, nil
}

// handleUserDeletion handles cleanup when a user is being deleted
func (r *UserReconciler) handleUserDeletion(ctx context.Context, user *platformv1alpha1.User, logger *zap.Logger) (reconcile.Result, error) {
	workMachineName := r.generateWorkMachineName(user)

	// Check if WorkMachine still exists
	workMachine := &machinesv1.WorkMachine{}
	err := r.Get(ctx, client.ObjectKey{Name: workMachineName}, workMachine)

	if err == nil {
		// WorkMachine still exists
		if workMachine.DeletionTimestamp == nil {
			// WorkMachine is not being deleted yet, delete it
			logger.Info("Deleting WorkMachine", zap.String("workMachine", workMachineName))
			if err := r.Delete(ctx, workMachine); err != nil && !apierrors.IsNotFound(err) {
				logger.Error("Failed to delete WorkMachine", zap.String("workMachine", workMachineName), zap.Error(err))
				return reconcile.Result{}, err
			}
			logger.Info("WorkMachine deletion initiated", zap.String("workMachine", workMachineName))
		} else {
			// WorkMachine is being deleted, wait for it to complete
			logger.Info("Waiting for WorkMachine deletion to complete",
				zap.String("workMachine", workMachineName),
				zap.Time("deletionTimestamp", workMachine.DeletionTimestamp.Time))
		}

		// Requeue to wait for WorkMachine deletion (WorkMachine finalizer will handle namespace cleanup)
		return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
	}

	if !apierrors.IsNotFound(err) {
		logger.Error("Failed to get WorkMachine", zap.String("workMachine", workMachineName), zap.Error(err))
		return reconcile.Result{}, err
	}

	// WorkMachine is fully deleted (which means namespace is also deleted by WorkMachine controller)
	logger.Info("WorkMachine is fully deleted, proceeding with user cleanup", zap.String("workMachine", workMachineName))

	// Remove finalizer to allow user deletion
	logger.Info("Removing finalizer from user")
	controllerutil.RemoveFinalizer(user, UserFinalizerName)
	if err := r.Update(ctx, user); err != nil {
		logger.Error("Failed to remove finalizer", zap.Error(err))
		return reconcile.Result{}, err
	}

	logger.Info("User cleanup completed successfully")
	return reconcile.Result{}, nil
}

// generateWorkMachineName generates a WorkMachine name from user resource name
func (r *UserReconciler) generateWorkMachineName(user *platformv1alpha1.User) string {
	// Use the User resource name directly
	return fmt.Sprintf("wm-%s", user.Name)
}

// buildWorkMachineForUser creates a WorkMachine resource for the user
func (r *UserReconciler) buildWorkMachineForUser(ctx context.Context, user *platformv1alpha1.User, workMachineName string) (*machinesv1.WorkMachine, error) {
	// Use the User resource name for the target namespace
	targetNamespace := fmt.Sprintf("wm-%s", user.Name)

	// Determine initial state based on user activation status
	initialState := machinesv1.MachineStateStopped
	isUserActive := user.Spec.Active != nil && *user.Spec.Active
	if !isUserActive {
		initialState = machinesv1.MachineStateDisabled
	}

	// Query for the default machine type
	machineTypeList := &machinesv1.MachineTypeList{}
	if err := r.List(ctx, machineTypeList); err != nil {
		return nil, fmt.Errorf("failed to list machine types: %v", err)
	}

	// Find the default machine type
	machineType := "standard-2vcpu-4gb" // Fallback default
	for _, mt := range machineTypeList.Items {
		if mt.Spec.IsDefault && mt.Spec.Active {
			machineType = mt.Name
			break
		}
	}

	return &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: workMachineName,
			Labels: map[string]string{
				"kloudlite.io/user-email": sanitizeForLabel(user.Spec.Email),
				"kloudlite.io/user-name":  user.Name,
				"kloudlite.io/managed":    "true",
				"kloudlite.io/created-by": "user-controller",
			},
			Annotations: map[string]string{
				"kloudlite.io/user-uid":       string(user.UID),
				"kloudlite.io/creation-reason": "auto-created-for-user",
			},
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         user.Spec.Email,
			MachineType:     machineType,
			TargetNamespace: targetNamespace,
			DesiredState:    initialState,
		},
	}, nil
}

// extractUsernameFromEmail extracts the username part from an email
func extractUsernameFromEmail(email string) string {
	username := email
	if idx := strings.Index(username, "@"); idx > 0 {
		username = username[:idx]
	}
	// Replace dots and special characters with hyphens for valid k8s names
	username = strings.ReplaceAll(username, ".", "-")
	username = strings.ReplaceAll(username, "_", "-")
	username = strings.ReplaceAll(username, "+", "-")
	username = strings.ToLower(username)

	// Ensure the username starts with a letter or number
	if len(username) > 0 && !isAlphanumeric(username[0]) {
		username = "u-" + username
	}

	// Limit length to ensure the full name stays within k8s limits
	if len(username) > 50 {
		username = username[:50]
	}

	// Trim trailing hyphens
	username = strings.TrimRight(username, "-")

	return username
}

// sanitizeForLabel sanitizes a string to be used as a label value
func sanitizeForLabel(value string) string {
	// Replace special characters with hyphens for label value
	sanitized := strings.ReplaceAll(value, "@", "-at-")
	sanitized = strings.ReplaceAll(sanitized, ".", "-dot-")
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	sanitized = strings.ReplaceAll(sanitized, "+", "-plus-")
	sanitized = strings.ToLower(sanitized)

	// Ensure it starts and ends with alphanumeric
	sanitized = strings.Trim(sanitized, "-")

	// Limit length to 63 characters (Kubernetes label value limit)
	if len(sanitized) > 63 {
		sanitized = sanitized[:63]
	}

	// Ensure it ends with alphanumeric
	sanitized = strings.TrimRight(sanitized, "-")

	return sanitized
}

// isAlphanumeric checks if a byte is alphanumeric
func isAlphanumeric(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9')
}


// SetupWithManager sets up the controller with the Manager
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&platformv1alpha1.User{}).
		Owns(&machinesv1.WorkMachine{}). // Watch WorkMachines owned by Users
		Complete(r)
}