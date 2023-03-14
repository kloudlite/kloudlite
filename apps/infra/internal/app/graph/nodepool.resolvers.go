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

func (r *nodePoolResolver) Spec(ctx context.Context, obj *entities.NodePool) (*model.NodePoolSpec, error) {
	var m model.NodePoolSpec
	if err := fn.JsonConversion(obj.Spec, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *nodePoolInResolver) Spec(ctx context.Context, obj *entities.NodePool, data *model.NodePoolSpecIn) error {
	if obj == nil {
		return nil
	}
	return fn.JsonConversion(data, &obj.Spec)
}

// NodePool returns generated.NodePoolResolver implementation.
func (r *Resolver) NodePool() generated.NodePoolResolver { return &nodePoolResolver{r} }

// NodePoolIn returns generated.NodePoolInResolver implementation.
func (r *Resolver) NodePoolIn() generated.NodePoolInResolver { return &nodePoolInResolver{r} }

type nodePoolResolver struct{ *Resolver }
type nodePoolInResolver struct{ *Resolver }
