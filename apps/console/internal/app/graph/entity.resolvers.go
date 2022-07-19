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

func (r *entityResolver) FindAppByID(ctx context.Context, id repos.ID) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *entityResolver) FindComputePlanByName(ctx context.Context, name string) (*model.ComputePlan, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *entityResolver) FindDeviceByID(ctx context.Context, id repos.ID) (*model.Device, error) {
	device, err := r.Domain.GetDevice(ctx, id)
	return &model.Device{
		ID:   device.Id,
		User: &model.User{ID: device.UserId},
		Name: device.Name,
		IP:   device.Ip,
	}, err
}

func (r *entityResolver) FindLambdaPlanByName(ctx context.Context, name string) (*model.LambdaPlan, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *entityResolver) FindStoragePlanByName(ctx context.Context, name string) (*model.StoragePlan, error) {
	panic(fmt.Errorf("not implemented"))
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
func (r *entityResolver) FindLamdaPlanByName(ctx context.Context, name string) (*model.LamdaPlan, error) {
	panic(fmt.Errorf("not implemented"))
}
