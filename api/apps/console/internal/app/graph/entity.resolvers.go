package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/app/graph/model"
	"kloudlite.io/pkg/repos"
)

func (r *entityResolver) FindClusterByID(ctx context.Context, id repos.ID) (*model.Cluster, error) {
	return r.Query().GetCluster(ctx, id)
}

func (r *entityResolver) FindDeviceByID(ctx context.Context, id repos.ID) (*model.Device, error) {
	return r.Query().GetDevice(ctx, id)
}

func (r *entityResolver) FindUserByID(ctx context.Context, id repos.ID) (*model.User, error) {
	return &model.User{ID: id}, nil
}

// Entity returns generated.EntityResolver implementation.
func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
