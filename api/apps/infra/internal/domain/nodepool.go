package domain

import (
	iamT "github.com/kloudlite/api/apps/iam/types"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	clustersv1 "github.com/kloudlite/operator/apis/clusters/v1"
	ct "github.com/kloudlite/operator/apis/common-types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
)

func (d *domain) applyNodePool(ctx InfraContext, np *entities.NodePool) error {
	addTrackingId(&np.NodePool, np.Id)
	return d.resDispatcher.ApplyToTargetCluster(ctx, np.ClusterName, &np.NodePool, np.RecordVersion)
}

func (d *domain) CreateNodePool(ctx InfraContext, clusterName string, nodepool entities.NodePool) (*entities.NodePool, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateNodepool); err != nil {
		return nil, errors.NewE(err)
	}

	nodepool.IncrementRecordVersion()
	nodepool.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	nodepool.LastUpdatedBy = nodepool.CreatedBy

	cluster, err := d.findCluster(ctx, clusterName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	ps, err := d.findProviderSecret(ctx, cluster.Spec.CredentialsRef.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	switch nodepool.Spec.CloudProvider {
	case ct.CloudProviderAWS:
		{

			awsSubnetID := cluster.Spec.AWS.VPC.GetSubnetId(nodepool.Spec.AWS.AvailabilityZone)
			if awsSubnetID == "" {
				return nil, errors.Newf("kloudlite VPC has no subnet configured for this availability zone (%s), please select another availability zone in your cluster's region (%s)", nodepool.Spec.AWS.AvailabilityZone, cluster.Spec.AWS.Region)
			}

			nodepool.Spec.AWS = &clustersv1.AWSNodePoolConfig{
				VPCId:       cluster.Spec.AWS.VPC.ID,
				VPCSubnetID: awsSubnetID,

				AvailabilityZone: nodepool.Spec.AWS.AvailabilityZone,
				NvidiaGpuEnabled: nodepool.Spec.AWS.NvidiaGpuEnabled,
				RootVolumeType:   "gp3",
				RootVolumeSize: func() int {
					if nodepool.Spec.AWS.NvidiaGpuEnabled {
						return 80
					}
					return 50
				}(),
				IAMInstanceProfileRole: &ps.AWS.CfParamInstanceProfileName,
				PoolType:               nodepool.Spec.AWS.PoolType,
				EC2Pool:                nodepool.Spec.AWS.EC2Pool,
				SpotPool: func() *clustersv1.AwsSpotPoolConfig {
					if nodepool.Spec.AWS.SpotPool == nil {
						return nil
					}
					return &clustersv1.AwsSpotPoolConfig{
						SpotFleetTaggingRoleName: ps.AWS.CfParamRoleName,
						CpuNode:                  nodepool.Spec.AWS.SpotPool.CpuNode,
						GpuNode:                  nodepool.Spec.AWS.SpotPool.GpuNode,
						Nodes:                    nodepool.Spec.AWS.SpotPool.Nodes,
					}
				}(),
			}
		}
	default:
		{
			return nil, errors.Newf("cloudprovider: %s, currently not supported", nodepool.Spec.CloudProvider)
		}
	}

	nodepool.AccountName = ctx.AccountName
	nodepool.ClusterName = clusterName
	nodepool.SyncStatus = t.GenSyncStatus(t.SyncActionApply, nodepool.RecordVersion)

	nodepool.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &nodepool.NodePool); err != nil {
		return nil, errors.NewE(err)
	}
	nodepool.IncrementRecordVersion()

	np, err := d.nodePoolRepo.Create(ctx, &nodepool)
	if err != nil {
		if d.nodePoolRepo.ErrAlreadyExists(err) {
			return nil, errors.Newf("nodepool with name %q already exists", nodepool.Name)
		}
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeNodePool, np.Name, PublishAdd)

	if err := d.applyNodePool(ctx, np); err != nil {
		return nil, errors.NewE(err)
	}

	return np, nil
}

func (d *domain) UpdateNodePool(ctx InfraContext, clusterName string, nodePoolIn entities.NodePool) (*entities.NodePool, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateNodepool); err != nil {
		return nil, errors.NewE(err)
	}

	nodePoolIn.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &nodePoolIn.NodePool); err != nil {
		return nil, errors.NewE(err)
	}

	np, err := d.findNodePool(ctx, clusterName, nodePoolIn.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if np.IsStateful != nodePoolIn.IsStateful {
		return nil, errors.Newf("You can't change stateful value, aborting update")
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&nodePoolIn,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.NodePoolSpecMinCount: nodePoolIn.Spec.MinCount,
				fc.NodePoolSpecMaxCount: nodePoolIn.Spec.MaxCount,
			},
		})

	unp, err := d.nodePoolRepo.Patch(
		ctx,
		repos.Filter{
			fields.ClusterName:  clusterName,
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: nodePoolIn.Name,
		},
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeNodePool, unp.Name, PublishUpdate)

	if err := d.applyNodePool(ctx, unp); err != nil {
		return nil, errors.NewE(err)
	}

	return unp, nil
}

func (d *domain) DeleteNodePool(ctx InfraContext, clusterName string, poolName string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteNodepool); err != nil {
		return errors.NewE(err)
	}

	unp, err := d.nodePoolRepo.Patch(
		ctx,
		repos.Filter{
			fields.ClusterName:  clusterName,
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: poolName,
		},
		common.PatchForMarkDeletion(),
	)
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeNodePool, unp.Name, PublishUpdate)
	return d.resDispatcher.DeleteFromTargetCluster(ctx, clusterName, &unp.NodePool)
}

func (d *domain) GetNodePool(ctx InfraContext, clusterName string, poolName string) (*entities.NodePool, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetNodepool); err != nil {
		return nil, errors.NewE(err)
	}
	np, err := d.findNodePool(ctx, clusterName, poolName)
	if err != nil {
		return nil, errors.NewE(err)
	}
	return np, nil
}

func (d *domain) ListNodePools(ctx InfraContext, clusterName string, matchFilters map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.NodePool], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListNodepools); err != nil {
		return nil, errors.NewE(err)
	}
	filter := repos.Filter{
		fields.AccountName: ctx.AccountName,
		fields.ClusterName: clusterName,
	}
	return d.nodePoolRepo.FindPaginated(ctx, d.nodePoolRepo.MergeMatchFilters(filter, matchFilters), pagination)
}

func (d *domain) findNodePool(ctx InfraContext, clusterName string, poolName string) (*entities.NodePool, error) {
	np, err := d.nodePoolRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
		fields.MetadataName: poolName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if np == nil {
		return nil, errors.Newf("nodepool with name %q not found", clusterName)
	}
	return np, nil
}

func (d *domain) ResyncNodePool(ctx InfraContext, clusterName string, poolName string) error {
	if err := func() error {
		if err := d.canPerformActionInAccount(ctx, iamT.UpdateNodepool); err != nil {
			return d.canPerformActionInAccount(ctx, iamT.DeleteNodepool)
		}
		return nil
	}(); err != nil {
		return errors.NewE(err)
	}
	np, err := d.findNodePool(ctx, clusterName, poolName)
	if err != nil {
		return errors.NewE(err)
	}

	return d.resyncToTargetCluster(ctx, np.SyncStatus.Action, clusterName, &np.NodePool, np.RecordVersion)
}

// on message events

func (d *domain) OnNodePoolDeleteMessage(ctx InfraContext, clusterName string, nodePool entities.NodePool) error {
	err := d.nodePoolRepo.DeleteOne(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.ClusterName:  clusterName,
			fields.MetadataName: nodePool.Name,
		},
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeNodePool, nodePool.Name, PublishDelete)
	return err
}

func (d *domain) OnNodePoolUpdateMessage(ctx InfraContext, clusterName string, nodePool entities.NodePool, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xnp, err := d.findNodePool(ctx, clusterName, nodePool.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xnp == nil {
		return errors.Newf("no nodepool found")
	}

	if _, err := d.matchRecordVersion(nodePool.Annotations, xnp.RecordVersion); err != nil {
		return d.resyncToTargetCluster(ctx, xnp.SyncStatus.Action, clusterName, &xnp.NodePool, xnp.RecordVersion)
	}

	recordVersion, err := d.matchRecordVersion(nodePool.Annotations, xnp.RecordVersion)
	if err != nil {
		return errors.NewE(err)
	}

	unp, err := d.nodePoolRepo.PatchById(
		ctx,
		xnp.Id,
		common.PatchForSyncFromAgent(&nodePool,
			recordVersion, status,
			common.PatchOpts{
				MessageTimestamp: opts.MessageTimestamp,
			}))
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeNodePool, unp.GetName(), PublishUpdate)
	return nil
}

// OnNodepoolApplyError implements Domain.
func (d *domain) OnNodepoolApplyError(ctx InfraContext, clusterName string, name string, errMsg string, opts UpdateAndDeleteOpts) error {
	unp, err := d.nodePoolRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.ClusterName:  clusterName,
			fields.MetadataName: name,
		},
		common.PatchForErrorFromAgent(
			errMsg,
			common.PatchOpts{
				MessageTimestamp: opts.MessageTimestamp,
			},
		),
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeNodePool, unp.Name, PublishUpdate)
	return errors.NewE(err)
}
