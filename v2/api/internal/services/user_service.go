package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	platformv1alpha1 "github.com/kloudlite/api/v2/pkg/apis/platform/v1alpha1"
	"github.com/kloudlite/api/v2/internal/repository"
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
	// Validate required fields
	if user.Spec.Email == "" {
		return nil, fmt.Errorf("email is required")
	}

	// Set defaults
	if user.Namespace == "" {
		user.Namespace = "default"
	}

	// Use GenerateName if Name is not provided
	if user.Name == "" && user.GenerateName == "" {
		user.GenerateName = "user-"
	}

	// Set default roles if not provided
	if len(user.Spec.Roles) == 0 {
		user.Spec.Roles = []string{"user"}
	}

	// Set active by default
	if user.Spec.Active == nil {
		user.Spec.Active = &[]bool{true}[0]
	}

	// Add email as a label for efficient lookups (sanitize for label requirements)
	if user.Labels == nil {
		user.Labels = make(map[string]string)
	}
	// Sanitize email for use as label value (labels must be alphanumeric, -, _, or .)
	sanitizedEmail := sanitizeEmailForLabel(user.Spec.Email)
	user.Labels["kloudlite.io/user-email-hash"] = hashEmail(user.Spec.Email)
	user.Labels["kloudlite.io/user-email"] = sanitizedEmail

	// Create user in repository
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

// GetUserByEmail retrieves a user by email address using label selector
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*platformv1alpha1.User, error) {
	// Use label selector to find user by email hash
	emailHash := hashEmail(email)
	labelSelector := fmt.Sprintf("kloudlite.io/user-email-hash=%s", emailHash)

	users, err := s.userRepo.List(ctx, "", repository.WithLabelSelector(labelSelector))
	if err != nil {
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	if len(users.Items) == 0 {
		return nil, fmt.Errorf("user not found with email: %s", email)
	}

	// Double-check the email matches (in case of hash collision)
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

// hashEmail creates a SHA256 hash of the email for use as a label
func hashEmail(email string) string {
	h := sha256.New()
	h.Write([]byte(strings.ToLower(email)))
	return hex.EncodeToString(h.Sum(nil))[:16] // Use first 16 chars for shorter label
}

// sanitizeEmailForLabel converts email to a valid label value
func sanitizeEmailForLabel(email string) string {
	// Replace @ with -at- and . with -dot- for readability
	// Remove any characters that aren't alphanumeric, -, _, or .
	email = strings.ToLower(email)
	email = strings.ReplaceAll(email, "@", "-at-")
	email = strings.ReplaceAll(email, ".", "-dot-")

	// Keep only valid label characters
	var result strings.Builder
	for _, r := range email {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			result.WriteRune(r)
		}
	}

	return result.String()
}