package repository

import (
	"context"

	environmentsv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/environments/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EnvironmentRepository provides operations for Environment resources (cluster-scoped)
type EnvironmentRepository interface {
	ClusterRepository[*environmentsv1.Environment, *environmentsv1.EnvironmentList]

	// Domain-specific methods
	GetByNamespace(ctx context.Context, namespace string) (*environmentsv1.Environment, error)
	ListActive(ctx context.Context) (*environmentsv1.EnvironmentList, error)
	ListInactive(ctx context.Context) (*environmentsv1.EnvironmentList, error)
	ActivateEnvironment(ctx context.Context, name string) error
	DeactivateEnvironment(ctx context.Context, name string) error
}

// environmentRepository implements EnvironmentRepository
type environmentRepository struct {
	ClusterRepository[*environmentsv1.Environment, *environmentsv1.EnvironmentList]
	client client.Client
}

// NewEnvironmentRepository creates a new EnvironmentRepository
func NewEnvironmentRepository(k8sClient client.Client) EnvironmentRepository {
	baseRepo := NewK8sClusterRepository(
		k8sClient,
		func() *environmentsv1.Environment { return &environmentsv1.Environment{} },
		func() *environmentsv1.EnvironmentList { return &environmentsv1.EnvironmentList{} },
	)

	return &environmentRepository{
		ClusterRepository: baseRepo,
		client:            k8sClient,
	}
}

// GetByNamespace retrieves an environment by its target namespace
func (r *environmentRepository) GetByNamespace(ctx context.Context, namespace string) (*environmentsv1.Environment, error) {
	// Use field selector to find environment by target namespace
	envs, err := r.List(ctx, WithFieldSelector("spec.targetNamespace="+namespace))
	if err != nil {
		return nil, err
	}

	if len(envs.Items) == 0 {
		return nil, ErrNotFound("environment with target namespace " + namespace + " not found")
	}

	if len(envs.Items) > 1 {
		return nil, ErrMultipleFound("multiple environments found with target namespace " + namespace)
	}

	return &envs.Items[0], nil
}

// ListActive retrieves all active environments
func (r *environmentRepository) ListActive(ctx context.Context) (*environmentsv1.EnvironmentList, error) {
	// Use field selector to find active environments
	return r.List(ctx, WithFieldSelector("spec.activated=true"))
}

// ListInactive retrieves all inactive environments
func (r *environmentRepository) ListInactive(ctx context.Context) (*environmentsv1.EnvironmentList, error) {
	// Use field selector to find inactive environments
	return r.List(ctx, WithFieldSelector("spec.activated=false"))
}

// ActivateEnvironment activates an environment by name
func (r *environmentRepository) ActivateEnvironment(ctx context.Context, name string) error {
	// Get the environment (cluster-scoped)
	env, err := r.Get(ctx, name)
	if err != nil {
		return err
	}

	// Check if already activated
	if env.Spec.Activated {
		return nil // Already activated
	}

	// Update the activation status
	env.Spec.Activated = true
	env.Status.State = environmentsv1.EnvironmentStateActivating

	// Update the environment
	if err := r.Update(ctx, env); err != nil {
		return err
	}

	// Update status
	return r.client.Status().Update(ctx, env)
}

// DeactivateEnvironment deactivates an environment by name
func (r *environmentRepository) DeactivateEnvironment(ctx context.Context, name string) error {
	// Get the environment (cluster-scoped)
	env, err := r.Get(ctx, name)
	if err != nil {
		return err
	}

	// Check if already deactivated
	if !env.Spec.Activated {
		return nil // Already deactivated
	}

	// Update the activation status
	env.Spec.Activated = false
	env.Status.State = environmentsv1.EnvironmentStateDeactivating

	// Update the environment
	if err := r.Update(ctx, env); err != nil {
		return err
	}

	// Update status
	return r.client.Status().Update(ctx, env)
}