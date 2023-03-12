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

func (r *appResolver) Spec(ctx context.Context, obj *entities.App) (*model.AppSpec, error) {
	var app model.AppSpec
	if err := fn.JsonConversion(obj.App, &app); err != nil {
		return nil, err
	}
	return &app, nil
}

func (r *appInResolver) Spec(ctx context.Context, obj *entities.App, data *model.AppSpecIn) error {
	return fn.JsonConversion(data, &obj.App)
}

// App returns generated.AppResolver implementation.
func (r *Resolver) App() generated.AppResolver { return &appResolver{r} }

// AppIn returns generated.AppInResolver implementation.
func (r *Resolver) AppIn() generated.AppInResolver { return &appInResolver{r} }

type appResolver struct{ *Resolver }
type appInResolver struct{ *Resolver }
