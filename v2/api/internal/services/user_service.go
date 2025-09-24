package services

import (
	"context"
	"fmt"
	"time"

	platformv1alpha1 "github.com/kloudlite/api/v2/pkg/apis/platform/v1alpha1"
	"github.com/kloudlite/api/v2/internal/repository"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// UserService provides business logic for User operations
type UserService interface {
	// CRUD operations
	CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
	GetUser(ctx context.Context, name, namespace string) (*UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*UserResponse, error)
	GetUserByUsername(ctx context.Context, username string) (*UserResponse, error)
	UpdateUser(ctx context.Context, name, namespace string, req *UpdateUserRequest) (*UserResponse, error)
	DeleteUser(ctx context.Context, name, namespace string) error
	ListUsers(ctx context.Context, namespace string, req *ListUsersRequest) (*ListUsersResponse, error)

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

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=50"`
	Namespace   string `json:"namespace,omitempty"`
	Email       string `json:"email" validate:"required,email"`
	Username    string `json:"username" validate:"required,min=3,max=30"`
	DisplayName string `json:"displayName,omitempty" validate:"max=100"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
	Role        string `json:"role,omitempty" validate:"oneof=admin developer viewer"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Email       *string `json:"email,omitempty" validate:"omitempty,email"`
	Username    *string `json:"username,omitempty" validate:"omitempty,min=3,max=30"`
	DisplayName *string `json:"displayName,omitempty" validate:"omitempty,max=100"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
	Role        *string `json:"role,omitempty" validate:"omitempty,oneof=admin developer viewer"`
	Active      *bool   `json:"active,omitempty"`
}

// ListUsersRequest represents a request to list users
type ListUsersRequest struct {
	Namespace     string `json:"namespace,omitempty"`
	LabelSelector string `json:"labelSelector,omitempty"`
	FieldSelector string `json:"fieldSelector,omitempty"`
	Limit         int64  `json:"limit,omitempty"`
	Continue      string `json:"continue,omitempty"`
}

// UserResponse represents a user response
type UserResponse struct {
	Name        string             `json:"name"`
	Namespace   string             `json:"namespace"`
	Email       string             `json:"email"`
	Username    string             `json:"username"`
	DisplayName string             `json:"displayName,omitempty"`
	AvatarURL   string             `json:"avatarUrl,omitempty"`
	Role        string             `json:"role,omitempty"`
	Active      bool               `json:"active"`
	Phase       string             `json:"phase,omitempty"`
	LastLogin   *time.Time         `json:"lastLogin,omitempty"`
	CreatedAt   *time.Time         `json:"createdAt,omitempty"`
	Labels      map[string]string  `json:"labels,omitempty"`
	Annotations map[string]string  `json:"annotations,omitempty"`
}

// ListUsersResponse represents a list users response
type ListUsersResponse struct {
	Items    []UserResponse `json:"items"`
	Continue string         `json:"continue,omitempty"`
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error) {
	if err := validateCreateUserRequest(req); err != nil {
		return nil, err
	}

	// Set default namespace if not provided
	namespace := req.Namespace
	if namespace == "" {
		namespace = "default"
	}

	// Set default role if not provided
	role := req.Role
	if role == "" {
		role = "developer"
	}

	// Create User object
	user := &platformv1alpha1.User{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: namespace,
		},
		Spec: platformv1alpha1.UserSpec{
			Email:       req.Email,
			Username:    req.Username,
			DisplayName: req.DisplayName,
			AvatarURL:   req.AvatarURL,
			Role:        role,
			Active:      &[]bool{true}[0], // Default to active
		},
	}

	// Create user in repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		if repository.IsAlreadyExists(err) {
			return nil, fmt.Errorf("user with name %s already exists in namespace %s", req.Name, namespace)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.userToResponse(user), nil
}

// GetUser retrieves a user by name and namespace
func (s *userService) GetUser(ctx context.Context, name, namespace string) (*UserResponse, error) {
	user, err := s.userRepo.Get(ctx, name, namespace)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, fmt.Errorf("user not found: %s/%s", namespace, name)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return s.userToResponse(user), nil
}

// GetUserByEmail retrieves a user by email address
func (s *userService) GetUserByEmail(ctx context.Context, email string) (*UserResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, fmt.Errorf("user not found with email: %s", email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return s.userToResponse(user), nil
}

// GetUserByUsername retrieves a user by username
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*UserResponse, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, fmt.Errorf("user not found with username: %s", username)
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return s.userToResponse(user), nil
}

// UpdateUser updates an existing user
func (s *userService) UpdateUser(ctx context.Context, name, namespace string, req *UpdateUserRequest) (*UserResponse, error) {
	if err := validateUpdateUserRequest(req); err != nil {
		return nil, err
	}

	// Get existing user
	user, err := s.userRepo.Get(ctx, name, namespace)
	if err != nil {
		if repository.IsNotFound(err) {
			return nil, fmt.Errorf("user not found: %s/%s", namespace, name)
		}
		return nil, fmt.Errorf("failed to get user for update: %w", err)
	}

	// Apply updates
	if req.Email != nil {
		user.Spec.Email = *req.Email
	}
	if req.Username != nil {
		user.Spec.Username = *req.Username
	}
	if req.DisplayName != nil {
		user.Spec.DisplayName = *req.DisplayName
	}
	if req.AvatarURL != nil {
		user.Spec.AvatarURL = *req.AvatarURL
	}
	if req.Role != nil {
		user.Spec.Role = *req.Role
	}
	if req.Active != nil {
		user.Spec.Active = req.Active
	}

	// Update user in repository
	if err := s.userRepo.Update(ctx, user); err != nil {
		if repository.IsConflict(err) {
			return nil, fmt.Errorf("user has been modified by another process, please retry")
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.userToResponse(user), nil
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
func (s *userService) ListUsers(ctx context.Context, namespace string, req *ListUsersRequest) (*ListUsersResponse, error) {
	if req == nil {
		req = &ListUsersRequest{}
	}

	// Build repository options
	var opts []repository.ListOption
	if req.LabelSelector != "" {
		opts = append(opts, repository.WithLabelSelector(req.LabelSelector))
	}
	if req.FieldSelector != "" {
		opts = append(opts, repository.WithFieldSelector(req.FieldSelector))
	}
	if req.Limit > 0 {
		opts = append(opts, repository.WithLimit(req.Limit))
	}
	if req.Continue != "" {
		opts = append(opts, repository.WithContinue(req.Continue))
	}

	// List users from repository
	userList, err := s.userRepo.List(ctx, namespace, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Convert to response
	response := &ListUsersResponse{
		Items: make([]UserResponse, len(userList.Items)),
	}

	for i, user := range userList.Items {
		response.Items[i] = *s.userToResponse(&user)
	}

	// Set continue token if available
	if userList.ListMeta.Continue != "" {
		response.Continue = userList.ListMeta.Continue
	}

	return response, nil
}

// ActivateUser activates a user
func (s *userService) ActivateUser(ctx context.Context, name, namespace string) error {
	_, err := s.UpdateUser(ctx, name, namespace, &UpdateUserRequest{
		Active: &[]bool{true}[0],
	})
	return err
}

// DeactivateUser deactivates a user
func (s *userService) DeactivateUser(ctx context.Context, name, namespace string) error {
	_, err := s.UpdateUser(ctx, name, namespace, &UpdateUserRequest{
		Active: &[]bool{false}[0],
	})
	return err
}

// userToResponse converts a User object to UserResponse
func (s *userService) userToResponse(user *platformv1alpha1.User) *UserResponse {
	response := &UserResponse{
		Name:        user.Name,
		Namespace:   user.Namespace,
		Email:       user.Spec.Email,
		Username:    user.Spec.Username,
		DisplayName: user.Spec.DisplayName,
		AvatarURL:   user.Spec.AvatarURL,
		Role:        user.Spec.Role,
		Active:      user.Spec.Active != nil && *user.Spec.Active,
		Phase:       user.Status.Phase,
		Labels:      user.Labels,
		Annotations: user.Annotations,
	}

	// Convert timestamps
	if user.Status.LastLogin != nil {
		t := user.Status.LastLogin.Time
		response.LastLogin = &t
	}
	if user.Status.CreatedAt != nil {
		t := user.Status.CreatedAt.Time
		response.CreatedAt = &t
	}

	return response
}

// validateCreateUserRequest validates create user request
func validateCreateUserRequest(req *CreateUserRequest) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	// Add more validation as needed
	return nil
}

// validateUpdateUserRequest validates update user request
func validateUpdateUserRequest(req *UpdateUserRequest) error {
	// Add validation as needed
	return nil
}