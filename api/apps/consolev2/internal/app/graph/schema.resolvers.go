package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"kloudlite.io/apps/consolev2/internal/app/graph/generated"
	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (r *mutationResolver) CoreCreateCloudProvider(ctx context.Context, in entities.CloudProvider) (*entities.CloudProvider, error) {
	return r.Domain.CreateCloudProvider(ctx, &in)
}

func (r *mutationResolver) CoreUpdateCloudProvider(ctx context.Context, in entities.CloudProvider) (bool, error) {
	if err := r.Domain.UpdateCloudProvider(ctx, &in); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreDeleteCloudProvider(ctx context.Context, name string) (bool, error) {
	if err := r.Domain.DeleteCloudProvider(ctx, name); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreSample(ctx context.Context, j map[string]interface{}) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CoreListCloudProviders(ctx context.Context, accountID string) ([]*entities.CloudProvider, error) {
	return r.Domain.ListCloudProviders(ctx, repos.ID(accountID))
}

func (r *queryResolver) CoreGetCloudProvider(ctx context.Context, name string) (*entities.CloudProvider, error) {
	return r.Domain.GetCloudProvider(ctx, name)
}

func (r *queryResolver) CoreSample(ctx context.Context) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
