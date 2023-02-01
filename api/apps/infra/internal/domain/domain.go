package domain

import (
	"context"

	"go.uber.org/fx"
	"kloudlite.io/apps/infra/internal/domain/entities"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/finance"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

type domain struct {
	clusterRepo   repos.DbRepo[*entities.Cluster]
	edgeRepo      repos.DbRepo[*entities.Edge]
	providerRepo  repos.DbRepo[*entities.CloudProvider]
	financeClient finance.FinanceClient
}

func (d *domain) CreateCluster(ctx context.Context, cluster entities.Cluster) (*entities.Cluster, error) {
	return d.clusterRepo.Create(ctx, &cluster)
}

func (d *domain) ListClusters(ctx context.Context, accountId repos.ID) ([]*entities.Cluster, error) {
	return d.clusterRepo.Find(ctx, repos.Query{Filter: repos.Filter{"accountId": accountId}})
}

func (d *domain) GetCluster(ctx context.Context, name string) (*entities.Cluster, error) {
	return d.clusterRepo.FindOne(ctx, repos.Filter{"name": name})
}

func (d *domain) UpdateCluster(ctx context.Context, cluster entities.Cluster) (*entities.Cluster, error) {
	exCluster, err := d.clusterRepo.FindOne(ctx, repos.Filter{"id": cluster.Id})
	if err != nil {
		return nil, err
	}
	if exCluster == nil {
		return nil, errors.Newf("cluster with (id=%s) not found", cluster.Id)
	}
	return d.clusterRepo.UpdateById(ctx, cluster.Id, &cluster)
}

func (d *domain) DeleteCluster(ctx context.Context, name string) error {
	return d.clusterRepo.DeleteOne(ctx, repos.Filter{"name": name})
}

func (d *domain) CreateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider) (*entities.CloudProvider, error) {
	return d.providerRepo.Create(ctx, &cloudProvider)
}

func (d *domain) ListCloudProviders(ctx context.Context, accountId repos.ID) ([]*entities.CloudProvider, error) {
	return d.providerRepo.Find(ctx, repos.Query{Filter: repos.Filter{"spec.accountId": accountId}})
}

func (d *domain) GetCloudProvider(ctx context.Context, name string) (*entities.CloudProvider, error) {
	return d.providerRepo.FindOne(ctx, repos.Filter{"metadata.name": name})
}

func (d *domain) UpdateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider) (*entities.CloudProvider, error) {
	return d.providerRepo.UpdateOne(ctx, repos.Filter{"id": cloudProvider.Id}, &cloudProvider)
}

func (d *domain) DeleteCloudProvider(ctx context.Context, name string) error {
	return d.providerRepo.DeleteOne(ctx, repos.Filter{"metadata.name": name})
}

func (d *domain) CreateEdge(ctx context.Context, edge entities.Edge) (*entities.Edge, error) {
	return d.edgeRepo.Create(ctx, &edge)
}

func (d *domain) ListEdges(ctx context.Context, providerName string) ([]*entities.Edge, error) {
	return d.edgeRepo.Find(ctx, repos.Query{Filter: repos.Filter{"spec.provider": providerName}})
}

func (d *domain) GetEdge(ctx context.Context, name string) (*entities.Edge, error) {
	return d.edgeRepo.FindOne(ctx, repos.Filter{"metadata.name": name})
}

func (d *domain) UpdateEdge(ctx context.Context, edge entities.Edge) (*entities.Edge, error) {
	return d.edgeRepo.UpdateOne(ctx, repos.Filter{"id": edge.Id}, &edge)
}

func (d *domain) DeleteEdge(ctx context.Context, name string) error {
	return d.edgeRepo.DeleteOne(ctx, repos.Filter{"metadata.name": name})
}

var Module = fx.Module("domain",
	fx.Provide(
		func(
			clusterRepo repos.DbRepo[*entities.Cluster],
			providerRepo repos.DbRepo[*entities.CloudProvider],
			edgeRepo repos.DbRepo[*entities.Edge],
			financeClient finance.FinanceClient,
		) Domain {
			return &domain{
				clusterRepo:   clusterRepo,
				providerRepo:  providerRepo,
				edgeRepo:      edgeRepo,
				financeClient: financeClient,
			}
		}),
)
