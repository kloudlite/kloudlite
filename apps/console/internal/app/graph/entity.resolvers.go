package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

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

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.
func (r *entityResolver) FindComputeInventoryItemByName(ctx context.Context, name string) (*model.ComputeInventoryItem, error) {
	plan, err := r.Domain.GetComputePlan(ctx, name)
	if err != nil {
		return nil, err
	}
	return &model.ComputeInventoryItem{
		Name: name,
		Data: &model.ComputeInventoryData{
			Memory: &model.ComputeInventoryMetricSize{
				Quantity: plan.Memory.Quantity,
				Unit:     plan.Memory.Unit,
			},
			CPU: &model.ComputeInventoryMetricSize{
				Quantity: plan.Cpu.Quantity,
				Unit:     plan.Cpu.Unit,
			},
		},
	}, nil
}
