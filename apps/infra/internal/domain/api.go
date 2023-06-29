package domain

import (
	"context"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"

	"kloudlite.io/apps/infra/internal/domain/entities"
)

type InfraContext struct {
	context.Context
	AccountName string
}

type Domain interface {
	CheckNameAvailability(ctx InfraContext, typeArg ResType, name string) (*CheckNameAvailabilityOutput, error)

	ListBYOCClusters(ctx InfraContext, pagination t.CursorPagination) (*repos.PaginatedRecord[*entities.BYOCCluster], error)
	GetBYOCCluster(ctx InfraContext, name string) (*entities.BYOCCluster, error)

	CreateBYOCCluster(ctx InfraContext, cluster entities.BYOCCluster) (*entities.BYOCCluster, error)
	UpdateBYOCCluster(ctx InfraContext, cluster entities.BYOCCluster) (*entities.BYOCCluster, error)
	DeleteBYOCCluster(ctx InfraContext, name string) error

	OnDeleteBYOCClusterMessage(ctx InfraContext, cluster entities.BYOCCluster) error
	OnBYOCClusterHelmUpdates(ctx InfraContext, cluster entities.BYOCCluster) error

	CreateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error)
	ListClusters(ctx InfraContext, pagination t.CursorPagination) (*repos.PaginatedRecord[*entities.Cluster], error)
	GetCluster(ctx InfraContext, name string) (*entities.Cluster, error)

	UpdateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error)
	DeleteCluster(ctx InfraContext, name string) error

	OnDeleteClusterMessage(ctx InfraContext, cluster entities.Cluster) error
	OnUpdateClusterMessage(ctx InfraContext, cluster entities.Cluster) error

	GetProviderSecret(ctx InfraContext, name string) (*entities.Secret, error)

	CreateCloudProvider(ctx InfraContext, cloudProvider entities.CloudProvider, providerSecret entities.Secret) (*entities.CloudProvider, error)
	ListCloudProviders(ctx InfraContext, pagination t.CursorPagination) (*repos.PaginatedRecord[*entities.CloudProvider], error)
	GetCloudProvider(ctx InfraContext, name string) (*entities.CloudProvider, error)
	UpdateCloudProvider(ctx InfraContext, cloudProvider entities.CloudProvider, providerSecret *entities.Secret) (*entities.CloudProvider, error)
	DeleteCloudProvider(ctx InfraContext, name string) error
	OnDeleteCloudProviderMessage(ctx InfraContext, cloudProvider entities.CloudProvider) error
	OnUpdateCloudProviderMessage(ctx InfraContext, cloudProvider entities.CloudProvider) error

	ListEdges(ctx InfraContext, clusterName string, providerName *string, pagination t.CursorPagination) (*repos.PaginatedRecord[*entities.Edge], error)
	GetEdge(ctx InfraContext, clusterName string, name string) (*entities.Edge, error)

	CreateEdge(ctx InfraContext, edge entities.Edge) (*entities.Edge, error)
	UpdateEdge(ctx InfraContext, edge entities.Edge) (*entities.Edge, error)
	DeleteEdge(ctx InfraContext, clusterName string, name string) error

	OnDeleteEdgeMessage(ctx InfraContext, edge entities.Edge) error
	OnUpdateEdgeMessage(ctx InfraContext, edge entities.Edge) error

	ListNodePools(ctx InfraContext, clusterName string, edgeName string, pagination t.CursorPagination) (*repos.PaginatedRecord[*entities.NodePool], error)
	GetNodePool(ctx InfraContext, clusterName string, edgeName string, poolName string) (*entities.NodePool, error)

	ListMasterNodes(ctx InfraContext, clusterName string) ([]*entities.MasterNode, error)

	ListWorkerNodes(ctx InfraContext, clusterName string, edgeName string) ([]*entities.WorkerNode, error)
	DeleteWorkerNode(ctx InfraContext, clusterName string, edgeName string, name string) (bool, error)

	OnDeleteWorkerNodeMessage(ctx InfraContext, workerNode entities.WorkerNode) error
	OnUpdateWorkerNodeMessage(ctx InfraContext, workerNode entities.WorkerNode) error
}
