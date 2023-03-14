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

func (r *masterNodeResolver) Spec(ctx context.Context, obj *entities.MasterNode) (*model.MasterNodeSpec, error) {
	var m model.MasterNodeSpec
	if err := fn.JsonConversion(obj.Spec, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *masterNodeInResolver) Spec(ctx context.Context, obj *entities.MasterNode, data *model.MasterNodeSpecIn) error {
	if obj == nil {
		return nil
	}
	return fn.JsonConversion(data, &obj.Spec)
}

// MasterNode returns generated.MasterNodeResolver implementation.
func (r *Resolver) MasterNode() generated.MasterNodeResolver { return &masterNodeResolver{r} }

// MasterNodeIn returns generated.MasterNodeInResolver implementation.
func (r *Resolver) MasterNodeIn() generated.MasterNodeInResolver { return &masterNodeInResolver{r} }

type masterNodeResolver struct{ *Resolver }
type masterNodeInResolver struct{ *Resolver }
