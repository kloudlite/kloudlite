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

func (r *masterNodeResolver) Metadata(ctx context.Context, obj *entities.MasterNode) (*v1.ObjectMeta, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *masterNodeResolver) Spec(ctx context.Context, obj *entities.MasterNode) (*model.MasterNodeSpec, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *masterNodeInResolver) Metadata(ctx context.Context, obj *entities.MasterNode, data *v1.ObjectMeta) error {
	panic(fmt.Errorf("not implemented"))
}

func (r *masterNodeInResolver) Spec(ctx context.Context, obj *entities.MasterNode, data *model.MasterNodeSpecIn) error {
	panic(fmt.Errorf("not implemented"))
}

// MasterNode returns generated.MasterNodeResolver implementation.
func (r *Resolver) MasterNode() generated.MasterNodeResolver { return &masterNodeResolver{r} }

// MasterNodeIn returns generated.MasterNodeInResolver implementation.
func (r *Resolver) MasterNodeIn() generated.MasterNodeInResolver { return &masterNodeInResolver{r} }

type masterNodeResolver struct{ *Resolver }
type masterNodeInResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//   - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//     it when you're done.
//   - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *masterNodeResolver) Status(ctx context.Context, obj *entities.MasterNode) (*operator.Status, error) {
	panic(fmt.Errorf("not implemented"))
}
