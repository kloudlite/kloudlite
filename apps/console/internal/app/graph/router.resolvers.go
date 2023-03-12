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

func (r *routerResolver) Spec(ctx context.Context, obj *entities.Router) (*model.RouterSpec, error) {
	var m model.RouterSpec
	if err := fn.JsonConversion(obj.Spec, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *routerInResolver) Spec(ctx context.Context, obj *entities.Router, data *model.RouterSpecIn) error {
	return fn.JsonConversion(data, obj.Spec)
}

// Router returns generated.RouterResolver implementation.
func (r *Resolver) Router() generated.RouterResolver { return &routerResolver{r} }

// RouterIn returns generated.RouterInResolver implementation.
func (r *Resolver) RouterIn() generated.RouterInResolver { return &routerInResolver{r} }

type routerResolver struct{ *Resolver }
type routerInResolver struct{ *Resolver }
