package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"time"

	"kloudlite.io/apps/finance/internal/app/graph/generated"
	"kloudlite.io/apps/finance/internal/app/graph/model"
	"kloudlite.io/pkg/repos"
	fn "kloudlite.io/pkg/functions"
)

func (r *entityResolver) FindAccountByName(ctx context.Context, name string) (*model.Account, error) {
	acc, err := r.domain.GetAccount(toFinanceContext(ctx), name)
	if err != nil {
		return nil, err
	}
	return &model.Account{
		// ID:           acc.Id,
		Name:         acc.Name,
		Billing:      nil,
		IsActive:     fn.DefaultIfNil(acc.IsActive, false),
		ContactEmail: acc.ContactEmail,
		ReadableID:   acc.ReadableId,
		Created:      acc.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (r *entityResolver) FindComputePlanByName(ctx context.Context, name string) (*model.ComputePlan, error) {
	// byName, err := r.domain.GetComputePlanByName(ctx, name)
	// if err != nil {
	// 	return nil, err
	// }
	// return &model.ComputePlan{
	// 	Name:           byName.Name,
	// 	SharedPrice:    byName.SharedPrice,
	// 	DedicatedPrice: byName.DedicatedPrice,
	// }, nil
	return nil, nil
}

func (r *entityResolver) FindLambdaPlanByName(ctx context.Context, name string) (*model.LambdaPlan, error) {
	// byName, err := r.domain.GetLambdaPlanByName(ctx, name)
	// if err != nil {
	// 	return nil, err
	// }
	// return &model.LambdaPlan{
	// 	Name:         byName.Name,
	// 	FreeTier:     byName.FreeTire,
	// 	PricePerGBHr: byName.PricePerGBHr,
	// }, nil
	return nil, nil
}

func (r *entityResolver) FindStoragePlanByName(ctx context.Context, name string) (*model.StoragePlan, error) {
	// byName, err := r.domain.GetStoragePlanByName(ctx, name)
	// if err != nil {
	// 	return nil, err
	// }
	// return &model.StoragePlan{
	// 	Name:       byName.Name,
	// 	PricePerGb: byName.PricePerGB,
	// }, nil
	return nil, nil
}

func (r *entityResolver) FindUserByID(ctx context.Context, id repos.ID) (*model.User, error) {
	return &model.User{ID: id}, nil
}

// Entity returns generated.EntityResolver implementation.
func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
