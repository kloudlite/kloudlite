package repository

import (
	"context"

	environmentsv1 "github.com/kloudlite/kloudlite/v2/api/pkg/apis/environments/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CompositionRepository provides operations for Composition resources (namespace-scoped)
type CompositionRepository interface {
	NamespacedRepository[*environmentsv1.Composition, *environmentsv1.CompositionList]

	// Domain-specific methods
	GetByEnvironment(ctx context.Context, environmentName string, namespace string) (*environmentsv1.CompositionList, error)
	ListByState(ctx context.Context, namespace string, state environmentsv1.CompositionState) (*environmentsv1.CompositionList, error)
	UpdateState(ctx context.Context, name string, namespace string, state environmentsv1.CompositionState, message string) error
}

// compositionRepository implements CompositionRepository
type compositionRepository struct {
	NamespacedRepository[*environmentsv1.Composition, *environmentsv1.CompositionList]
	client client.Client
}

// NewCompositionRepository creates a new CompositionRepository
func NewCompositionRepository(k8sClient client.Client) CompositionRepository {
	baseRepo := NewK8sNamespacedRepository(
		k8sClient,
		func() *environmentsv1.Composition { return &environmentsv1.Composition{} },
		func() *environmentsv1.CompositionList { return &environmentsv1.CompositionList{} },
	)

	return &compositionRepository{
		NamespacedRepository: baseRepo,
		client:               k8sClient,
	}
}

// GetByEnvironment retrieves all compositions in a specific environment namespace
func (r *compositionRepository) GetByEnvironment(ctx context.Context, environmentName string, namespace string) (*environmentsv1.CompositionList, error) {
	// Since compositions are namespace-scoped and the namespace represents the environment,
	// we can simply list all compositions in that namespace
	return r.List(ctx, namespace)
}

// ListByState retrieves all compositions with a specific state
func (r *compositionRepository) ListByState(ctx context.Context, namespace string, state environmentsv1.CompositionState) (*environmentsv1.CompositionList, error) {
	// List all compositions and filter by state
	allCompositions, err := r.List(ctx, namespace)
	if err != nil {
		return nil, err
	}

	result := &environmentsv1.CompositionList{}
	for _, comp := range allCompositions.Items {
		if comp.Status.State == state {
			result.Items = append(result.Items, comp)
		}
	}

	return result, nil
}

// UpdateState updates the state of a composition
func (r *compositionRepository) UpdateState(ctx context.Context, name string, namespace string, state environmentsv1.CompositionState, message string) error {
	// Get the composition
	composition, err := r.Get(ctx, namespace, name)
	if err != nil {
		return err
	}

	// Update the state and message in status
	composition.Status.State = state
	composition.Status.Message = message

	// Use status subresource to update only status
	return r.client.Status().Update(ctx, composition)
}
