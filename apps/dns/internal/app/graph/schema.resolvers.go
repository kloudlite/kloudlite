package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"kloudlite.io/apps/dns/internal/app/graph/generated"
	"kloudlite.io/apps/dns/internal/app/graph/model"
	"kloudlite.io/pkg/repos"
)

func (r *accountResolver) DomainClaims(ctx context.Context, obj *model.Account) ([]*model.SiteClaim, error) {
	claims, err := r.d.GetSiteClaims(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	scs := make([]*model.SiteClaim, 0)
	for _, e := range claims {
		scs = append(scs, &model.SiteClaim{
			ID: e.Id,
			Account: &model.Account{
				ID: e.AccountId,
			},
			Site: &model.Site{
				ID: e.SiteId,
			},
		})
	}
	return scs, nil
}

func (r *accountResolver) Sites(ctx context.Context, obj *model.Account) ([]*model.Site, error) {
	sitesEntities, err := r.d.GetSites(ctx, string(obj.ID))
	if err != nil {
		return nil, err
	}
	sites := make([]*model.Site, 0)
	for _, e := range sitesEntities {
		sites = append(sites, &model.Site{
			ID:        e.Id,
			AccountID: e.AccountId,
			Domain:    e.Domain,
		})
	}
	return sites, nil
}

func (r *accountResolver) NameServers(ctx context.Context, obj *model.Account) ([]string, error) {
	return r.d.GetNameServers(ctx, obj.ID)
}

func (r *mutationResolver) DNSCreateSite(ctx context.Context, domain string, accountID repos.ID) (bool, error) {
	err := r.d.CreateSite(ctx, domain, repos.ID(accountID))
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) DNSCreateRecord(ctx context.Context, siteID repos.ID, recordType string, host string, answer string, ttl int, priority *int) (*model.Record, error) {
	record, err := r.d.CreateRecord(ctx, siteID, recordType, host, answer, uint32(ttl), int64(*priority))
	if err != nil {
		return nil, err
	}
	p := int(record.Priority)
	return &model.Record{
		ID:         record.Id,
		SiteID:     record.SiteId,
		RecordType: record.Type,
		Host:       record.Host,
		Answer:     record.Answer,
		TTL:        int(record.TTL),
		Priority:   &p,
	}, nil
}

func (r *mutationResolver) DNSDeleteRecord(ctx context.Context, recordID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DNSUpdateRecord(ctx context.Context, recordID repos.ID, recordType string, host string, answer string, ttl int, priority *int) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DNSVerifySite(ctx context.Context, vid repos.ID) (bool, error) {
	err := r.d.VerifySite(ctx, vid)
	return err == nil, err
}

func (r *queryResolver) DNSGetSite(ctx context.Context, siteID repos.ID) (*model.Site, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *siteResolver) Records(ctx context.Context, obj *model.Site, siteID repos.ID) ([]*model.Record, error) {
	records, err := r.d.GetRecords(ctx, string(siteID))
	if err != nil {
		return nil, err
	}
	rs := make([]*model.Record, 0)
	for _, e := range records {
		p := int(e.Priority)
		rs = append(rs, &model.Record{
			ID:         e.Id,
			SiteID:     e.SiteId,
			RecordType: e.Type,
			Host:       e.Host,
			Answer:     e.Answer,
			TTL:        int(e.TTL),
			Priority:   &p,
		})
	}
	return rs, nil
}

func (r *siteClaimResolver) Site(ctx context.Context, obj *model.SiteClaim) (*model.Site, error) {
	site, err := r.d.GetSite(ctx, string(obj.Site.ID))
	if err != nil {
		return nil, err
	}
	return &model.Site{
		ID:        site.Id,
		AccountID: site.AccountId,
		Domain:    site.Domain,
	}, nil
}

// Account returns generated.AccountResolver implementation.
func (r *Resolver) Account() generated.AccountResolver { return &accountResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Site returns generated.SiteResolver implementation.
func (r *Resolver) Site() generated.SiteResolver { return &siteResolver{r} }

// SiteClaim returns generated.SiteClaimResolver implementation.
func (r *Resolver) SiteClaim() generated.SiteClaimResolver { return &siteClaimResolver{r} }

type accountResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type siteResolver struct{ *Resolver }
type siteClaimResolver struct{ *Resolver }
