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

func (r *mutationResolver) DNSCreateSite(ctx context.Context, domain *string, accountID *string) (*model.Site, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DNSCreateRecord(ctx context.Context, siteID *repos.ID, recordType *string, host *string, ttl *int, priority *int) (*model.Record, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DNSVerifySite(ctx context.Context, vid *repos.ID) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) DNSGetSites(ctx context.Context, accountID *string) ([]*model.Site, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) DNSGetSite(ctx context.Context, siteID *string) (*model.Site, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) DNSGetRecords(ctx context.Context, siteID *string) ([]*model.Record, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
