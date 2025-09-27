package services

import (
	"context"
	"fmt"

	platformv1alpha1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/platform/v1alpha1"
	"github.com/kloudlite/kloudlite/v2/api/internal/repository"
)

// UserService provides business logic for User operations
type UserService interface {
	// CRUD operations
	CreateUser(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error)
	GetUser(ctx context.Context, name, namespace string) (*platformv1alpha1.User, error)
	GetUserByEmail(ctx context.Context, email string) (*platformv1alpha1.User, error)
	GetUserByUsername(ctx context.Context, username string) (*platformv1alpha1.User, error)
	UpdateUser(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error)
	DeleteUser(ctx context.Context, name, namespace string) error
	ListUsers(ctx context.Context, namespace string, opts ...repository.ListOption) (*platformv1alpha1.UserList, error)

	// Business operations
	ActivateUser(ctx context.Context, name, namespace string) error
	DeactivateUser(ctx context.Context, name, namespace string) error
}

// userService implements UserService
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService creates a new UserService
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, user *platformv1alpha1.User) (*platformv1alpha1.User, error) {
	// Set namespace if not provided
	if user.Namespace == "" {
		user.Namespace = "default"
	}

	// All validations and mutations are handled by webhooks
	// Just create the user in repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		if repository.IsAlreadyExists(err) {
			return nil, fmt.Errorf("user already exists: %w", err)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetUser retrieves a user by name and namespace
func (s *userService) GetUser(ctx context.Context, name, namespace string) (*platformv1alpha1.User, error) {
	user, err := s.userRepo.Get(ctx, name, namespace)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, fmt.Errorf("user not found: %s/%s", namespace, name)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// GetUserByEmail retrieves a user by email address
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*platformv1alpha1.User, error) {
	// This would need to iterate through users or use a label selector
	// The webhook adds labels for efficient lookup
	users, err := s.userRepo.List(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	for _, user := range users.Items {
		if user.Spec.Email == email {
			return &user, nil
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
	existing, err := s.userRepo.Get(ctx, user.Name, user.Namespace)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, fmt.Errorf("user not found: %s/%s", user.Namespace, user.Name)
		}
		return nil, fmt.Errorf("failed to get user for update: %w", err)
	}

	// Update the spec while preserving metadata
	existing.Spec = user.Spec

	// Update user in repository
	if err := s.userRepo.Update(ctx, existing); err != nil {
		if repository.IsConflict(err) {
			return nil, fmt.Errorf("user has been modified by another process, please retry")
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return existing, nil
}

// DeleteUser deletes a user
func (s *userService) DeleteUser(ctx context.Context, name, namespace string) error {
	if err := s.userRepo.Delete(ctx, name, namespace); err != nil {
		if repository.IsNotFound(err) {
			return fmt.Errorf("user not found: %s/%s", namespace, name)
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// ListUsers lists users with optional filters
func (s *userService) ListUsers(ctx context.Context, namespace string, opts ...repository.ListOption) (*platformv1alpha1.UserList, error) {
	userList, err := s.userRepo.List(ctx, namespace, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return userList, nil
}

// ActivateUser activates a user
func (s *userService) ActivateUser(ctx context.Context, name, namespace string) error {
	user, err := s.GetUser(ctx, name, namespace)
	if err != nil {
		return err
	}

	user.Spec.Active = &[]bool{true}[0]
	_, err = s.UpdateUser(ctx, user)
	return err
}

// DeactivateUser deactivates a user
func (s *userService) DeactivateUser(ctx context.Context, name, namespace string) error {
	user, err := s.GetUser(ctx, name, namespace)
	if err != nil {
		return err
	}

	user.Spec.Active = &[]bool{false}[0]
	_, err = s.UpdateUser(ctx, user)
	return err
}

