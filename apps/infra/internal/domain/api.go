package domain

import (
	"context"
	"kloudlite.io/apps/infra/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

type Domain interface {
	CreateCluster(ctx context.Context, cluster entities.Cluster) (*entities.Cluster, error)
	ListClusters(ctx context.Context, accountName string) ([]*entities.Cluster, error)
	GetCluster(ctx context.Context, name string) (*entities.Cluster, error)
	UpdateCluster(ctx context.Context, cluster entities.Cluster) (*entities.Cluster, error)
	DeleteCluster(ctx context.Context, name string) error

	CreateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider, providerSecret entities.Secret) (*entities.CloudProvider, error)
	ListCloudProviders(ctx context.Context, accountId repos.ID) ([]*entities.CloudProvider, error)
	GetCloudProvider(ctx context.Context, accountId repos.ID, name string) (*entities.CloudProvider, error)
	UpdateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider, providerSecret *entities.Secret) (*entities.CloudProvider, error)
	DeleteCloudProvider(ctx context.Context, accountId repos.ID, name string) error

	CreateEdge(ctx context.Context, edge entities.Edge) (*entities.Edge, error)
	ListEdges(ctx context.Context, clusterName string, providerName string) ([]*entities.Edge, error)
	GetEdge(ctx context.Context, clusterName string, name string) (*entities.Edge, error)
	UpdateEdge(ctx context.Context, edge entities.Edge) (*entities.Edge, error)
	DeleteEdge(ctx context.Context, clusterName string, name string) error
	GetNodePools(ctx context.Context, clusterName string) ([]*entities.NodePool, error)
	GetMasterNodes(ctx context.Context, clusterName string) ([]*entities.MasterNode, error)
	GetWorkerNodes(ctx context.Context, clusterName string, edgeName string) ([]*entities.WorkerNode, error)
	DeleteWorkerNode(ctx context.Context, clusterName string, edgeName string, name string) (bool, error)
}
