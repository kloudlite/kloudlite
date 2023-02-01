package domain

import (
	"context"
	"kloudlite.io/apps/infra/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	CreateCluster(ctx context.Context, cluster entities.Cluster) (*entities.Cluster, error)
	ListClusters(ctx context.Context, accountId repos.ID) ([]*entities.Cluster, error)
	GetCluster(ctx context.Context, name string) (*entities.Cluster, error)
	UpdateCluster(ctx context.Context, cluster entities.Cluster) (*entities.Cluster, error)
	DeleteCluster(ctx context.Context, name string) error

	CreateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider) (*entities.CloudProvider, error)
	ListCloudProviders(ctx context.Context, accountId repos.ID) ([]*entities.CloudProvider, error)
	GetCloudProvider(ctx context.Context, name string) (*entities.CloudProvider, error)
	UpdateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider) (*entities.CloudProvider, error)
	DeleteCloudProvider(ctx context.Context, name string) error

	CreateEdge(ctx context.Context, edge entities.Edge) (*entities.Edge, error)
	ListEdges(ctx context.Context, providerName string) ([]*entities.Edge, error)
	GetEdge(ctx context.Context, name string) (*entities.Edge, error)
	UpdateEdge(ctx context.Context, edge entities.Edge) (*entities.Edge, error)
	DeleteEdge(ctx context.Context, edge string) error
}
