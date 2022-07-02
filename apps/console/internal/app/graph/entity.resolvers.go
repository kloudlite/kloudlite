package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/app/graph/model"
	"kloudlite.io/pkg/repos"
)

func (r *entityResolver) FindAccountByID(ctx context.Context, id repos.ID) (*model.Account, error) {
	return &model.Account{ID: id}, nil
}

func (r *entityResolver) FindClusterByID(ctx context.Context, id repos.ID) (*model.Cluster, error) {
	return r.Query().InfraGetCluster(ctx, id)
}

func (r *entityResolver) FindComputePlanByName(ctx context.Context, name string) (*model.ComputePlan, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *entityResolver) FindDeviceByID(ctx context.Context, id repos.ID) (*model.Device, error) {
	device, err := r.Domain.GetDevice(ctx, id)
	return &model.Device{
		ID:      device.Id,
		User:    &model.User{ID: device.UserId},
		Name:    device.Name,
		Cluster: &model.Cluster{ID: device.ClusterId},
		IP:      device.Ip,
	}, err
}

func (r *entityResolver) FindUserByID(ctx context.Context, id repos.ID) (*model.User, error) {
	return &model.User{ID: id}, nil
}

// Entity returns generated.EntityResolver implementation.
func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
