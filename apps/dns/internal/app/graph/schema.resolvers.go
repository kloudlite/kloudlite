package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"kloudlite.io/apps/dns/internal/app/graph/generated"
	"kloudlite.io/apps/dns/internal/app/graph/model"
	"kloudlite.io/pkg/repos"
)

func (r *accountResolver) Sites(ctx context.Context, obj *model.Account) ([]*model.Site, error) {
	sitesEntities, err := r.d.GetSites(ctx, string(obj.ID))
	if err != nil {
		return nil, err
	}
	sites := make([]*model.Site, 0)
	for _, e := range sitesEntities {
		edgeCname, err := r.d.GetAccountEdgeCName(ctx, string(e.AccountId), e.RegionId)
		if err != nil {
			edgeCname = ""
		}
		sites = append(sites, &model.Site{
			ID:         e.Id,
			RegionID:   e.RegionId,
			AccountID:  e.AccountId,
			IsVerified: e.Verified,
			Domain:     e.Domain,
			EdgeCname:  edgeCname,
		})
	}
	return sites, nil
}

func (r *mutationResolver) DNSCreateSite(ctx context.Context, domain string, accountID repos.ID, regionID repos.ID) (bool, error) {
	err := r.d.CreateSite(ctx, domain, accountID, regionID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) DNSDeleteSite(ctx context.Context, siteID repos.ID) (bool, error) {
	err := r.d.DeleteSite(ctx, siteID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) DNSVerifySite(ctx context.Context, siteID repos.ID) (bool, error) {
	err := r.d.VerifySite(ctx, siteID)
	return err == nil, err
}

func (r *queryResolver) DNSGetSite(ctx context.Context, siteID repos.ID) (*model.Site, error) {
	site, err := r.d.GetSite(ctx, string(siteID))
	if err != nil {
		return nil, err
	}
	edgeCname, err := r.d.GetAccountEdgeCName(ctx, string(site.AccountId), site.RegionId)
	if err != nil {
		return nil, err
	}

	return &model.Site{
		ID:         site.Id,
		RegionID:   site.RegionId,
		AccountID:  site.AccountId,
		IsVerified: site.Verified,
		Domain:     site.Domain,
		EdgeCname:  edgeCname,
	}, nil
}

// Account returns generated.AccountResolver implementation.
func (r *Resolver) Account() generated.AccountResolver { return &accountResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type accountResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
