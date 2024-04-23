package domain

import (
	"fmt"
	"math"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	common_types "github.com/kloudlite/operator/apis/common-types"
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *domain) reconGlobalVPNs(ctx InfraContext, vpnName string) error {
	vpns, err := d.globalVPNRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: vpnName,
		},
	})
	if err != nil {
		return errors.NewE(err)
	}

	peers := make([]wgv1.Peer, 0)

	for _, c := range vpns {
		if c.ParsedWgParams != nil {
			if c.ParsedWgParams.WgPublicKey == "" {
				continue
			}

			if c.ParsedWgParams.NodePort == nil {
				d.logger.Infof("nodeport not available for gvpn %s", c.Name)
				continue
			}

			if c.CIDR == "" {
				d.logger.Infof("cidr not available for gvpn %s", c.Name)
				continue
			}

			peers = append(peers, wgv1.Peer{
				IP:         c.ParsedWgParams.IP,
				PublicKey:  c.ParsedWgParams.WgPublicKey,
				Endpoint:   fmt.Sprintf("%s:%s", c.ClusterPublicEndpoint, *c.ParsedWgParams.NodePort),
				AllowedIPs: []string{c.CIDR},
			})
		}
	}

	for _, xcc := range vpns {
		if fmt.Sprintf("%#v", xcc.Spec.Peers) == fmt.Sprintf("%#v", peers) {
			continue
		}

		xcc.Spec.Peers = peers
		unp, err := d.globalVPNRepo.Patch(
			ctx,
			repos.Filter{
				fields.AccountName:  ctx.AccountName,
				fields.ClusterName:  xcc.ClusterName,
				fields.MetadataName: xcc.Name,
			},
			common.PatchForUpdate(ctx, xcc, common.PatchOpts{XPatch: map[string]any{fc.GlobalVPNSpecPeers: peers}}),
		)
		if err != nil {
			return errors.NewE(err)
		}

		if err := d.applyGlobalVPN(ctx, unp); err != nil {
			return errors.NewE(err)
		}
	}

	return nil
}

func (d *domain) createGlobalVPN(ctx InfraContext, gvpn entities.GlobalVPN) (*entities.GlobalVPN, error) {
	gvpn.ResourceMetadata.CreatedBy = common.CreatedOrUpdatedByKloudlite
	gvpn.ResourceMetadata.LastUpdatedBy = common.CreatedOrUpdatedByKloudlite

	if gvpn.CIDR == "" {
		gvpn.CIDR = d.env.BaseCIDR
	}

	if gvpn.AllocatableCIDRSuffix == 0 {
		gvpn.AllocatableCIDRSuffix = d.env.AllocatableCIDRSuffix
	}

	if gvpn.ClusterOffset == 0 {
		gvpn.ClusterOffset = d.env.ClustersOffset
	}

	if gvpn.Spec.WgInterface == nil {
		gvpn.Spec.WgInterface = fn.New("kl0")
	}

	gvpn.SyncStatus = t.GenSyncStatus(t.SyncActionApply, 0)

	addrPool, err := d.findDeviceAddressPool(ctx, gvpn.Name)
	if err != nil {
		return nil, err
	}

	if addrPool == nil {
		if _, err := d.createDeviceAddressPool(ctx, entities.GlobalVPNDeviceAddressPool{
			AccountName:   ctx.AccountName,
			GlobalVPNName: gvpn.Name,
			CIDR:          gvpn.CIDR,
			MinOffset:     0,
			MaxOffset:     gvpn.ClusterOffset * int(math.Pow(2, float64(32-gvpn.AllocatableCIDRSuffix))),
			RunningOffset: 0,
		}); err != nil {
			return nil, err
		}
	}

	gatewayAddr, err := d.getNextDeviceAddress(ctx, gvpn.Name)
	if err != nil {
		return nil, err
	}

	gvpn.GatewayIPAddr = gatewayAddr

	gv, err := d.globalVPNRepo.Create(ctx, &gvpn)
	if err != nil {
		return nil, err
	}

	if err := d.applyGlobalVPN(ctx, gv); err != nil {
		return nil, err
	}

	return gv, nil
}

func (d *domain) ensureGlobalVPN(ctx InfraContext, clusterName string, groupName string) (*entities.GlobalVPN, error) {
	gvpn, err := d.globalVPNRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
		fields.MetadataName: groupName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if gvpn != nil {
		if err := d.applyGlobalVPN(ctx, gvpn); err != nil {
			return nil, err
		}
		return gvpn, nil
	}

	return d.createGlobalVPN(ctx, entities.GlobalVPN{
		BaseEntity: repos.BaseEntity{},
		GlobalVPN: wgv1.GlobalVPN{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "wireguard.kloudlite.io/v1",
				Kind:       "GlobalVPN",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: groupName,
			},
			Spec: wgv1.GlobVPNSpec{
				WgRef: common_types.SecretRef{
					Name:      fmt.Sprintf("global-vpn-params-%s", groupName),
					Namespace: "kube-system",
				},
			},
		},
		ResourceMetadata:      common.ResourceMetadata{DisplayName: groupName, CreatedBy: common.CreatedOrUpdatedByKloudlite, LastUpdatedBy: common.CreatedOrUpdatedByKloudlite},
		AccountName:           ctx.AccountName,
		ClusterName:           clusterName,
		ClusterPublicEndpoint: fmt.Sprintf("%s.%s.tenants.%s", clusterName, ctx.AccountName, d.env.PublicDNSHostSuffix),
		ParsedWgParams:        nil,
	})
}

func (d *domain) applyGlobalVPN(ctx InfraContext, gvpn *entities.GlobalVPN) error {
	if err := d.resDispatcher.ApplyToTargetCluster(ctx, gvpn.ClusterName, &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      gvpn.Spec.WgRef.Name,
			Namespace: gvpn.Spec.WgRef.Namespace,
		},
		StringData: map[string]string{
			"ip": gvpn.GatewayIPAddr,
		},
	}, 0); err != nil {
		return err
	}
	return d.resDispatcher.ApplyToTargetCluster(ctx, gvpn.ClusterName, &gvpn.GlobalVPN, gvpn.RecordVersion)
}

func (d *domain) findGlobalVPN(ctx InfraContext, clusterName string, groupName string) (*entities.GlobalVPN, error) {
	cc, err := d.globalVPNRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
		fields.MetadataName: groupName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if cc == nil {
		return nil, errors.Newf("global vpn with name (%s) not found, for cluster (%s)", groupName, clusterName)
	}
	return cc, nil
}

func (d *domain) OnGlobalVPNDeleteMessage(ctx InfraContext, clusterName string, gvpn entities.GlobalVPN) error {
	currRecord, err := d.globalVPNRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
		fields.MetadataName: gvpn.Name,
	})
	if err != nil {
		return err
	}

	if err := d.globalVPNRepo.DeleteOne(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.ClusterName:  clusterName,
			fields.MetadataName: gvpn.Name,
		},
	); err != nil {
		return errors.NewE(err)
	}

	if err := d.addToFreeAddressPool(ctx, gvpn.Name, currRecord.GatewayIPAddr); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeClusterConnection, gvpn.Name, PublishDelete)
	return err
}

func (d *domain) OnGlobalVPNUpdateMessage(ctx InfraContext, clusterName string, gvpn entities.GlobalVPN, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xconn, err := d.findGlobalVPN(ctx, clusterName, gvpn.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xconn == nil {
		return errors.Newf("no global vpn found")
	}

	if _, err := d.matchRecordVersion(gvpn.Annotations, xconn.RecordVersion); err != nil {
		return d.resyncToTargetCluster(ctx, xconn.SyncStatus.Action, clusterName, &xconn.GlobalVPN, xconn.RecordVersion)
	}

	recordVersion, err := d.matchRecordVersion(gvpn.Annotations, xconn.RecordVersion)
	if err != nil {
		return errors.NewE(err)
	}

	patchDoc := common.PatchForSyncFromAgent(&gvpn,
		recordVersion, status,
		common.PatchOpts{
			MessageTimestamp: opts.MessageTimestamp,
		})

	if gvpn.ParsedWgParams != nil {
		patchDoc[fc.GlobalVPNParsedWgParams] = gvpn.ParsedWgParams
	}

	ugvpn, err := d.globalVPNRepo.PatchById(ctx, xconn.Id, patchDoc)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.reconGlobalVPNs(ctx, xconn.Name); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeClusterConnection, ugvpn.Name, PublishUpdate)
	return nil
}

func (d *domain) OnGlobalVPNApplyError(ctx InfraContext, clusterName string, name string, errMsg string, opts UpdateAndDeleteOpts) error {
	unp, err := d.globalVPNRepo.Patch(
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
