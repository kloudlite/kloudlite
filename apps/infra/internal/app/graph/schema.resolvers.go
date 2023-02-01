package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (r *mutationResolver) InfraCreateCluster(ctx context.Context, cluster entities.Cluster) (*entities.Cluster, error) {
	return r.Domain.CreateCluster(ctx, cluster)
}

func (r *mutationResolver) InfraUpdateCluster(ctx context.Context, cluster entities.Cluster) (*entities.Cluster, error) {
	return r.Domain.UpdateCluster(ctx, cluster)
}

func (r *mutationResolver) InfraDeleteCluster(ctx context.Context, name string) (bool, error) {
	if err := r.Domain.DeleteCluster(ctx, name); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) InfraCreateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider, creds entities.ProviderSecrets) (*entities.CloudProvider, error) {
	return r.Domain.CreateCloudProvider(ctx, cloudProvider)
}

func (r *mutationResolver) InfraUpdateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider, creds entities.ProviderSecrets) (*entities.CloudProvider, error) {
	return r.Domain.UpdateCloudProvider(ctx, cloudProvider)
}

func (r *mutationResolver) InfraDeleteCloudProvider(ctx context.Context, name string) (bool, error) {
	if err := r.Domain.DeleteCloudProvider(ctx, name); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) InfraCreateEdge(ctx context.Context, edge entities.Edge) (*entities.Edge, error) {
	return r.Domain.CreateEdge(ctx, edge)
}

func (r *mutationResolver) InfraUpdateEdge(ctx context.Context, edge entities.Edge) (*entities.Edge, error) {
	return r.Domain.UpdateEdge(ctx, edge)
}

func (r *mutationResolver) InfraDeleteEdge(ctx context.Context, name string) (bool, error) {
	if err := r.Domain.DeleteEdge(ctx, name); err != nil {
		return false, err
	}
	return true, nil
}

func (r *queryResolver) InfraListClusters(ctx context.Context, accountID repos.ID) ([]*entities.Cluster, error) {
	return r.Domain.ListClusters(ctx, accountID)
}

func (r *queryResolver) InfraGetCluster(ctx context.Context, name string) (*entities.Cluster, error) {
	return r.Domain.GetCluster(ctx, name)
}

func (r *queryResolver) InfraListCloudProviders(ctx context.Context, accountID string) ([]*entities.CloudProvider, error) {
	return r.Domain.ListCloudProviders(ctx, repos.ID(accountID))
}

func (r *queryResolver) InfraGetCloudProvider(ctx context.Context, name string) (*entities.CloudProvider, error) {
	return r.Domain.GetCloudProvider(ctx, name)
}

func (r *queryResolver) InfraListEdges(ctx context.Context, providerName string) ([]*entities.Edge, error) {
	return r.Domain.ListEdges(ctx, providerName)
}

func (r *queryResolver) InfraGetEdge(ctx context.Context, name string) (*entities.Edge, error) {
	return r.Domain.GetEdge(ctx, name)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
