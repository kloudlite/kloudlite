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

func (r *workerNodeResolver) Spec(ctx context.Context, obj *entities.WorkerNode) (*model.WorkerNodeSpec, error) {
	var m model.WorkerNodeSpec
	if err := fn.JsonConversion(obj.Spec, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *workerNodeInResolver) Spec(ctx context.Context, obj *entities.WorkerNode, data *model.WorkerNodeSpecIn) error {
	if obj == nil {
		return nil
	}
	return fn.JsonConversion(data, &obj.Spec)
}

// WorkerNode returns generated.WorkerNodeResolver implementation.
func (r *Resolver) WorkerNode() generated.WorkerNodeResolver { return &workerNodeResolver{r} }

// WorkerNodeIn returns generated.WorkerNodeInResolver implementation.
func (r *Resolver) WorkerNodeIn() generated.WorkerNodeInResolver { return &workerNodeInResolver{r} }

type workerNodeResolver struct{ *Resolver }
type workerNodeInResolver struct{ *Resolver }
