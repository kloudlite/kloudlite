package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/domain/entities"
	fn "kloudlite.io/pkg/functions"
)

func (r *configResolver) Data(ctx context.Context, obj *entities.Config) (map[string]interface{}, error) {
	m := make(map[string]any, len(obj.Data))
	if err := fn.JsonConversion(obj.Data, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *configInResolver) Data(ctx context.Context, obj *entities.Config, data map[string]interface{}) error {
	return fn.JsonConversion(data, &obj.Data)
}

// Config returns generated.ConfigResolver implementation.
func (r *Resolver) Config() generated.ConfigResolver { return &configResolver{r} }

// ConfigIn returns generated.ConfigInResolver implementation.
func (r *Resolver) ConfigIn() generated.ConfigInResolver { return &configInResolver{r} }

type configResolver struct{ *Resolver }
type configInResolver struct{ *Resolver }
