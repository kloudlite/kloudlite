package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/app/graph/model"
	"kloudlite.io/apps/console/internal/domain/entities"
	fn "kloudlite.io/pkg/functions"
)

func (r *managedResourceResolver) Spec(ctx context.Context, obj *entities.MRes) (*model.ManagedResourceSpec, error) {
	if obj == nil {
		return nil, nil
	}
	var m model.ManagedResourceSpec
	if err := fn.JsonConversion(obj.Spec, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *managedResourceInResolver) Spec(ctx context.Context, obj *entities.MRes, data *model.ManagedResourceSpecIn) error {
	if obj == nil {
		return nil
	}
	return fn.JsonConversion(data, &obj.Spec)
}

// ManagedResource returns generated.ManagedResourceResolver implementation.
func (r *Resolver) ManagedResource() generated.ManagedResourceResolver {
	return &managedResourceResolver{r}
}

// ManagedResourceIn returns generated.ManagedResourceInResolver implementation.
func (r *Resolver) ManagedResourceIn() generated.ManagedResourceInResolver {
	return &managedResourceInResolver{r}
}

type managedResourceResolver struct{ *Resolver }
type managedResourceInResolver struct{ *Resolver }
