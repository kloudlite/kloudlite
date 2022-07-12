package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"kloudlite.io/apps/finance/internal/app/graph/generated"
	"kloudlite.io/apps/finance/internal/app/graph/model"
	"kloudlite.io/pkg/repos"
)

func (r *entityResolver) FindAccountByID(ctx context.Context, id repos.ID) (*model.Account, error) {
	accountEntity, err := r.domain.GetAccount(ctx, id)
	if err != nil {
		return nil, err
	}
	return AccountModelFromEntity(accountEntity), nil
}

func (r *entityResolver) FindComputePlanByName(ctx context.Context, name string) (*model.ComputePlan, error) {
	byName, err := r.domain.GetComputePlanByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return &model.ComputePlan{
		Name:           byName.Name,
		SharedPrice:    byName.SharedPrice,
		DedicatedPrice: byName.DedicatedPrice,
	}, nil
}

func (r *entityResolver) FindLamdaPlanByName(ctx context.Context, name string) (*model.LamdaPlan, error) {
	byName, err := r.domain.GetLambdaPlanByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return &model.LamdaPlan{
		Name:         byName.Name,
		FreeTier:     byName.FreeTire,
		PricePerGBHr: byName.PricePerGBHr,
	}, nil
}

func (r *entityResolver) FindStoragePlanByName(ctx context.Context, name string) (*model.StoragePlan, error) {
	byName, err := r.domain.GetStoragePlanByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return &model.StoragePlan{
		Name:       byName.Name,
		PricePerGb: byName.PricePerGB,
	}, nil
}

func (r *entityResolver) FindUserByID(ctx context.Context, id repos.ID) (*model.User, error) {
	return &model.User{
		ID: id,
	}, nil
}

// Entity returns generated.EntityResolver implementation.
func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
