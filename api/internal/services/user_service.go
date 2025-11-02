package services

import (
	"context"
	"fmt"
	"strings"

	platformv1alpha1 "github.com/kloudlite/kloudlite/api/internal/controllers/user/v1alpha1"
	machinesv1 "github.com/kloudlite/kloudlite/api/internal/controllers/workmachine/v1"
	"github.com/kloudlite/kloudlite/api/internal/repository"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UserService provides business logic for User operations
type UserService interface {
	// CRUD operations
	CreateUser(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error)
	GetUser(ctx context.Context, name string) (*platformv1alpha1.User, error)
	GetUserByEmail(ctx context.Context, email string) (*platformv1alpha1.User, error)
	GetUserByUsername(ctx context.Context, username string) (*platformv1alpha1.User, error)
	UpdateUser(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error)
	DeleteUser(ctx context.Context, name string) error
	ListUsers(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error)

	// Domain-specific operations
	ActivateUser(ctx context.Context, name string) (*platformv1alpha1.User, error)
	DeactivateUser(ctx context.Context, name string) (*platformv1alpha1.User, error)
	ResetUserPassword(ctx context.Context, name, newPassword string) error
	UpdateUserLastLogin(ctx context.Context, name string) error
}

// userService implements UserService
type userService struct {
	userRepo        repository.UserRepository
	workMachineRepo repository.WorkMachineRepository
}

// NewUserService creates a new UserService
func NewUserService(userRepo repository.UserRepository, workMachineRepo repository.WorkMachineRepository) UserService {
	return &userService{
		userRepo:        userRepo,
		workMachineRepo: workMachineRepo,
	}
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
	// Users are cluster-scoped resources, no namespace needed

	// All validations and mutations are handled by webhooks
	// Just create the user in repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		if repository.IsAlreadyExists(err) {
			return nil, fmt.Errorf("user already exists: %w", err)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create WorkMachine only if user has 'user' role
	if s.hasUserRole(user) {
		if err := s.createWorkMachineForUser(ctx, user); err != nil {
			// Log error but don't fail user creation
			fmt.Printf("Warning: Failed to create WorkMachine for user %s: %v\n", user.Spec.Email, err)
		}
	}

	return user, nil
}

// GetUser retrieves a user by name (cluster-scoped)
func (s *userService) GetUser(ctx context.Context, name string) (*platformv1alpha1.User, error) {
	user, err := s.userRepo.Get(ctx, name)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, fmt.Errorf("user not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email address
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*platformv1alpha1.User, error) {
	// This would need to iterate through users or use a label selector
	// The webhook adds labels for efficient lookup
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	for i := range users.Items {
		if users.Items[i].Spec.Email == email {
			return &users.Items[i], nil
		}
	}

	return nil, fmt.Errorf("user not found with email: %s", email)
}

// GetUserByUsername retrieves a user by username
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*platformv1alpha1.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, fmt.Errorf("user not found with username: %s", username)
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return user, nil
}

// UpdateUser updates an existing user
func (s *userService) UpdateUser(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
	// Get existing user to preserve system fields
	existing, err := s.userRepo.Get(ctx, user.Name)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, fmt.Errorf("user not found: %s", user.Name)
		}
		return nil, fmt.Errorf("failed to get user for update: %w", err)
	}

	// Check if 'user' role changed
	hadUserRole := s.hasUserRole(existing)
	hasUserRole := s.hasUserRole(user)

	// Update the spec while preserving metadata
	existing.Spec = user.Spec

	// Update user in repository
	if err := s.userRepo.Update(ctx, existing); err != nil {
		if repository.IsConflict(err) {
			return nil, fmt.Errorf("user has been modified by another process, please retry")
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Handle WorkMachine based on role changes
	if !hadUserRole && hasUserRole {
		// Role added: create WorkMachine
		if err := s.createWorkMachineForUser(ctx, existing); err != nil {
			fmt.Printf("Warning: Failed to create WorkMachine for user %s: %v\n", existing.Spec.Email, err)
		}
	} else if hadUserRole && !hasUserRole {
		// Role removed: delete WorkMachine
		if err := s.deleteWorkMachineForUser(ctx, existing); err != nil {
			fmt.Printf("Warning: Failed to delete WorkMachine for user %s: %v\n", existing.Spec.Email, err)
		}
	}

	return existing, nil
}

// DeleteUser deletes a user
func (s *userService) DeleteUser(ctx context.Context, name string) error {
	if err := s.userRepo.Delete(ctx, name); err != nil {
		if repository.IsNotFound(err) {
			return fmt.Errorf("user not found: %s", name)
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ResetUserPassword resets a user's password
func (s *userService) ResetUserPassword(ctx context.Context, name, newPassword string) error {
	// Get the user first
	user, err := s.userRepo.Get(ctx, name)
	if err != nil {
		if repository.IsNotFound(err) {
			return fmt.Errorf("user not found: %s", name)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Set the passwordString field - controller will hash it and update spec.password
	user.Spec.PasswordString = newPassword

	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	return nil
}

// ListUsers lists users with optional filters
func (s *userService) ListUsers(ctx context.Context, opts ...repository.ListOption) (*platformv1alpha1.UserList, error) {
	userList, err := s.userRepo.List(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return userList, nil
}

// UpdateUserLastLogin updates the last login timestamp in the user's status
func (s *userService) UpdateUserLastLogin(ctx context.Context, name string) error {
	// Get the user
	user, err := s.userRepo.Get(ctx, name)
	if err != nil {
		if repository.IsNotFound(err) {
			return fmt.Errorf("user not found: %s", name)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Update the last login time in status
	now := metav1.Now()
	user.Status.LastLogin = &now

	// Update user status in repository
	if err := s.userRepo.UpdateStatus(ctx, user); err != nil {
		return fmt.Errorf("failed to update user last login: %w", err)
	}

	return nil
}

// createWorkMachineForUser creates a WorkMachine for a newly created user
func (s *userService) createWorkMachineForUser(ctx context.Context, user *platformv1alpha1.User) error {
	// Extract username from email (part before @)
	username := user.Spec.Email
	if idx := strings.Index(username, "@"); idx > 0 {
		username = username[:idx]
	}
	// Replace dots and special characters with hyphens for valid k8s names
	username = strings.ReplaceAll(username, ".", "-")
	username = strings.ReplaceAll(username, "_", "-")
	username = strings.ToLower(username)

	// Create WorkMachine name and targetNamespace
	workMachineName := fmt.Sprintf("wm-%s", username)
	targetNamespace := fmt.Sprintf("wm-%s", username)

	// Create WorkMachine object
	workMachine := &machinesv1.WorkMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name: workMachineName,
			Labels: map[string]string{
				"kloudlite.io/user-email": sanitizeForLabel(user.Spec.Email),
				"kloudlite.io/managed":    "true",
			},
		},
		Spec: machinesv1.WorkMachineSpec{
			OwnedBy:         user.Spec.Email,
			MachineType:     "standard-2vcpu-4gb", // Default machine type
			TargetNamespace: targetNamespace,
			State:           machinesv1.MachineStateStopped,
		},
	}

	// Check if WorkMachine already exists
	existing, err := s.workMachineRepo.Get(ctx, workMachineName)
	if err == nil && existing != nil {
		// WorkMachine already exists, skip creation
		return nil
	}

	// Create the WorkMachine
	if err := s.workMachineRepo.Create(ctx, workMachine); err != nil {
		if repository.IsAlreadyExists(err) {
			// WorkMachine already exists, that's fine
			return nil
		}
		return fmt.Errorf("failed to create WorkMachine: %w", err)
	}

	return nil
}

// hasUserRole checks if a user has the 'user' role
func (s *userService) hasUserRole(user *platformv1alpha1.User) bool {
	for _, role := range user.Spec.Roles {
		if role == platformv1alpha1.RoleUser {
			return true
		}
	}
	return false
}

// deleteWorkMachineForUser deletes the WorkMachine for a user
func (s *userService) deleteWorkMachineForUser(ctx context.Context, user *platformv1alpha1.User) error {
	// Extract username from email (part before @)
	username := user.Spec.Email
	if idx := strings.Index(username, "@"); idx > 0 {
		username = username[:idx]
	}
	// Replace dots and special characters with hyphens for valid k8s names
	username = strings.ReplaceAll(username, ".", "-")
	username = strings.ReplaceAll(username, "_", "-")
	username = strings.ToLower(username)

	// Create WorkMachine name
	workMachineName := fmt.Sprintf("wm-%s", username)

	// Check if WorkMachine exists
	existing, err := s.workMachineRepo.Get(ctx, workMachineName)
	if err != nil {
		if repository.IsNotFound(err) {
			// WorkMachine doesn't exist, nothing to delete
			return nil
		}
		return fmt.Errorf("failed to get WorkMachine: %w", err)
	}

	// Delete the WorkMachine
	if err := s.workMachineRepo.Delete(ctx, existing.Name); err != nil {
		if repository.IsNotFound(err) {
			// WorkMachine already deleted, that's fine
			return nil
		}
		return fmt.Errorf("failed to delete WorkMachine: %w", err)
	}

	return nil
}

// sanitizeForLabel sanitizes a string to be used as a label value
func sanitizeForLabel(value string) string {
	// Replace special characters with hyphens for label value
	sanitized := strings.ReplaceAll(value, "@", "-at-")
	sanitized = strings.ReplaceAll(sanitized, ".", "-dot-")
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	sanitized = strings.ToLower(sanitized)

	// Ensure it starts and ends with alphanumeric
	sanitized = strings.Trim(sanitized, "-")

	// Limit length to 63 characters (Kubernetes label value limit)
	if len(sanitized) > 63 {
		sanitized = sanitized[:63]
	}

	return sanitized
}

// ActivateUser activates a user account
func (s *userService) ActivateUser(ctx context.Context, name string) (*platformv1alpha1.User, error) {
	// Use patch to update only the active field
	active := true
	patchData := map[string]interface{}{
		"spec": map[string]interface{}{
			"active": active,
		},
	}

	user, err := s.userRepo.Patch(ctx, name, patchData)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, fmt.Errorf("user not found: %s", name)
		}
		if repository.IsConflict(err) {
			return nil, fmt.Errorf("user has been modified by another process, please retry")
		}
		return nil, fmt.Errorf("failed to activate user: %w", err)
	}
	return user, nil
}

// DeactivateUser deactivates a user account
func (s *userService) DeactivateUser(ctx context.Context, name string) (*platformv1alpha1.User, error) {
	// Use patch to update only the active field
	active := false
	patchData := map[string]interface{}{
		"spec": map[string]interface{}{
			"active": active,
		},
	}

	user, err := s.userRepo.Patch(ctx, name, patchData)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, fmt.Errorf("user not found: %s", name)
		}
		if repository.IsConflict(err) {
			return nil, fmt.Errorf("user has been modified by another process, please retry")
		}
		return nil, fmt.Errorf("failed to deactivate user: %w", err)
	}
	return user, nil
}
