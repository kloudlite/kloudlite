package user

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	userv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/internal/pkg/statusutil"
	"github.com/kloudlite/kloudlite/api/pkg/utils"
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

		// Check user activation status and update user status accordingly
		isUserActive := user.Spec.Active != nil && *user.Spec.Active

		// Update user status based on activation
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

				return nil
			}, logger); err != nil {
				logger.Warn("Failed to update user status", zap.Error(err))
			} else {
				logger.Info("Updated user status based on activation",
					zap.String("phase", user.Status.Phase),
					zap.Bool("userActive", active))
			}
		}

		// Note: WorkMachine state management is now handled by the WorkMachine controller
		// based on the user's activation status and other conditions

		return reconcile.Result{}, nil
	}

	if !apierrors.IsNotFound(err) {
		logger.Error("Failed to check existing WorkMachine", zap.Error(err))
		return reconcile.Result{}, err
	}

	// Check if user has 'user' role - only users with 'user' role get WorkMachines
	hasUserRole := false
	for _, role := range user.Spec.Roles {
		if role == userv1alpha1.RoleUser {
			hasUserRole = true
			break
		}
	}

	if !hasUserRole {
		logger.Info("User does not have 'user' role - skipping WorkMachine creation",
			zap.String("user", user.Name),
			zap.Strings("roles", rolesToStrings(user.Spec.Roles)))
		return reconcile.Result{}, nil
	}

	// Create WorkMachine for the user
	logger.Info("Creating WorkMachine for user", zap.String("workMachine", workMachineName))

	// Note: Namespace creation is now handled by the WorkMachine controller
	// This removes the tight coupling between User and WorkMachine namespace management

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

	// Update user status to track WorkMachine information
	// Capture variables for use in updateFunc
	wmName := workMachineName
	isUserActive := user.Spec.Active != nil && *user.Spec.Active

	if err := r.updateUserStatus(ctx, user, func() error {
		// Initialize status if needed
		if user.Status.Conditions == nil {
			user.Status.Conditions = []metav1.Condition{}
		}

		// Add WorkMachineReady condition
		now := metav1.Now()
		workMachineCondition := metav1.Condition{
			Type:               "WorkMachineReady",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: now,
			Reason:             "WorkMachineCreated",
			Message:            fmt.Sprintf("WorkMachine %s has been successfully created", wmName),
		}

		// Update existing condition or add new one
		conditionUpdated := false
		for i, condition := range user.Status.Conditions {
			if condition.Type == "WorkMachineReady" {
				user.Status.Conditions[i] = workMachineCondition
				conditionUpdated = true
				break
			}
		}
		if !conditionUpdated {
			user.Status.Conditions = append(user.Status.Conditions, workMachineCondition)
		}

		// Set user phase based on activity status
		if isUserActive {
			user.Status.Phase = "active"
		} else {
			user.Status.Phase = "inactive"
		}

		if user.Status.CreatedAt == nil {
			user.Status.CreatedAt = &now
		}

		return nil
	}, logger); err != nil {
		logger.Warn("Failed to update user status", zap.Error(err))
		// Don't fail the reconciliation for status update failures
	}

	return reconcile.Result{}, nil
}

// handleUserDeletion handles cleanup when a user is being deleted
func (r *UserReconciler) handleUserDeletion(ctx context.Context, user *userv1alpha1.User, logger *zap.Logger) (reconcile.Result, error) {
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
func (r *UserReconciler) generateWorkMachineName(user *userv1alpha1.User) string {
	// Use the User resource name directly
	return fmt.Sprintf("wm-%s", user.Name)
}

// buildWorkMachineForUser creates a WorkMachine resource for the user
func (r *UserReconciler) buildWorkMachineForUser(ctx context.Context, user *userv1alpha1.User, workMachineName string) (*machinesv1.WorkMachine, error) {
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
				"kloudlite.io/user-email": utils.SanitizeForLabel(user.Spec.Email),
				"kloudlite.io/user-name":  user.Name,
				"kloudlite.io/managed":    "true",
				"kloudlite.io/created-by": "user-controller",
			},
			Annotations: map[string]string{
				"kloudlite.io/user-uid":        string(user.UID),
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


// rolesToStrings converts RoleType slice to string slice for logging
func rolesToStrings(roles []userv1alpha1.RoleType) []string {
	result := make([]string, len(roles))
	for i, role := range roles {
		result[i] = string(role)
	}
	return result
}

// SetupWithManager sets up the controller with the Manager
func (r *UserReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&userv1alpha1.User{}).
		Owns(&machinesv1.WorkMachine{}). // Watch WorkMachines owned by Users for garbage collection
		Complete(r)
}
