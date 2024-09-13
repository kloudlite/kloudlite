package domain

import (
	"context"
	"time"

	networkingv1 "k8s.io/api/networking/v1"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

type InfraContext struct {
	context.Context
	UserId      repos.ID
	UserEmail   string
	UserName    string
	AccountName string
}

func (i InfraContext) GetUserId() repos.ID {
	return i.UserId
}

func (i InfraContext) GetUserEmail() string {
	return i.UserEmail
}

func (i InfraContext) GetUserName() string {
	return i.UserName
}

type UpdateAndDeleteOpts struct {
	MessageTimestamp time.Time
}

type ResourceType string

const (
	ResourceTypeCluster               ResourceType = "cluster"
	ResourceTypeClusterGroup          ResourceType = "cluster_group"
	ResourceTypeBYOKCluster           ResourceType = "byok_cluster"
	ResourceTypeDomainEntries         ResourceType = "domain_entries"
	ResourceTypeHelmRelease           ResourceType = "helm_release"
	ResourceTypeNodePool              ResourceType = "nodepool"
	ResourceTypeClusterConnection     ResourceType = "cluster_connection"
	ResourceTypeClusterManagedService ResourceType = "cluster_managed_service"
	ResourceTypePVC                   ResourceType = "persistance_volume_claim"
	ResourceTypePV                    ResourceType = "persistance_volume"
	ResourceTypeVolumeAttachment      ResourceType = "volume_attachment"
)

type Domain interface {
	CheckNameAvailability(ctx InfraContext, typeArg ResType, clusterName *string, name string) (*CheckNameAvailabilityOutput, error)

	CreateGlobalVPN(ctx InfraContext, cluster entities.GlobalVPN) (*entities.GlobalVPN, error)
	UpdateGlobalVPN(ctx InfraContext, cluster entities.GlobalVPN) (*entities.GlobalVPN, error)
	DeleteGlobalVPN(ctx InfraContext, name string) error

	ListGlobalVPN(ctx InfraContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.GlobalVPN], error)
	GetGlobalVPN(ctx InfraContext, name string) (*entities.GlobalVPN, error)

	CreateGlobalVPNDevice(ctx InfraContext, device entities.GlobalVPNDevice) (*entities.GlobalVPNDevice, error)
	UpdateGlobalVPNDevice(ctx InfraContext, device entities.GlobalVPNDevice) (*entities.GlobalVPNDevice, error)
	DeleteGlobalVPNDevice(ctx InfraContext, gvpn string, device string) error

	ListGlobalVPNDevice(ctx InfraContext, gvpn string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.GlobalVPNDevice], error)
	GetGlobalVPNDevice(ctx InfraContext, gvpn string, device string) (*entities.GlobalVPNDevice, error)
	GetGlobalVPNDeviceWgConfig(ctx InfraContext, gvpn string, device string) (string, error)

	CreateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error)
	UpdateCluster(ctx InfraContext, cluster entities.Cluster) (*entities.Cluster, error)
	DeleteCluster(ctx InfraContext, name string) error

	CreateBYOKCluster(ctx InfraContext, cluster entities.BYOKCluster) (*entities.BYOKCluster, error)
	UpdateBYOKCluster(ctx InfraContext, clusterName string, displayName string) (*entities.BYOKCluster, error)
	ListBYOKCluster(ctx InfraContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.BYOKCluster], error)
	GetBYOKCluster(ctx InfraContext, name string) (*entities.BYOKCluster, error)
	GetBYOKClusterSetupInstructions(ctx InfraContext, name string, onlyHelmValues bool) ([]BYOKSetupInstruction, error)

	DeleteBYOKCluster(ctx InfraContext, name string) error
	UpsertBYOKClusterKubeconfig(ctx InfraContext, clusterName string, kubeconfig []byte) error

	UpgradeHelmKloudliteAgent(ctx InfraContext, clusterName string) error

	ListClusters(ctx InfraContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Cluster], error)
	GetCluster(ctx InfraContext, name string) (*entities.Cluster, error)

	GetClusterAdminKubeconfig(ctx InfraContext, clusterName string) (*string, error)

	OnClusterDeleteMessage(ctx InfraContext, cluster entities.Cluster) error
	OnClusterUpdateMessage(ctx InfraContext, cluster entities.Cluster, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	MarkClusterOnlineAt(ctx InfraContext, clusterName string, timestamp *time.Time) error

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

	OnNodePoolDeleteMessage(ctx InfraContext, clusterName string, nodePool entities.NodePool) error
	OnNodePoolUpdateMessage(ctx InfraContext, clusterName string, nodePool entities.NodePool, status types.ResourceStatus, opts UpdateAndDeleteOpts) error
	OnNodepoolApplyError(ctx InfraContext, clusterName string, name string, errMsg string, opts UpdateAndDeleteOpts) error

	// ListGlobalVPNs(ctx InfraContext, clusterName string) (*entities.GlobalVPNConnection, error)
	EnsureGlobalVPNConnection(ctx InfraContext, clusterName string, groupName string, dispatchAddr *entities.DispatchAddr) (*entities.GlobalVPNConnection, error)

	OnGlobalVPNConnectionDeleteMessage(ctx InfraContext, clusterName string, clusterConn entities.GlobalVPNConnection) error
	OnGlobalVPNConnectionUpdateMessage(ctx InfraContext, dispatchAddr entities.DispatchAddr, clusterConn entities.GlobalVPNConnection, status types.ResourceStatus, opts UpdateAndDeleteOpts) error
	OnGlobalVPNConnectionApplyError(ctx InfraContext, clusterName string, name string, errMsg string, opts UpdateAndDeleteOpts) error

	ListNodes(ctx InfraContext, clusterName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Node], error)
	GetNode(ctx InfraContext, clusterName string, nodeName string) (*entities.Node, error)

	OnNodeUpdateMessage(ctx InfraContext, clusterName string, node entities.Node) error
	OnNodeDeleteMessage(ctx InfraContext, clusterName string, node entities.Node) error

	ListHelmReleases(ctx InfraContext, clusterName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.HelmRelease], error)
	GetHelmRelease(ctx InfraContext, clusterName string, serviceName string) (*entities.HelmRelease, error)
	CreateHelmRelease(ctx InfraContext, clusterName string, service entities.HelmRelease) (*entities.HelmRelease, error)
	UpdateHelmRelease(ctx InfraContext, clusterName string, service entities.HelmRelease) (*entities.HelmRelease, error)

	DeleteHelmRelease(ctx InfraContext, clusterName string, name string) error
	OnHelmReleaseApplyError(ctx InfraContext, clusterName string, name string, errMsg string, opts UpdateAndDeleteOpts) error
	OnHelmReleaseDeleteMessage(ctx InfraContext, clusterName string, service entities.HelmRelease) error
	OnHelmReleaseUpdateMessage(ctx InfraContext, clusterName string, service entities.HelmRelease, status types.ResourceStatus, opts UpdateAndDeleteOpts) error

	ListManagedSvcTemplates() ([]*entities.MsvcTemplate, error)
	GetManagedSvcTemplate(category string, name string) (*entities.MsvcTemplateEntry, error)

	// kubernetes native resources
	ListPVCs(ctx InfraContext, clusterName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.PersistentVolumeClaim], error)
	GetPVC(ctx InfraContext, clusterName string, pvcName string) (*entities.PersistentVolumeClaim, error)
	OnPVCUpdateMessage(ctx InfraContext, clusterName string, pvc entities.PersistentVolumeClaim, status types.ResourceStatus, opts UpdateAndDeleteOpts) error
	OnPVCDeleteMessage(ctx InfraContext, clusterName string, pvc entities.PersistentVolumeClaim) error

	ListNamespaces(ctx InfraContext, clusterName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Namespace], error)
	GetNamespace(ctx InfraContext, clusterName string, namespace string) (*entities.Namespace, error)
	OnNamespaceUpdateMessage(ctx InfraContext, clusterName string, namespace entities.Namespace, status types.ResourceStatus, opts UpdateAndDeleteOpts) error
	OnNamespaceDeleteMessage(ctx InfraContext, clusterName string, namespace entities.Namespace) error

	ListPVs(ctx InfraContext, clusterName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.PersistentVolume], error)
	GetPV(ctx InfraContext, clusterName string, pvName string) (*entities.PersistentVolume, error)
	OnPVUpdateMessage(ctx InfraContext, clusterName string, pv entities.PersistentVolume, status types.ResourceStatus, opts UpdateAndDeleteOpts) error
	OnPVDeleteMessage(ctx InfraContext, clusterName string, pv entities.PersistentVolume) error
	DeletePV(ctx InfraContext, clusterName string, pvName string) error

	OnIngressUpdateMessage(ctx InfraContext, clusterName string, ingress networkingv1.Ingress, status types.ResourceStatus, opts UpdateAndDeleteOpts) error
	OnIngressDeleteMessage(ctx InfraContext, clusterName string, ingress networkingv1.Ingress) error

	ListVolumeAttachments(ctx InfraContext, clusterName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.VolumeAttachment], error)
	GetVolumeAttachment(ctx InfraContext, clusterName string, volAttachmentName string) (*entities.VolumeAttachment, error)
	OnVolumeAttachmentUpdateMessage(ctx InfraContext, clusterName string, volumeAttachment entities.VolumeAttachment, status types.ResourceStatus, opts UpdateAndDeleteOpts) error
	OnVolumeAttachmentDeleteMessage(ctx InfraContext, clusterName string, volumeAttachment entities.VolumeAttachment) error
}
