package domain

import (
	"context"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"time"
)

type InfraContext struct {
	context.Context
	UserId      repos.ID
	UserEmail   string
	UserName    string
	AccountName string
}

type UpdateAndDeleteOpts struct {
	MessageTimestamp time.Time
}

type Domain interface {
	CheckNameAvailability(ctx InfraContext, typeArg ResType, clusterName *string, name string) (*CheckNameAvailabilityOutput, error)

	CreateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error)
	UpdateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error)
	DeleteCluster(ctx InfraContext, name string) error

	ListClusters(ctx InfraContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Cluster], error)
	GetCluster(ctx InfraContext, name string) (*entities.Cluster, error)

	GetClusterAdminKubeconfig(ctx InfraContext, clusterName string) (*string, error)

	OnDeleteClusterMessage(ctx InfraContext, cluster entities.Cluster) error
	OnUpdateClusterMessage(ctx InfraContext, cluster entities.Cluster, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	CreateProviderSecret(ctx InfraContext, secret entities.CloudProviderSecret) (*entities.CloudProviderSecret, error)
	UpdateProviderSecret(ctx InfraContext, secret entities.CloudProviderSecret) (*entities.CloudProviderSecret, error)
	DeleteProviderSecret(ctx InfraContext, secretName string) error

	ListProviderSecrets(ctx InfraContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.CloudProviderSecret], error)
	GetProviderSecret(ctx InfraContext, name string) (*entities.CloudProviderSecret, error)

	ValidateProviderSecretAWSAccess(ctx InfraContext, name string) (*AWSAccessValidationOutput, error)

	ListDomainEntries(ctx InfraContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.DomainEntry], error)
	GetDomainEntry(ctx InfraContext, name string) (*entities.DomainEntry, error)

	CreateDomainEntry(ctx InfraContext, domainName entities.DomainEntry) (*entities.DomainEntry, error)
	UpdateDomainEntry(ctx InfraContext, domainName entities.DomainEntry) (*entities.DomainEntry, error)
	DeleteDomainEntry(ctx InfraContext, name string) error

	CreateNodePool(ctx InfraContext, clusterName string, nodePool entities.NodePool) (*entities.NodePool, error)
	UpdateNodePool(ctx InfraContext, clusterName string, nodePool entities.NodePool) (*entities.NodePool, error)
	DeleteNodePool(ctx InfraContext, clusterName string, poolName string) error

	ListNodePools(ctx InfraContext, clusterName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.NodePool], error)
	GetNodePool(ctx InfraContext, clusterName string, poolName string) (*entities.NodePool, error)

	OnDeleteNodePoolMessage(ctx InfraContext, clusterName string, nodePool entities.NodePool) error
	OnUpdateNodePoolMessage(ctx InfraContext, clusterName string, nodePool entities.NodePool, status types.ResourceStatus, opts UpdateAndDeleteOpts) error
	OnNodepoolApplyError(ctx InfraContext, clusterName string, name string, errMsg string, opts UpdateAndDeleteOpts) error

	ListNodes(ctx InfraContext, clusterName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Node], error)
	GetNode(ctx InfraContext, clusterName string, nodeName string) (*entities.Node, error)

	OnNodeUpdateMessage(ctx InfraContext, clusterName string, node entities.Node) error
	OnNodeDeleteMessage(ctx InfraContext, clusterName string, node entities.Node) error

	ListVPNDevices(ctx context.Context, accountName string, clusterName *string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.VPNDevice], error)
	GetVPNDevice(ctx InfraContext, clusterName string, deviceName string) (*entities.VPNDevice, error)
	GetWgConfigForDevice(ctx InfraContext, clusterName string, deviceName string) (*string, error)

	CreateVPNDevice(ctx InfraContext, clusterName string, device entities.VPNDevice) (*entities.VPNDevice, error)
	UpdateVPNDevice(ctx InfraContext, clusterName string, device entities.VPNDevice) (*entities.VPNDevice, error)
	DeleteVPNDevice(ctx InfraContext, clusterName string, name string) error

	OnVPNDeviceApplyError(ctx InfraContext, clusterName string, name string, errMsg string, opts UpdateAndDeleteOpts) error
	OnVPNDeviceDeleteMessage(ctx InfraContext, clusterName string, device entities.VPNDevice) error
	OnVPNDeviceUpdateMessage(ctx InfraContext, clusterName string, device entities.VPNDevice, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ListPVCs(ctx InfraContext, clusterName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.PersistentVolumeClaim], error)
	GetPVC(ctx InfraContext, clusterName string, pvcName string) (*entities.PersistentVolumeClaim, error)
	OnPVCUpdateMessage(ctx InfraContext, clusterName string, pvc entities.PersistentVolumeClaim, status types.ResourceStatus, opts UpdateAndDeleteOpts) error
	OnPVCDeleteMessage(ctx InfraContext, clusterName string, pvc entities.PersistentVolumeClaim) error

	ListClusterManagedServices(ctx InfraContext, clusterName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ClusterManagedService], error)
	GetClusterManagedService(ctx InfraContext, clusterName string, serviceName string) (*entities.ClusterManagedService, error)
	CreateClusterManagedService(ctx InfraContext, clusterName string, service entities.ClusterManagedService) (*entities.ClusterManagedService, error)
	UpdateClusterManagedService(ctx InfraContext, clusterName string, service entities.ClusterManagedService) (*entities.ClusterManagedService, error)
	DeleteClusterManagedService(ctx InfraContext, clusterName string, name string) error

	OnClusterManagedServiceApplyError(ctx InfraContext, clusterName, name, errMsg string, opts UpdateAndDeleteOpts) error
	OnClusterManagedServiceDeleteMessage(ctx InfraContext, clusterName string, service entities.ClusterManagedService) error
	OnClusterManagedServiceUpdateMessage(ctx InfraContext, clusterName string, service entities.ClusterManagedService, status types.ResourceStatus, opts UpdateAndDeleteOpts) error
}
