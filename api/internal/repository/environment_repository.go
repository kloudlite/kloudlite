package repository

import (
	"context"

	environmentsv1 "github.com/kloudlite/kloudlite/api/internal/controllers/environment/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EnvironmentRepository provides operations for Environment resources (namespace-scoped)
type EnvironmentRepository interface {
	NamespacedRepository[*environmentsv1.Environment, *environmentsv1.EnvironmentList]

	// Domain-specific methods
	GetByTargetNamespace(ctx context.Context, targetNamespace string) (*environmentsv1.Environment, error)
	ListActive(ctx context.Context, namespace string) (*environmentsv1.EnvironmentList, error)
	ListInactive(ctx context.Context, namespace string) (*environmentsv1.EnvironmentList, error)
	ActivateEnvironment(ctx context.Context, namespace, name string) error
	DeactivateEnvironment(ctx context.Context, namespace, name string) error
}

// environmentRepository implements EnvironmentRepository
type environmentRepository struct {
	NamespacedRepository[*environmentsv1.Environment, *environmentsv1.EnvironmentList]
	client client.WithWatch
}

// NewEnvironmentRepository creates a new EnvironmentRepository
func NewEnvironmentRepository(k8sClient client.WithWatch) EnvironmentRepository {
	baseRepo := NewK8sNamespacedRepository(
		k8sClient,
		func() *environmentsv1.Environment { return &environmentsv1.Environment{} },
		func() *environmentsv1.EnvironmentList { return &environmentsv1.EnvironmentList{} },
	)

	return &environmentRepository{
		NamespacedRepository: baseRepo,
		client:               k8sClient,
	}
}

// GetByTargetNamespace retrieves an environment by its target namespace
// This searches across all namespaces since targetNamespace is unique
func (r *environmentRepository) GetByTargetNamespace(ctx context.Context, targetNamespace string) (*environmentsv1.Environment, error) {
	// Use label selector for efficient server-side filtering across all namespaces
	envs, err := r.List(ctx, "", WithLabelSelector("kloudlite.io/target-namespace="+targetNamespace))
	if err != nil {
		return nil, err
	}

	if len(envs.Items) == 0 {
		return nil, ErrNotFound("environment with target namespace " + targetNamespace + " not found")
	}

	if len(envs.Items) > 1 {
		return nil, ErrMultipleFound("multiple environments found with target namespace " + targetNamespace)
	}

	return &envs.Items[0], nil
}

// ListActive retrieves all active environments in a namespace
func (r *environmentRepository) ListActive(ctx context.Context, namespace string) (*environmentsv1.EnvironmentList, error) {
	// Use label selector for efficient server-side filtering
	return r.List(ctx, namespace, WithLabelSelector("kloudlite.io/activated=true"))
}

// ListInactive retrieves all inactive environments in a namespace
func (r *environmentRepository) ListInactive(ctx context.Context, namespace string) (*environmentsv1.EnvironmentList, error) {
	// Use label selector for efficient server-side filtering
	return r.List(ctx, namespace, WithLabelSelector("kloudlite.io/activated=false"))
}

// ActivateEnvironment activates an environment by namespace and name
func (r *environmentRepository) ActivateEnvironment(ctx context.Context, namespace, name string) error {
	// Get the environment (namespace-scoped)
	env, err := r.Get(ctx, namespace, name)
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

// DeactivateEnvironment deactivates an environment by namespace and name
func (r *environmentRepository) DeactivateEnvironment(ctx context.Context, namespace, name string) error {
	// Get the environment (namespace-scoped)
	env, err := r.Get(ctx, namespace, name)
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
