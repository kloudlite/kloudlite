package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/kloudlite/operator/pkg/operator"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/app/graph/model"
	"kloudlite.io/apps/infra/internal/domain/entities"
)

func (r *workerNodeResolver) Metadata(ctx context.Context, obj *entities.WorkerNode) (*v1.ObjectMeta, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *workerNodeResolver) Spec(ctx context.Context, obj *entities.WorkerNode) (*model.WorkerNodeSpec, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *workerNodeInResolver) Metadata(ctx context.Context, obj *entities.WorkerNode, data *v1.ObjectMeta) error {
	panic(fmt.Errorf("not implemented"))
}

func (r *workerNodeInResolver) Spec(ctx context.Context, obj *entities.WorkerNode, data *model.WorkerNodeSpecIn) error {
	panic(fmt.Errorf("not implemented"))
}

// WorkerNode returns generated.WorkerNodeResolver implementation.
func (r *Resolver) WorkerNode() generated.WorkerNodeResolver { return &workerNodeResolver{r} }

// WorkerNodeIn returns generated.WorkerNodeInResolver implementation.
func (r *Resolver) WorkerNodeIn() generated.WorkerNodeInResolver { return &workerNodeInResolver{r} }

type workerNodeResolver struct{ *Resolver }
type workerNodeInResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//   - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//     it when you're done.
//   - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *workerNodeResolver) Status(ctx context.Context, obj *entities.WorkerNode) (*operator.Status, error) {
	panic(fmt.Errorf("not implemented"))
}
