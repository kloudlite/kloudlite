package domain

import (
	"context"
	entities "kloudlite.io/apps/infra/internal/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

type InfraContext struct {
	context.Context
	UserId      repos.ID
	AccountName string
}

type Domain interface {
	CheckNameAvailability(ctx InfraContext, typeArg ResType, clusterName *string, name string) (*CheckNameAvailabilityOutput, error)

	CreateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error)
	ListClusters(ctx InfraContext, pagination t.CursorPagination) (*repos.PaginatedRecord[*entities.Cluster], error)
	GetCluster(ctx InfraContext, name string) (*entities.Cluster, error)

	UpdateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error)
	DeleteCluster(ctx InfraContext, name string) error

	OnDeleteClusterMessage(ctx InfraContext, cluster entities.Cluster) error
	OnUpdateClusterMessage(ctx InfraContext, cluster entities.Cluster) error

	CreateBYOCCluster(ctx InfraContext, byocCluster entities.BYOCCluster) (*entities.BYOCCluster, error)
	UpdateBYOCCluster(ctx InfraContext, byocCluster entities.BYOCCluster) (*entities.BYOCCluster, error)
	DeleteBYOCCluster(ctx InfraContext, name string) error

	ListBYOCClusters(ctx InfraContext, pagination t.CursorPagination) (*repos.PaginatedRecord[*entities.BYOCCluster], error)
	GetBYOCCluster(ctx InfraContext, name string) (*entities.BYOCCluster, error)

	OnDeleteBYOCClusterMessage(ctx InfraContext, byocCluster entities.BYOCCluster) error
	OnUpdateBYOCClusterMessage(ctx InfraContext, byocCluster entities.BYOCCluster) error

	CreateProviderSecret(ctx InfraContext, secret entities.CloudProviderSecret) (*entities.CloudProviderSecret, error)
	UpdateProviderSecret(ctx InfraContext, secret entities.CloudProviderSecret) (*entities.CloudProviderSecret, error)
	DeleteProviderSecret(ctx InfraContext, secretName string) error

	ListProviderSecrets(ctx InfraContext, pagination t.CursorPagination) (*repos.PaginatedRecord[*entities.CloudProviderSecret], error)
	GetProviderSecret(ctx InfraContext, name string) (*entities.CloudProviderSecret, error)

	CreateNodePool(ctx InfraContext, clusterName string, nodePool entities.NodePool) (*entities.NodePool, error)
	UpdateNodePool(ctx InfraContext, clusterName string, nodePool entities.NodePool) (*entities.NodePool, error)

	DeleteNodePool(ctx InfraContext, clusterName string, poolName string) error
	ListNodePools(ctx InfraContext, clusterName string, pagination t.CursorPagination) (*repos.PaginatedRecord[*entities.NodePool], error)
	GetNodePool(ctx InfraContext, clusterName string, poolName string) (*entities.NodePool, error)

	OnDeleteNodePoolMessage(ctx InfraContext, clusterName string, nodePool entities.NodePool) error
	OnUpdateNodePoolMessage(ctx InfraContext, clusterName string, nodePool entities.NodePool) error
}
