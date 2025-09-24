package repository

import (
	"context"

	platformv1alpha1 "github.com/kloudlite/api/v2/pkg/apis/platform/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// UserRepository provides operations for User resources
type UserRepository interface {
	Repository[*platformv1alpha1.User, *platformv1alpha1.UserList]

	// Domain-specific methods can be added here
	GetByEmail(ctx context.Context, email string) (*platformv1alpha1.User, error)
	GetByUsername(ctx context.Context, username string) (*platformv1alpha1.User, error)
	ListActive(ctx context.Context) (*platformv1alpha1.UserList, error)
}

// userRepository implements UserRepository
type userRepository struct {
	Repository[*platformv1alpha1.User, *platformv1alpha1.UserList]
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(k8sClient client.Client) UserRepository {
	baseRepo := NewK8sRepository(
		k8sClient,
		func() *platformv1alpha1.User { return &platformv1alpha1.User{} },
		func() *platformv1alpha1.UserList { return &platformv1alpha1.UserList{} },
	)

	return &userRepository{
		Repository: baseRepo,
	}
}

// GetByEmail retrieves a user by email address
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*platformv1alpha1.User, error) {
	// Use field selector to find user by email
	users, err := r.List(ctx, "", WithFieldSelector("spec.email="+email))
	if err != nil {
		return nil, err
	}

	if len(users.Items) == 0 {
		return nil, ErrNotFound("user with email " + email + " not found")
	}

	if len(users.Items) > 1 {
		return nil, ErrMultipleFound("multiple users found with email " + email)
	}

	return &users.Items[0], nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*platformv1alpha1.User, error) {
	// Use field selector to find user by username
	users, err := r.List(ctx, "", WithFieldSelector("spec.username="+username))
	if err != nil {
		return nil, err
	}

	if len(users.Items) == 0 {
		return nil, ErrNotFound("user with username " + username + " not found")
	}

	if len(users.Items) > 1 {
		return nil, ErrMultipleFound("multiple users found with username " + username)
	}

	return &users.Items[0], nil
}

// ListActive retrieves all active users
func (r *userRepository) ListActive(ctx context.Context) (*platformv1alpha1.UserList, error) {
	// Use field selector to find active users
	return r.List(ctx, "", WithFieldSelector("spec.active=true"))
}