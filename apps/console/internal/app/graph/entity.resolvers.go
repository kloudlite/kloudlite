package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
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
	}, err
}

func (r *entityResolver) FindLambdaPlanByName(ctx context.Context, name string) (*model.LambdaPlan, error) {
	return &model.LambdaPlan{
		Name: name,
	}, nil
}

func (r *entityResolver) FindStoragePlanByName(ctx context.Context, name string) (*model.StoragePlan, error) {
	plans, err := r.Domain.GetStoragePlans(ctx)
	if err != nil {
		return nil, err
	}

	for _, plan := range plans {
		if plan.Name == name {
			return &model.StoragePlan{
				Name: plan.Name,
			}, nil
		}
	}
	return nil, errors.New("storage plan not found")
}

func (r *entityResolver) FindUserByID(ctx context.Context, id repos.ID) (*model.User, error) {
	return &model.User{ID: id}, nil
}

// Entity returns generated.EntityResolver implementation.
func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
