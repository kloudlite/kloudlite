package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"kloudlite.io/apps/dns/internal/app/graph/generated"
	"kloudlite.io/apps/dns/internal/app/graph/model"
	"kloudlite.io/pkg/repos"
)

func (r *entityResolver) FindAccountByID(ctx context.Context, id repos.ID) (*model.Account, error) {
	return &model.Account{
		ID: id,
	}, nil
}

func (r *entityResolver) FindSiteByID(ctx context.Context, id repos.ID) (*model.Site, error) {
	site, err := r.d.GetSite(ctx, string(id))
	if err != nil {
		return nil, err
	}
	return &model.Site{
		ID:        id,
		AccountID: site.AccountId,
		Domain:    site.Domain,
	}, nil
}

// Entity returns generated.EntityResolver implementation.
func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
