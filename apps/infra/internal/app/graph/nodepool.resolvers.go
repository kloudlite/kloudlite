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

func (r *nodePoolResolver) Status(ctx context.Context, obj *entities.NodePool) (*model.Status, error) {
	if obj == nil {
		return nil, nil
	}
	var status model.Status
	if err := fn.JsonConversion(obj.Status, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// NodePool returns generated.NodePoolResolver implementation.
func (r *Resolver) NodePool() generated.NodePoolResolver { return &nodePoolResolver{r} }

type nodePoolResolver struct{ *Resolver }
