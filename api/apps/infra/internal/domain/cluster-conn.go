package domain

import (
	"fmt"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) reconClusterConns(ctx InfraContext, clusterGroup string) error {
	conns, err := d.clusterConnRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{fields.AccountName: ctx.AccountName, fc.ClusterClusterGroupName: clusterGroup},
	})

	if err != nil {
		return errors.NewE(err)
	}

	peers := make([]wgv1.Peer, 0)

	for _, c := range conns {
		if c.Spec.PublicKey == nil {
			continue
		}

		peers = append(peers, wgv1.Peer{
			PublicKey:  *c.Spec.PublicKey,
			Endpoint:   c.Endpoint,
			Id:         c.Spec.Id,
			AllowedIPs: []string{c.CIDR},
		})
	}

	for _, xcc := range conns {
		if fmt.Sprintf("%#v", xcc.Spec.Peers) == fmt.Sprintf("%#v", peers) {
			continue
		}

		xcc.Spec.Peers = peers
		unp, err := d.clusterConnRepo.Patch(
			ctx,
			repos.Filter{
				fields.AccountName:  ctx.AccountName,
				fields.ClusterName:  xcc.ClusterName,
				fields.MetadataName: xcc.Name,
			},
			common.PatchForUpdate(ctx, xcc, common.PatchOpts{XPatch: map[string]any{fc.ClusterConnectionSpecPeers: peers}}),
		)

		if err != nil {
			return errors.NewE(err)
		}

		if err := d.resDispatcher.ApplyToTargetCluster(ctx,
			unp.ClusterName,
			&unp.ClusterConnection,
			unp.RecordVersion,
		); err != nil {
			return errors.NewE(err)
		}
	}

	return nil
}

func (d *domain) findClusterConn(ctx InfraContext, clusterName string, connName string) (*entities.ClusterConnection, error) {
	cc, err := d.clusterConnRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
		fields.MetadataName: connName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if cc == nil {
		return nil, errors.Newf("cluster connection with name %q not found", clusterName)
	}
	return cc, nil
}

func (d *domain) findClusterConns(ctx InfraContext, clusterGroup string) ([]*entities.ClusterConnection, error) {
	cc, err := d.clusterConnRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			fields.AccountName:         ctx.AccountName,
			fc.ClusterClusterGroupName: clusterGroup,
		},
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	return cc, nil
}

func (d *domain) OnClusterConnDeleteMessage(ctx InfraContext, clusterName string, clusterConn entities.ClusterConnection) error {
	err := d.clusterConnRepo.DeleteOne(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.ClusterName:  clusterName,
			fields.MetadataName: clusterConn.Name,
		},
	)
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeClusterConnection, clusterConn.Name, PublishDelete)
	return err
}

func (d *domain) OnClusterConnUpdateMessage(ctx InfraContext, clusterName string, clusterConn entities.ClusterConnection, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xconn, err := d.findClusterConn(ctx, clusterName, clusterConn.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xconn == nil {
		return errors.Newf("no cluster connection found")
	}

	if _, err := d.matchRecordVersion(clusterConn.Annotations, xconn.RecordVersion); err != nil {
		return d.resyncToTargetCluster(ctx, xconn.SyncStatus.Action, clusterName, &xconn.ClusterConnection, xconn.RecordVersion)
	}

	recordVersion, err := d.matchRecordVersion(clusterConn.Annotations, xconn.RecordVersion)
	if err != nil {
		return errors.NewE(err)
	}

	unp, err := d.clusterConnRepo.PatchById(
		ctx,
		xconn.Id,
		common.PatchForSyncFromAgent(&clusterConn,
			recordVersion, status,
			common.PatchOpts{
				MessageTimestamp: opts.MessageTimestamp,
			}))
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.reconClusterConns(ctx, xconn.ClusterGroupName); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeClusterConnection, unp.Name, PublishUpdate)
	return nil
}

func (d *domain) OnClusterConnApplyError(ctx InfraContext, clusterName string, name string, errMsg string, opts UpdateAndDeleteOpts) error {
	unp, err := d.clusterConnRepo.Patch(
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
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeClusterConnection, unp.Name, PublishUpdate)
	return errors.NewE(err)
}
