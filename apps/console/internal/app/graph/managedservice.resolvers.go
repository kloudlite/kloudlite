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

func (r *managedServiceResolver) Spec(ctx context.Context, obj *entities.MSvc) (*model.ManagedServiceSpec, error) {
	var m model.ManagedServiceSpec
	if err := fn.JsonConversion(obj.Spec, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *managedServiceInResolver) Spec(ctx context.Context, obj *entities.MSvc, data *model.ManagedServiceSpecIn) error {
	if err := fn.JsonConversion(data, &obj.Spec); err != nil {
		return err
	}
	return nil
}

// ManagedService returns generated.ManagedServiceResolver implementation.
func (r *Resolver) ManagedService() generated.ManagedServiceResolver {
	return &managedServiceResolver{r}
}

// ManagedServiceIn returns generated.ManagedServiceInResolver implementation.
func (r *Resolver) ManagedServiceIn() generated.ManagedServiceInResolver {
	return &managedServiceInResolver{r}
}

type managedServiceResolver struct{ *Resolver }
type managedServiceInResolver struct{ *Resolver }
