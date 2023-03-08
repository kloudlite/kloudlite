package domain

import (
	"context"
	"kloudlite.io/apps/infra/internal/domain/entities"
)

type Domain interface {
	CreateCluster(ctx context.Context, cluster entities.Cluster) (*entities.Cluster, error)
	ListClusters(ctx context.Context, accountName string) ([]*entities.Cluster, error)
	GetCluster(ctx context.Context, name string) (*entities.Cluster, error)
	UpdateCluster(ctx context.Context, cluster entities.Cluster) (*entities.Cluster, error)
	DeleteCluster(ctx context.Context, name string) error
	OnDeleteClusterMessage(ctx context.Context, cluster entities.Cluster) error
	OnUpdateClusterMessage(ctx context.Context, cluster entities.Cluster) error

	CreateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider, providerSecret entities.Secret) (*entities.CloudProvider, error)
	ListCloudProviders(ctx context.Context, accountName string) ([]*entities.CloudProvider, error)
	GetCloudProvider(ctx context.Context, accountName string, name string) (*entities.CloudProvider, error)
	UpdateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider, providerSecret *entities.Secret) (*entities.CloudProvider, error)
	DeleteCloudProvider(ctx context.Context, accountName string, name string) error
	OnDeleteCloudProviderMessage(ctx context.Context, cloudProvider entities.CloudProvider) error
	OnUpdateCloudProviderMessage(ctx context.Context, cloudProvider entities.CloudProvider) error

	CreateEdge(ctx context.Context, edge entities.Edge) (*entities.Edge, error)
	ListEdges(ctx context.Context, clusterName string, providerName *string) ([]*entities.Edge, error)
	GetEdge(ctx context.Context, clusterName string, name string) (*entities.Edge, error)
	UpdateEdge(ctx context.Context, edge entities.Edge) (*entities.Edge, error)
	DeleteEdge(ctx context.Context, clusterName string, name string) error
	OnDeleteEdgeMessage(ctx context.Context, edge entities.Edge) error
	OnUpdateEdgeMessage(ctx context.Context, edge entities.Edge) error

	GetNodePools(ctx context.Context, clusterName string, edgeName string) ([]*entities.NodePool, error)
	GetMasterNodes(ctx context.Context, clusterName string) ([]*entities.MasterNode, error)
	GetWorkerNodes(ctx context.Context, clusterName string, edgeName string) ([]*entities.WorkerNode, error)
	DeleteWorkerNode(ctx context.Context, clusterName string, edgeName string, name string) (bool, error)
	OnDeleteWorkerNodeMessage(ctx context.Context, workerNode entities.WorkerNode) error
	OnUpdateWorkerNodeMessage(ctx context.Context, workerNode entities.WorkerNode) error
}
