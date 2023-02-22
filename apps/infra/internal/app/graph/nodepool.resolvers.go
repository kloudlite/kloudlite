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

func (r *nodePoolResolver) Metadata(ctx context.Context, obj *entities.NodePool) (*v1.ObjectMeta, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *nodePoolResolver) Spec(ctx context.Context, obj *entities.NodePool) (*model.NodePoolSpec, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *nodePoolInResolver) Metadata(ctx context.Context, obj *entities.NodePool, data *v1.ObjectMeta) error {
	panic(fmt.Errorf("not implemented"))
}

func (r *nodePoolInResolver) Spec(ctx context.Context, obj *entities.NodePool, data *model.NodePoolSpecIn) error {
	panic(fmt.Errorf("not implemented"))
}

// NodePool returns generated.NodePoolResolver implementation.
func (r *Resolver) NodePool() generated.NodePoolResolver { return &nodePoolResolver{r} }

// NodePoolIn returns generated.NodePoolInResolver implementation.
func (r *Resolver) NodePoolIn() generated.NodePoolInResolver { return &nodePoolInResolver{r} }

type nodePoolResolver struct{ *Resolver }
type nodePoolInResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//   - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//     it when you're done.
//   - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *nodePoolResolver) Status(ctx context.Context, obj *entities.NodePool) (*operator.Status, error) {
	panic(fmt.Errorf("not implemented"))
}
