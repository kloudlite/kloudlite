package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	v11 "github.com/kloudlite/cluster-operator/apis/infra/v1"
	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/app/graph/model"
	"kloudlite.io/apps/infra/internal/domain/entities"
)

func (r *workerNodeResolver) Status(ctx context.Context, obj *entities.WorkerNode) (*model.Status, error) {
	if obj == nil {
		return nil, nil
	}
	return toModelStatus(obj.Status)
}

func (r *workerNodeSpecResolver) NodeIndex(ctx context.Context, obj *v11.WorkerNodeSpec) (*int, error) {
	if obj == nil {
		return nil, nil
	}
	return &obj.Index, nil
}

func (r *workerNodeSpecInResolver) NodeIndex(ctx context.Context, obj *v11.WorkerNodeSpec, data *int) error {
	if obj == nil {
		return nil
	}
	obj.Index = *data
	return nil
}

// WorkerNode returns generated.WorkerNodeResolver implementation.
func (r *Resolver) WorkerNode() generated.WorkerNodeResolver { return &workerNodeResolver{r} }

// WorkerNodeSpec returns generated.WorkerNodeSpecResolver implementation.
func (r *Resolver) WorkerNodeSpec() generated.WorkerNodeSpecResolver {
	return &workerNodeSpecResolver{r}
}

// WorkerNodeSpecIn returns generated.WorkerNodeSpecInResolver implementation.
func (r *Resolver) WorkerNodeSpecIn() generated.WorkerNodeSpecInResolver {
	return &workerNodeSpecInResolver{r}
}

type workerNodeResolver struct{ *Resolver }
type workerNodeSpecResolver struct{ *Resolver }
type workerNodeSpecInResolver struct{ *Resolver }
