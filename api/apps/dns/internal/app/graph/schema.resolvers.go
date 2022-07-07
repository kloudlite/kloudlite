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

func (r *accountResolver) DomainVerifications(ctx context.Context, obj *model.Account) ([]*model.Verification, error) {
	verifications, err := r.d.GetVerifications(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	vs := make([]*model.Verification, 0)
	for _, v := range verifications {
		vs = append(vs, &model.Verification{
			ID:         v.Id,
			VerifyText: v.VerifyText,
			Site: &model.Site{
				ID: v.SiteId,
			},
		})
	}
	return vs, nil
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
			Verified:  e.Verified,
		})
	}
	return sites, nil
}

func (r *mutationResolver) DNSCreateSite(ctx context.Context, domain string, accountID repos.ID) (*model.Verification, error) {
	vE, err := r.d.CreateSite(ctx, domain, repos.ID(accountID))
	if err != nil {
		return nil, err
	}
	return &model.Verification{
		ID:         vE.Id,
		VerifyText: vE.VerifyText,
	}, nil
}

func (r *mutationResolver) DNSDeleteSite(ctx context.Context, siteID repos.ID) (*model.Verification, error) {
	panic(fmt.Errorf("not implemented"))
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

func (r *queryResolver) DNSGetRecords(ctx context.Context, siteID repos.ID) ([]*model.Record, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *siteResolver) Verification(ctx context.Context, obj *model.Site, accountID repos.ID) (*model.Verification, error) {
	verification, err := r.d.GetVerification(ctx, obj.AccountID, obj.ID)
	if err != nil {
		return nil, err
	}
	return &model.Verification{
		ID:         verification.Id,
		VerifyText: verification.VerifyText,
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

type accountResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type siteResolver struct{ *Resolver }
