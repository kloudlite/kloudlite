package user

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	userv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
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

// updateUserStatus updates the User status with retry logic
func (r *UserReconciler) updateUserStatus(ctx context.Context, user *userv1alpha1.User, updateFunc func() error, logger *zap.Logger) error {
	return statusutil.UpdateStatusWithRetry(ctx, r.Client, user, updateFunc, logger)
}

// Reconcile handles User events and ensures each user has a WorkMachine
func (r *UserReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := r.Logger.With(
		zap.String("user", req.Name),
		zap.String("namespace", req.Namespace),
	)

	logger.Info("Reconciling User")

	// Fetch the User instance
	user := &userv1alpha1.User{}
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

	// Handle password updates with change detection
	if user.Spec.PasswordString != "" {
		// Calculate hash of the new password string for comparison
		newPasswordHash := sha256.Sum256([]byte(user.Spec.PasswordString))
		newPasswordHashStr := base64.StdEncoding.EncodeToString(newPasswordHash[:])

		// Only update if password has actually changed
		if user.Status.PasswordHash != newPasswordHashStr {
			logger.Info("Updating user password")

			b, err := bcrypt.GenerateFromPassword([]byte(user.Spec.PasswordString), bcrypt.DefaultCost)
			if err != nil {
				logger.Error("Failed to hash password", zap.Error(err))
				return ctrl.Result{}, err
			}

			user.Spec.Password = base64.StdEncoding.EncodeToString(b)
			user.Spec.PasswordString = ""

			// Update status to track the password hash
			user.Status.PasswordHash = newPasswordHashStr

			// Initialize status if needed
			if user.Status.Conditions == nil {
				user.Status.Conditions = []metav1.Condition{}
			}

			// Add or update password condition
			now := metav1.Now()
			passwordCondition := metav1.Condition{
				Type:               "PasswordSet",
				Status:             metav1.ConditionTrue,
				LastTransitionTime: now,
				Reason:             "PasswordUpdated",
				Message:            "User password has been successfully updated",
			}

			// Update existing condition or add new one
			conditionUpdated := false
			for i, condition := range user.Status.Conditions {
				if condition.Type == "PasswordSet" {
					user.Status.Conditions[i] = passwordCondition
					conditionUpdated = true
					break
				}
			}
			if !conditionUpdated {
				user.Status.Conditions = append(user.Status.Conditions, passwordCondition)
			}

			if err := r.Update(ctx, user); err != nil {
				logger.Error("Failed to update user with new password", zap.Error(err))
				return reconcile.Result{}, err
			}

			logger.Info("Successfully updated user password")
			return reconcile.Result{Requeue: true}, nil
		} else {
			// Password hasn't changed, just clear the PasswordString field
			user.Spec.PasswordString = ""
			if err := r.Update(ctx, user); err != nil {
				logger.Error("Failed to clear PasswordString field", zap.Error(err))
				return reconcile.Result{}, err
			}
			return reconcile.Result{Requeue: true}, nil
		}
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

	// Note: WorkMachine auto-creation is now handled by handlers, not the controller.
	// The controller only handles reconciliation of existing resources and status updates.
	// This follows the architectural pattern:
	// - Handlers: HTTP request handling and business logic (auto-creation based on requests)
	// - Webhooks: Resource-level validation
	// - Controllers: Reconciliation of existing resources

	// Update user status based on activation
	isUserActive := user.Spec.Active != nil && *user.Spec.Active

	needsStatusUpdate := false
	if user.Status.Phase == "" ||
		(isUserActive && user.Status.Phase != "active") ||
		(!isUserActive && user.Status.Phase != "inactive") {
		needsStatusUpdate = true
	}

	if needsStatusUpdate {
		// Capture isUserActive for use in updateFunc
		active := isUserActive

		if err := r.updateUserStatus(ctx, user, func() error {
			if user.Status.Conditions == nil {
				user.Status.Conditions = []metav1.Condition{}
			}

			// Update phase
			if active {
				user.Status.Phase = "active"
			} else {
				user.Status.Phase = "inactive"
			}

			// Add/update activation condition
			now := metav1.Now()
			activationCondition := metav1.Condition{
				Type:               "Active",
				Status:             metav1.ConditionTrue,
				LastTransitionTime: now,
				Reason:             "UserStatusUpdated",
				Message:            fmt.Sprintf("User is %s", user.Status.Phase),
			}
			if !active {
				activationCondition.Status = metav1.ConditionFalse
				activationCondition.Reason = "UserDeactivated"
				activationCondition.Message = "User has been deactivated"
			}

			// Update existing condition or add new one
			conditionUpdated := false
			for i, condition := range user.Status.Conditions {
				if condition.Type == "Active" {
					user.Status.Conditions[i] = activationCondition
					conditionUpdated = true
					break
				}
			}
			if !conditionUpdated {
				user.Status.Conditions = append(user.Status.Conditions, activationCondition)
			}

			if user.Status.CreatedAt == nil {
				user.Status.CreatedAt = &now
			}

			return nil
		}, logger); err != nil {
			logger.Warn("Failed to update user status", zap.Error(err))
		} else {
			logger.Info("Updated user status based on activation",
				zap.String("phase", user.Status.Phase),
				zap.Bool("userActive", active))
		}
	}

	return reconcile.Result{}, nil
}

// handleUserDeletion handles cleanup when a user is being deleted
func (r *UserReconciler) handleUserDeletion(ctx context.Context, user *userv1alpha1.User, logger *zap.Logger) (reconcile.Result, error) {
	// Note: WorkMachine deletion is handled by Kubernetes garbage collection via OwnerReferences.
	// The WorkMachine controller set up owner references when creating WorkMachines,
	// so they will be automatically deleted when the user is deleted.
	// We just need to remove the finalizer to allow user deletion.

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

// Note: WorkMachine creation helper functions removed.
// WorkMachine auto-creation is now handled by the UserService in handlers.

// SetupWithManager sets up the controller with the Manager
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&userv1alpha1.User{}).
		Owns(&machinesv1.WorkMachine{}). // Watch WorkMachines owned by Users for garbage collection
		Complete(r)
}
