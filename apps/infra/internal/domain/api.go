package domain

import (
	"context"

	"kloudlite.io/apps/infra/internal/entities"
	"kloudlite.io/pkg/repos"
)

type InfraContext struct {
	context.Context
	UserId      repos.ID
	UserEmail   string
	UserName    string
	AccountName string
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
	OnUpdateClusterMessage(ctx InfraContext, cluster entities.Cluster) error

	CreateBYOCCluster(ctx InfraContext, byocCluster entities.BYOCCluster) (*entities.BYOCCluster, error)
	UpdateBYOCCluster(ctx InfraContext, byocCluster entities.BYOCCluster) (*entities.BYOCCluster, error)
	DeleteBYOCCluster(ctx InfraContext, name string) error

	ListBYOCClusters(ctx InfraContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.BYOCCluster], error)
	GetBYOCCluster(ctx InfraContext, name string) (*entities.BYOCCluster, error)

	OnDeleteBYOCClusterMessage(ctx InfraContext, byocCluster entities.BYOCCluster) error
	OnUpdateBYOCClusterMessage(ctx InfraContext, byocCluster entities.BYOCCluster) error

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
	OnUpdateNodePoolMessage(ctx InfraContext, clusterName string, nodePool entities.NodePool) error
	OnNodepoolApplyError(ctx InfraContext, clusterName string, name string, errMsg string) error

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

	OnVPNDeviceApplyError(ctx InfraContext, clusterName string, name string, errMsg string) error
	OnVPNDeviceDeleteMessage(ctx InfraContext, clusterName string, device entities.VPNDevice) error
	OnVPNDeviceUpdateMessage(ctx InfraContext, clusterName string, device entities.VPNDevice) error

	ListPVCs(ctx InfraContext, clusterName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.PersistentVolumeClaim], error)
	GetPVC(ctx InfraContext, clusterName string, pvcName string) (*entities.PersistentVolumeClaim, error)
	OnPVCUpdateMessage(ctx InfraContext, clusterName string, pvc entities.PersistentVolumeClaim) error
	OnPVCDeleteMessage(ctx InfraContext, clusterName string, pvc entities.PersistentVolumeClaim) error

	ListBuildRuns(ctx InfraContext, repoName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.BuildRun], error)
	GetBuildRun(ctx InfraContext, repoName string, runName string) (*entities.BuildRun, error)
	OnBuildRunUpdateMessage(ctx InfraContext, clusterName string, buildRun entities.BuildRun) error
	OnBuildRunDeleteMessage(ctx InfraContext, clusterName string, buildRun entities.BuildRun) error
}
