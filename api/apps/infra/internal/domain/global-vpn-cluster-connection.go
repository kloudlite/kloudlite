package domain

import (
	"crypto/md5"
	"fmt"
	"math"
	"sort"
	"strings"

	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/iputils"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	gvpnConnectionDeviceMethod     = "gvpn-connection"
	kloudliteGlobalVPNDeviceMethod = "kloudlite-global-vpn-device"
)

type getGlobalVPNConnectionPeersArgs struct {
	ExcludeCluster       string
	GlobalVPNConnections []*entities.GlobalVPNConnection
	OnlyPublicClusters   bool
	OnlyPrivateClusters  bool
}

func (d *domain) getGlobalVPNConnectionPeers(args getGlobalVPNConnectionPeersArgs) []networkingv1.Peer {
	peers := make([]networkingv1.Peer, 0, len(args.GlobalVPNConnections))

	for _, c := range args.GlobalVPNConnections {
		if args.OnlyPublicClusters && c.Visibility.Mode == entities.ClusterVisibilityModePrivate {
			continue
		}

		if args.OnlyPrivateClusters && c.Visibility.Mode != entities.ClusterVisibilityModePrivate {
			continue
		}

		if c.ClusterName == args.ExcludeCluster {
			continue
		}

		if c.ParsedWgParams != nil {
			if c.ParsedWgParams.PublicKey == "" {
				continue
			}

			peer := networkingv1.Peer{
				DNSHostname: fmt.Sprintf("%s.device.local", c.Name),
				Comments:    fmt.Sprintf("gateway/%s/%s", c.GlobalVPNName, c.ClusterName),
				PublicKey:   c.ParsedWgParams.PublicKey,
				AllowedIPs:  []string{c.ClusterCIDR, fmt.Sprintf("%s/32", c.DeviceRef.IPAddr)},
				IP:          c.Spec.GlobalIP,
				DNSSuffix:   &c.Spec.DNSSuffix,
			}

			if c.Visibility.Mode != entities.ClusterVisibilityModePrivate {
				if c.Spec.LoadBalancer == nil {
					d.logger.Infof("loadbalancer not available for gvpn %s", c.Name)
					continue
				}

				endpoints := make([]string, 0, len(c.Spec.LoadBalancer.Hosts))
				for _, host := range c.Spec.LoadBalancer.Hosts {
					endpoints = append(endpoints, fmt.Sprintf("%s:%d", host, c.Spec.LoadBalancer.Port))
				}

				peer.PublicEndpoint = fn.New(strings.Join(endpoints, ", "))
			}

			peers = append(peers, peer)
		}
	}

	return peers
}

func (d *domain) listGlobalVPNConnections(ctx InfraContext, vpnName string) ([]*entities.GlobalVPNConnection, error) {
	return d.gvpnConnRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: vpnName,
		},
	})
}

func hashPeer(peer networkingv1.Peer) string {
	sort.Strings(peer.AllowedIPs)
	return fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s:%s:%s:%s:%s", peer.IP, peer.PublicKey, fn.DefaultIfNil(peer.PublicEndpoint), fn.DefaultIfNil(peer.DNSSuffix), strings.Join(peer.AllowedIPs, ",")))))
}

func hashPeers(peers []networkingv1.Peer) string {
	x := ""
	for _, p := range peers {
		x = fmt.Sprintf("%x", md5.Sum([]byte(x+hashPeer(p))))
	}
	return x
}

func (d *domain) reconGlobalVPNConnections(ctx InfraContext, vpnName string) error {
	gvpn, err := d.findGlobalVPN(ctx, vpnName)
	if err != nil {
		return errors.NewEf(err, "failed to find global vpn %s", vpnName)
	}

	gvpnConns, err := d.listGlobalVPNConnections(ctx, vpnName)
	if err != nil {
		return errors.NewE(err)
	}

	klDevice, err := d.findGlobalVPNDevice(ctx, gvpn.Name, gvpn.KloudliteDevice.Name)
	if err != nil {
		return errors.NewEf(err, "failed to find kloudlite device %s", gvpn.KloudliteDevice.Name)
	}

	klDevicePeer := d.buildPeerFromGlobalVPNDevice(ctx, gvpn, klDevice)

	// INFO: all private cluster gateway peers, are supposed to be routed via kloudlite gateway
	for _, c := range gvpnConns {
		if c.Visibility.Mode == entities.ClusterVisibilityModePrivate {
			klDevicePeer.AllowedIPs = append(klDevicePeer.AllowedIPs, c.ClusterCIDR)
		}
	}

	publicGVPNPeers := d.getGlobalVPNConnectionPeers(getGlobalVPNConnectionPeersArgs{
		GlobalVPNConnections: gvpnConns,
		OnlyPublicClusters:   true,
	})

	publicAllowedIPs := make([]string, 0, len(publicGVPNPeers))
	for i := range publicGVPNPeers {
		publicAllowedIPs = append(publicAllowedIPs, publicGVPNPeers[i].AllowedIPs...)
	}

	for _, xcc := range gvpnConns {
		peers := d.getGlobalVPNConnectionPeers(getGlobalVPNConnectionPeersArgs{
			GlobalVPNConnections: gvpnConns,
			ExcludeCluster:       xcc.ClusterName,
			OnlyPublicClusters:   true,
		})

		peers = append(peers, *klDevicePeer)
		if xcc.Visibility.Mode == entities.ClusterVisibilityModePrivate {
			peers = []networkingv1.Peer{*klDevicePeer}
			peers[0].AllowedIPs = append(peers[0].AllowedIPs, publicAllowedIPs...)
		}

		if hashPeers(xcc.Spec.Peers) == hashPeers(peers) {
			continue
		}

		xcc.Spec.Peers = peers
		unp, err := d.gvpnConnRepo.Patch(
			ctx,
			repos.Filter{
				fields.AccountName:  ctx.AccountName,
				fields.ClusterName:  xcc.ClusterName,
				fields.MetadataName: xcc.Name,
			},
			common.PatchForUpdate(ctx, xcc, common.PatchOpts{XPatch: map[string]any{fc.GlobalVPNConnectionSpecPeers: peers}}),
		)
		if err != nil {
			return errors.NewE(err)
		}

		if err := d.applyGlobalVPNConnection(ctx, unp); err != nil {
			return errors.NewE(err)
		}
	}

	return nil
}

func (d *domain) claimNextClusterCIDR(ctx InfraContext, clusterName string, gvpnName string) (string, error) {
	var cidrFilter *repos.MatchFilter
	for {
		filter := repos.Filter{
			fc.AccountName:                     ctx.AccountName,
			fc.FreeClusterSvcCIDRGlobalVPNName: gvpnName,
		}
		if cidrFilter != nil {
			filter = d.freeClusterSvcCIDRRepo.MergeMatchFilters(filter, map[string]repos.MatchFilter{fc.FreeClusterSvcCIDRClusterSvcCIDR: *cidrFilter})
		}

		freeCIDR, err := d.freeClusterSvcCIDRRepo.FindOne(ctx, filter)
		if err != nil {
			return "", err
		}

		if freeCIDR == nil {
			gvpn, err := d.findGlobalVPN(ctx, gvpnName)
			if err != nil {
				return "", err
			}

			numIPsPerCluster := int(math.Pow(2, float64(32-gvpn.AllocatableCIDRSuffix)))

			ipv4StartingAddr, err := iputils.GenIPAddr(gvpn.CIDR, gvpn.NumReservedIPsForNonClusterUse+gvpn.NumAllocatedClusterCIDRs*numIPsPerCluster)
			if err != nil {
				if errors.Is(err, iputils.ErrIPsMaxedOut) {
					return "", errors.NewEf(err, "max IPs exceeded, won't be able to allocate any more IPs")
				}
				return "", err
			}

			if _, err := d.freeClusterSvcCIDRRepo.Create(ctx, &entities.FreeClusterSvcCIDR{
				AccountName:    ctx.AccountName,
				GlobalVPNName:  gvpnName,
				ClusterSvcCIDR: fmt.Sprintf("%s/%d", ipv4StartingAddr, gvpn.AllocatableCIDRSuffix),
			}); err != nil {
				// FIXME: handle gracefully
				continue
			}
			if _, err := d.gvpnRepo.PatchById(ctx, gvpn.Id, repos.Document{"$inc": map[string]any{fc.GlobalVPNNumAllocatedClusterCIDRs: 1}}); err != nil {
				continue
			}
			continue
		}

		cidr := freeCIDR.ClusterSvcCIDR

		claimed, err := d.claimClusterSvcCIDRRepo.FindOne(ctx, repos.Filter{
			fc.AccountName:                         ctx.AccountName,
			fc.ClaimClusterSvcCIDRGlobalVPNName:    gvpnName,
			fc.ClaimClusterSvcCIDRClaimedByCluster: clusterName,
		})
		if err != nil {
			return "", err
		}

		if claimed != nil {
			return claimed.ClusterSvcCIDR, nil
		}

		if _, err := d.claimClusterSvcCIDRRepo.Create(ctx, &entities.ClaimClusterSvcCIDR{
			AccountName:      ctx.AccountName,
			GlobalVPNName:    gvpnName,
			ClusterSvcCIDR:   cidr,
			ClaimedByCluster: clusterName,
		}); err != nil {
			d.logger.Warnf("cluster svc CIDR %s, already claimed, trying another", cidr)
			if cidrFilter == nil {
				cidrFilter = &repos.MatchFilter{}
			}
			cidrFilter.MatchType = repos.MatchTypeNotInArray
			cidrFilter.NotInArray = append(cidrFilter.NotInArray, cidr)
			continue
		}

		if err := d.freeClusterSvcCIDRRepo.DeleteById(ctx, freeCIDR.Id); err != nil {
			return "", err
		}

		return cidr, nil
	}
}

func (d *domain) createGlobalVPNConnection(ctx InfraContext, gvpnConn entities.GlobalVPNConnection) (*entities.GlobalVPNConnection, error) {
	gvpnConn.CreatedBy = common.CreatedOrUpdatedByKloudlite
	gvpnConn.LastUpdatedBy = common.CreatedOrUpdatedByKloudlite

	gvpn, err := d.ensureGlobalVPN(ctx, gvpnConn.GlobalVPNName)
	if err != nil {
		return nil, err
	}

	// if gvpnConn.Spec.WgInterface == nil {
	// 	gvpnConn.Spec.WgInterface = &gvpn.WgInterface
	// }

	gvpnConn.SyncStatus = t.GenSyncStatus(t.SyncActionApply, 0)

	clusterCIDR, err := d.claimNextClusterCIDR(ctx, gvpnConn.ClusterName, gvpn.Name)
	if err != nil {
		return nil, err
	}

	sp := strings.SplitN(clusterCIDR, "/", 2)
	if len(sp) != 2 {
		return nil, errors.Newf("cluster CIDR %s is not in CIDR/N format", clusterCIDR)
	}

	svcCIDR := fmt.Sprintf("%s/%d", sp[0], d.env.AllocatableSvcCIDRSuffix)

	gvpnConn.ClusterCIDR = clusterCIDR

	gvpnDevice, err := d.createGlobalVPNDevice(ctx, entities.GlobalVPNDevice{
		ObjectMeta: metav1.ObjectMeta{
			Name: fmt.Sprintf("cluster-gateway-%s", gvpnConn.ClusterName),
		},
		AccountName:    ctx.AccountName,
		GlobalVPNName:  gvpnConn.Name,
		CreationMethod: gvpnConnectionDeviceMethod,
	})
	if err != nil {
		return nil, err
	}

	gvpnConn.DeviceRef = entities.GlobalVPNConnDeviceRef{
		Name:   gvpnDevice.Name,
		IPAddr: gvpnDevice.IPAddr,
	}

	gvpnConn.Gateway.Spec = networkingv1.GatewaySpec{
		GlobalIP:    gvpnDevice.IPAddr,
		ClusterCIDR: clusterCIDR,
		SvcCIDR:     svcCIDR,
		DNSSuffix:   fmt.Sprintf("svc.%s.local", gvpnConn.ClusterName),
		Peers:       nil,
	}
	gvpnConn.Gateway.EnsureGVK()

	gv, err := d.gvpnConnRepo.Create(ctx, &gvpnConn)
	if err != nil {
		return nil, err
	}

	if err := d.applyGlobalVPNConnection(ctx, gv); err != nil {
		return nil, err
	}

	return gv, nil
}

func (d *domain) deleteGlobalVPNConnection(ctx InfraContext, clusterName string, gvpnName string) error {
	gv, err := d.findGlobalVPNConnection(ctx, clusterName, gvpnName)
	if err != nil {
		if !errors.OfType[errors.ErrNotFound](err) {
			return errors.NewE(err)
		}
	}

	if gv == nil {
		// INFO: global vpn connection not found, nothing to do
		return nil
	}

	if err := d.deleteGlobalVPNDevice(ctx, gvpnName, gv.DeviceRef.Name); err != nil {
		return errors.NewE(err)
	}

	records, err := d.claimClusterSvcCIDRRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		fc.AccountName:                       ctx.AccountName,
		fc.ClaimClusterSvcCIDRClusterSvcCIDR: gv.ClusterCIDR,
	}})
	if err != nil {
		return errors.NewE(err)
	}

	for _, r := range records {
		if err := d.claimClusterSvcCIDRRepo.DeleteById(ctx, r.Id); err != nil {
			return errors.NewE(err)
		}

		if _, err := d.freeClusterSvcCIDRRepo.Create(ctx, &entities.FreeClusterSvcCIDR{
			AccountName:    r.AccountName,
			GlobalVPNName:  gvpnName,
			ClusterSvcCIDR: r.ClusterSvcCIDR,
		}); err != nil {
			return errors.NewE(err)
		}
	}

	if err := d.gvpnConnRepo.DeleteById(ctx, gv.Id); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) ensureGlobalVPNConnection(ctx InfraContext, clusterName string, groupName string) (*entities.GlobalVPNConnection, error) {
	gvpnConn, err := d.gvpnConnRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
		fields.MetadataName: groupName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if gvpnConn != nil {
		if err := d.applyGlobalVPNConnection(ctx, gvpnConn); err != nil {
			return nil, err
		}
		return gvpnConn, nil
	}

	gvpnGateway := networkingv1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: groupName}}
	gvpnGateway.EnsureGVK()

	return d.createGlobalVPNConnection(ctx, entities.GlobalVPNConnection{
		Gateway:          gvpnGateway,
		GlobalVPNName:    groupName,
		ResourceMetadata: common.ResourceMetadata{DisplayName: groupName, CreatedBy: common.CreatedOrUpdatedByKloudlite, LastUpdatedBy: common.CreatedOrUpdatedByKloudlite},
		AccountName:      ctx.AccountName,
		ClusterName:      clusterName,
		ParsedWgParams:   nil,
	})
}

func (d *domain) applyGlobalVPNConnection(ctx InfraContext, gvpn *entities.GlobalVPNConnection) error {
	// if err := d.resDispatcher.ApplyToTargetCluster(ctx, gvpn.ClusterName, &corev1.Secret{
	// 	TypeMeta: metav1.TypeMeta{
	// 		APIVersion: "v1",
	// 		Kind:       "Secret",
	// 	},
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      gvpn.Spec.WgRef.Name,
	// 		Namespace: gvpn.Spec.WgRef.Namespace,
	// 	},
	// 	StringData: map[string]string{
	// 		"ip": gvpn.DeviceRef.IPAddr,
	// 	},
	// }, 0); err != nil {
	// 	return err
	// }
	return d.resDispatcher.ApplyToTargetCluster(ctx, gvpn.ClusterName, &gvpn.Gateway, gvpn.RecordVersion)
}

func (d *domain) findGlobalVPNConnection(ctx InfraContext, clusterName string, groupName string) (*entities.GlobalVPNConnection, error) {
	cc, err := d.gvpnConnRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
		fields.MetadataName: groupName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if cc == nil {
		return nil, errors.ErrNotFound{Message: "global vpn connection not found"}
	}
	return cc, nil
}

func (d *domain) OnGlobalVPNConnectionDeleteMessage(ctx InfraContext, clusterName string, gvpnConn entities.GlobalVPNConnection) error {
	currRecord, err := d.gvpnConnRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
		fields.MetadataName: gvpnConn.Name,
	})
	if err != nil {
		return err
	}

	if err := d.gvpnConnRepo.DeleteOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
		fields.MetadataName: gvpnConn.Name,
	}); err != nil {
		return errors.NewE(err)
	}

	if currRecord != nil && currRecord.DeviceRef.Name != "" {
		if err := d.deleteGlobalVPNDevice(ctx, currRecord.GlobalVPNName, currRecord.DeviceRef.Name); err != nil {
			return errors.NewE(err)
		}
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeClusterConnection, gvpnConn.Name, PublishDelete)
	return err
}

func (d *domain) OnGlobalVPNConnectionUpdateMessage(ctx InfraContext, clusterName string, gvpn entities.GlobalVPNConnection, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xconn, err := d.findGlobalVPNConnection(ctx, clusterName, gvpn.Name)
	if err != nil {
		return errors.NewE(err)
	}

	// INFO: BYOK cluster does not have any status update message
	if d.isBYOKCluster(ctx, xconn.ClusterName) {
		if _, err := d.byokClusterRepo.PatchOne(ctx, entities.UniqueBYOKClusterFilter(ctx.AccountName, clusterName), repos.Document{
			fc.SyncStatusState:        t.SyncStateUpdatedAtAgent,
			fc.SyncStatusLastSyncedAt: opts.MessageTimestamp,
			fc.SyncStatusError:        nil,
		}); err != nil {
			return errors.NewE(err)
		}
	}

	if _, err := d.matchRecordVersion(gvpn.Annotations, xconn.RecordVersion); err != nil {
		return d.resyncToTargetCluster(ctx, xconn.SyncStatus.Action, clusterName, &xconn.Gateway, xconn.RecordVersion)
	}

	recordVersion, err := d.matchRecordVersion(gvpn.Annotations, xconn.RecordVersion)
	if err != nil {
		return errors.NewE(err)
	}

	patchDoc := common.PatchForSyncFromAgent(&gvpn, recordVersion, status, common.PatchOpts{
		MessageTimestamp: opts.MessageTimestamp,
	})

	if gvpn.ParsedWgParams != nil {
		patchDoc[fc.GlobalVPNConnectionParsedWgParams] = gvpn.ParsedWgParams
		patchDoc[fc.GlobalVPNConnectionSpecAdminNamespace] = gvpn.Spec.AdminNamespace
		patchDoc[fc.GlobalVPNConnectionSpecLoadBalancer] = gvpn.Spec.LoadBalancer
		patchDoc[fc.GlobalVPNConnectionSpecWireguardKeysRef] = gvpn.Spec.WireguardKeysRef
	}

	ugvpn, err := d.gvpnConnRepo.PatchById(ctx, xconn.Id, patchDoc)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.reconGlobalVPNConnections(ctx, xconn.Name); err != nil {
		return errors.NewE(err)
	}

	// FIXME: move to sync kloudlite gateway
	// if err := d.syncKloudliteDeviceOnCluster(ctx, xconn.GlobalVPNName); err != nil {
	// 	return errors.NewE(err)
	// }

	if err := d.syncKloudliteGateway(ctx, xconn.GlobalVPNName); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeClusterConnection, ugvpn.Name, PublishUpdate)
	return nil
}

func (d *domain) OnGlobalVPNConnectionApplyError(ctx InfraContext, clusterName string, name string, errMsg string, opts UpdateAndDeleteOpts) error {
	unp, err := d.gvpnConnRepo.Patch(
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
