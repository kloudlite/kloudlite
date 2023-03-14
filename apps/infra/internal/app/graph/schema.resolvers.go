package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/kloudlite/wg-operator/apis/wg/v1"
	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/domain/entities"
)

func (r *mutationResolver) InfraCreateAccount(ctx context.Context, account v1.Account) (*v1.Account, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) InfraUpdateAccount(ctx context.Context, account v1.Account) (*v1.Account, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) InfraDeleteAccount(ctx context.Context, name *string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

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

func (r *mutationResolver) InfraCreateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider, providerSecret entities.Secret) (*entities.CloudProvider, error) {
	return r.Domain.CreateCloudProvider(ctx, cloudProvider, providerSecret)
}

func (r *mutationResolver) InfraUpdateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider, providerSecret *entities.Secret) (*entities.CloudProvider, error) {
	return r.Domain.UpdateCloudProvider(ctx, cloudProvider, providerSecret)
}

func (r *mutationResolver) InfraDeleteCloudProvider(ctx context.Context, accountName string, name string) (bool, error) {
	if err := r.Domain.DeleteCloudProvider(ctx, accountName, name); err != nil {
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

func (r *mutationResolver) InfraDeleteEdge(ctx context.Context, clusterName string, name string) (bool, error) {
	if err := r.Domain.DeleteEdge(ctx, clusterName, name); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) InfraDeleteWorkerNode(ctx context.Context, clusterName string, edgeName string, name string) (bool, error) {
	return r.Domain.DeleteWorkerNode(ctx, clusterName, edgeName, name)
}

func (r *queryResolver) InfraListAccounts(ctx context.Context) ([]*v1.Account, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) InfraGetAccount(ctx context.Context, name string) (*v1.Account, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) InfraListClusters(ctx context.Context, accountName string) ([]*entities.Cluster, error) {
	cls, err := r.Domain.ListClusters(ctx, accountName)
	if cls == nil {
		cls = make([]*entities.Cluster, 0)
	}
	return cls, err
}

func (r *queryResolver) InfraGetCluster(ctx context.Context, name string) (*entities.Cluster, error) {
	return r.Domain.GetCluster(ctx, name)
}

func (r *queryResolver) InfraListCloudProviders(ctx context.Context, accountName string) ([]*entities.CloudProvider, error) {
	cp, err := r.Domain.ListCloudProviders(ctx, accountName)
	if cp == nil {
		cp = make([]*entities.CloudProvider, 0)
	}
	return cp, err
}

func (r *queryResolver) InfraGetCloudProvider(ctx context.Context, accountName string, name string) (*entities.CloudProvider, error) {
	return r.Domain.GetCloudProvider(ctx, accountName, name)
}

func (r *queryResolver) InfraListEdges(ctx context.Context, clusterName string, providerName *string) ([]*entities.Edge, error) {
	e, err := r.Domain.ListEdges(ctx, clusterName, providerName)
	if e == nil {
		e = make([]*entities.Edge, 0)
	}
	return e, err
}

func (r *queryResolver) InfraGetEdge(ctx context.Context, clusterName string, name string) (*entities.Edge, error) {
	return r.Domain.GetEdge(ctx, clusterName, name)
}

func (r *queryResolver) InfraGetMasterNodes(ctx context.Context, clusterName string) ([]*entities.MasterNode, error) {
	return r.Domain.GetMasterNodes(ctx, clusterName)
}

func (r *queryResolver) InfraGetWorkerNodes(ctx context.Context, clusterName string, edgeName string) ([]*entities.WorkerNode, error) {
	return r.Domain.GetWorkerNodes(ctx, clusterName, edgeName)
}

func (r *queryResolver) InfraGetNodePools(ctx context.Context, clusterName string, edgeName string) ([]*entities.NodePool, error) {
	return r.Domain.GetNodePools(ctx, clusterName, edgeName)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
