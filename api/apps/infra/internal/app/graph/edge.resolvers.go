package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/app/graph/model"
	"kloudlite.io/apps/infra/internal/domain/entities"
	fn "kloudlite.io/pkg/functions"
)

func (r *edgeResolver) Status(ctx context.Context, obj *entities.Edge) (*model.Status, error) {
	if obj == nil {
		return nil, nil
	}
	var status model.Status
	if err := fn.JsonConversion(obj.Status, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// Edge returns generated.EdgeResolver implementation.
func (r *Resolver) Edge() generated.EdgeResolver { return &edgeResolver{r} }

type edgeResolver struct{ *Resolver }
