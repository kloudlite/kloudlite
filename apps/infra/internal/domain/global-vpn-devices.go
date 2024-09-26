package domain

import (
	"fmt"
	"time"

	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/wgutils"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/iputils"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/repos"
	networkingv1 "github.com/kloudlite/operator/apis/networking/v1"
)

func (d *domain) claimNextFreeDeviceIP(ctx InfraContext, deviceName string, gvpnName string) (string, error) {
	var ipAddrFilter *repos.MatchFilter
	offsetIdx := 0
	for {
		filter := repos.Filter{
			fields.AccountName:           ctx.AccountName,
			fc.FreeDeviceIPGlobalVPNName: gvpnName,
		}
		if ipAddrFilter != nil {
			filter = d.freeDeviceIpRepo.MergeMatchFilters(filter, map[string]repos.MatchFilter{fc.FreeDeviceIPIpAddr: *ipAddrFilter})
		}
		freeIp, err := d.freeDeviceIpRepo.FindOne(ctx, filter)
		if err != nil {
			return "", err
		}

		if freeIp == nil {
			gvpn, err := d.findGlobalVPN(ctx, gvpnName)
			if err != nil {
				return "", err
			}

			ip, err := iputils.GetIPAddrInARange(gvpn.CIDR, gvpn.NumAllocatedDevices+1+offsetIdx, gvpn.NumReservedIPsForNonClusterUse)
			if err != nil {
				return "", err
			}

			if _, err := d.freeDeviceIpRepo.Create(ctx, &entities.FreeDeviceIP{
				AccountName:   ctx.AccountName,
				GlobalVPNName: gvpnName,
				IPAddr:        ip,
			}); err != nil {
				offsetIdx += 1
				continue
			}

			if _, err := d.gvpnRepo.PatchById(ctx, gvpn.Id, repos.Document{"$inc": map[string]any{fc.GlobalVPNNumAllocatedDevices: 1}}); err != nil {
				continue
			}

			continue
		}

		ipAddr := freeIp.IPAddr

		claimed, err := d.claimDeviceIPRepo.FindOne(ctx, repos.Filter{
			fc.AccountName:                ctx.AccountName,
			fc.ClaimDeviceIPGlobalVPNName: gvpnName,
			fc.ClaimDeviceIPClaimedBy:     deviceName,
		})
		if err != nil {
			return "", err
		}

		if claimed != nil {
			return claimed.IPAddr, nil
		}

		if _, err := d.claimDeviceIPRepo.Create(ctx, &entities.ClaimDeviceIP{
			AccountName:   ctx.AccountName,
			GlobalVPNName: gvpnName,
			IPAddr:        ipAddr,
			ClaimedBy:     deviceName,
		}); err != nil {
			if ipAddrFilter == nil {
				ipAddrFilter = &repos.MatchFilter{}
			}
			ipAddrFilter.MatchType = repos.MatchTypeNotInArray
			ipAddrFilter.NotInArray = append(ipAddrFilter.NotInArray, ipAddr)

			d.logger.Warn("ip addr already claimed, will retry again, got", "err", err)
			<-time.After(50 * time.Millisecond)
			continue
		}

		if err := d.freeDeviceIpRepo.DeleteById(ctx, freeIp.Id); err != nil {
			return "", err
		}

		return ipAddr, nil
	}
}

func (d *domain) UpdateGlobalVPNDevice(ctx InfraContext, device entities.GlobalVPNDevice) (*entities.GlobalVPNDevice, error) {
	panic("implement me")
}

func (d *domain) deleteGlobalVPNDevice(ctx InfraContext, gvpn string, deviceName string) error {
	device, err := d.findGlobalVPNDevice(ctx, gvpn, deviceName)
	if err != nil {
		if errors.OfType[errors.ErrNotFound](err) {
			return nil
		}
		return err
	}

	if err := d.claimDeviceIPRepo.DeleteOne(ctx, repos.Filter{
		fc.AccountName:                  ctx.AccountName,
		fc.GlobalVPNDeviceGlobalVPNName: gvpn,
		fc.ClaimDeviceIPClaimedBy:       deviceName,
	}); err != nil {
		return err
	}

	if _, err := d.freeDeviceIpRepo.Create(ctx, &entities.FreeDeviceIP{
		AccountName:   ctx.AccountName,
		GlobalVPNName: gvpn,
		IPAddr:        device.IPAddr,
	}); err != nil {
		return err
	}

	if err := d.gvpnDevicesRepo.DeleteById(ctx, device.Id); err != nil {
		return err
	}

	if err := d.syncKloudliteDeviceOnPlatform(ctx, gvpn); err != nil {
		return err
	}

	if err := d.syncKloudliteGatewayDevice(ctx, gvpn); err != nil {
		return err
	}

	return nil
}

func (d *domain) DeleteGlobalVPNDevice(ctx InfraContext, gvpn string, deviceName string) error {
	return d.deleteGlobalVPNDevice(ctx, gvpn, deviceName)
}

func (d *domain) ListGlobalVPNDevice(ctx InfraContext, gvpn string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.GlobalVPNDevice], error) {
	filter := d.gvpnDevicesRepo.MergeMatchFilters(
		repos.Filter{
			fc.AccountName:                  ctx.AccountName,
			fc.GlobalVPNDeviceGlobalVPNName: gvpn,
		},
		map[string]repos.MatchFilter{
			fc.CreatedByUserId: {
				MatchType: repos.MatchTypeExact,
				Exact:     ctx.UserId,
			},
			fc.GlobalVPNDeviceCreationMethod: {
				MatchType:  repos.MatchTypeNotInArray,
				NotInArray: []any{gvpnConnectionDeviceMethod, kloudliteGlobalVPNDevice},
			},
		},
		search,
	)
	return d.gvpnDevicesRepo.FindPaginated(ctx, filter, pagination)
}

func (d *domain) CreateGlobalVPNDevice(ctx InfraContext, gvpnDevice entities.GlobalVPNDevice) (*entities.GlobalVPNDevice, error) {
	return d.createGlobalVPNDevice(ctx, gvpnDevice)
}

func (d *domain) createGlobalVPNDevice(ctx InfraContext, gvpnDevice entities.GlobalVPNDevice) (*entities.GlobalVPNDevice, error) {
	gvpnDevice.AccountName = ctx.AccountName
	gvpnDevice.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	gvpnDevice.LastUpdatedBy = gvpnDevice.CreatedBy

	privateKey, publicKey, err := wgutils.GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	gvpnDevice.PrivateKey = privateKey
	gvpnDevice.PublicKey = publicKey

	ip, err := d.claimNextFreeDeviceIP(ctx, gvpnDevice.Name, gvpnDevice.GlobalVPNName)
	if err != nil {
		return nil, err
	}

	gvpnDevice.IPAddr = ip

	gv, err := d.gvpnDevicesRepo.Upsert(ctx, repos.Filter{
		fc.AccountName:                  ctx.AccountName,
		fc.GlobalVPNDeviceGlobalVPNName: gvpnDevice.GlobalVPNName,
		fc.MetadataName:                 gvpnDevice.Name,
	}, &gvpnDevice)
	if err != nil {
		return nil, err
	}

	if err := d.syncKloudliteDeviceOnPlatform(ctx, gvpnDevice.GlobalVPNName); err != nil {
		return nil, err
	}

	if err := d.syncKloudliteGatewayDevice(ctx, gvpnDevice.GlobalVPNName); err != nil {
		return nil, err
	}

	return gv, nil
}

func (d *domain) buildPeerFromGlobalVPNDevice(_ InfraContext, gvpn *entities.GlobalVPN, device *entities.GlobalVPNDevice) *networkingv1.Peer {
	allowedIPs := []string{fmt.Sprintf("%s/32", device.IPAddr)}

	// privateConns, err := d.gvpnConnRepo.Find(ctx, repos.Query{
	// 	Filter: repos.Filter{
	// 		fc.GlobalVPNConnectionGlobalVPNName:  gvpn.Name,
	// 		fc.GlobalVPNConnectionVisibilityMode: entities.ClusterVisibilityModePrivate,
	// 	},
	// })
	// if err != nil {
	// 	return nil
	// }

	// privateCIDRs := make([]string, 0, len(privateConns))
	// for _, conn := range privateConns {
	// 	privateCIDRs = append(privateCIDRs, conn.ClusterCIDR)
	// }

	if device.IPAddr == gvpn.KloudliteGatewayDevice.IPAddr {
		// FIXME: this should not be used
		// allowedIPs = append(allowedIPs, gvpn.NonClusterUseAllowedIPs...)

		// NOTE: i don't even remember what is the use case of non-cluster-use allowed IPs
		allowedIPs = append(allowedIPs, "100.64.0.0/10")
		// allowedIPs = append(allowedIPs, privateCIDRs...)
	}

	return &networkingv1.Peer{
		Comments:       fmt.Sprintf("device/%s/%s", device.GlobalVPNName, device.Name),
		DNSHostname:    fmt.Sprintf("%s.device.local", device.Name),
		PublicKey:      device.PublicKey,
		PublicEndpoint: device.PublicEndpoint,
		IP:             &device.IPAddr,
		DNSSuffix:      nil,
		AllowedIPs:     allowedIPs,
	}
}

func (d *domain) buildPeersFromGlobalVPNDevices(ctx InfraContext, gvpn string) (publicPeers []networkingv1.Peer, privatePeers []networkingv1.Peer, err error) {
	devices, err := d.gvpnDevicesRepo.Find(ctx, repos.Query{
		Filter: map[string]any{
			fc.AccountName:                  ctx.AccountName,
			fc.GlobalVPNDeviceGlobalVPNName: gvpn,
			fc.GlobalVPNDeviceCreationMethod: map[string]any{
				"$ne": gvpnConnectionDeviceMethod,
			},
		},
	})
	if err != nil {
		return nil, nil, err
	}

	gv, err := d.findGlobalVPN(ctx, gvpn)
	if err != nil {
		return nil, nil, err
	}

	publicPeers = make([]networkingv1.Peer, 0, 10) // 10 is just a random low number
	privatePeers = make([]networkingv1.Peer, 0, len(devices))
	for i := range devices {
		allowedIPs := []string{fmt.Sprintf("%s/32", devices[i].IPAddr)}
		if devices[i].PublicEndpoint != nil {
			allowedIPs := []string{fmt.Sprintf("%s/32", devices[i].IPAddr)}

			if devices[i].Name == gv.KloudliteGatewayDevice.Name {
				allowedIPs = append(allowedIPs, gv.NonClusterUseAllowedIPs...)
			}

			publicPeers = append(publicPeers, networkingv1.Peer{
				Comments:       fmt.Sprintf("device/%s/%s", devices[i].GlobalVPNName, devices[i].Name),
				DNSHostname:    fmt.Sprintf("%s.device.local", devices[i].Name),
				PublicKey:      devices[i].PublicKey,
				PublicEndpoint: devices[i].PublicEndpoint,
				IP:             &devices[i].IPAddr,
				DNSSuffix:      nil,
				AllowedIPs:     allowedIPs,
			})
			continue
		}

		privatePeers = append(privatePeers, networkingv1.Peer{
			Comments:    fmt.Sprintf("device/%s/%s", devices[i].GlobalVPNName, devices[i].Name),
			DNSHostname: fmt.Sprintf("%s.device.local", devices[i].Name),
			PublicKey:   devices[i].PublicKey,
			IP:          &devices[i].IPAddr,
			DNSSuffix:   nil,
			AllowedIPs:  allowedIPs,
		})
	}

	return publicPeers, privatePeers, nil
}

func (d *domain) GetGlobalVPNDevice(ctx InfraContext, gvpn string, gvpnDevice string) (*entities.GlobalVPNDevice, error) {
	if gvpn == "" || gvpnDevice == "" {
		return nil, errors.New("invalid global vpn or device")
	}

	return d.findGlobalVPNDevice(ctx, gvpn, gvpnDevice)
}

func (d *domain) GetGlobalVPNDeviceWgConfig(ctx InfraContext, gvpn string, gvpnDevice string) (string, error) {
	return d.getGlobalVPNDeviceWgConfig(ctx, gvpn, gvpnDevice, nil)
}

func (d *domain) buildGlobalVPNDeviceWgBaseParams(ctx InfraContext, gvpnConns []*entities.GlobalVPNConnection, gvpnDevice *entities.GlobalVPNDevice) (wgparams *wgutils.WgConfigParams, deviceHosts map[string]string, err error) {
	gvpnConnPeers := d.getGlobalVPNConnectionPeers(getGlobalVPNConnectionPeersArgs{
		GlobalVPNConnections: gvpnConns,
	})

	deviceHosts = make(map[string]string)

	pubPeers, privPeers, err := d.buildPeersFromGlobalVPNDevices(ctx, gvpnDevice.GlobalVPNName)
	if err != nil {
		return nil, deviceHosts, err
	}

	publicPeers := make([]wgutils.PublicPeer, 0, len(pubPeers)+len(gvpnConnPeers))

	for _, peer := range gvpnConnPeers {
		if peer.IP != nil {
			deviceHosts[peer.DNSHostname] = *peer.IP
		}
		if peer.DNSHostname == fmt.Sprintf("%s.device.local", gvpnDevice.Name) {
			continue
		}
		if peer.PublicEndpoint == nil {
			privPeers = append(privPeers, peer)
			continue
		}

		publicPeers = append(publicPeers, wgutils.PublicPeer{
			DisplayName: fmt.Sprintf("Global VPN Gateway (%s)", peer.Comments),
			PublicKey:   peer.PublicKey,
			AllowedIPs:  peer.AllowedIPs,
			Endpoint:    *peer.PublicEndpoint,
		})
	}

	for _, peer := range pubPeers {
		if peer.IP != nil {
			deviceHosts[peer.DNSHostname] = *peer.IP
		}
		if peer.DNSHostname == fmt.Sprintf("%s.device.local", gvpnDevice.Name) {
			continue
		}
		if peer.PublicEndpoint == nil {
			privPeers = append(privPeers, peer)
			continue
		}

		publicPeers = append(publicPeers, wgutils.PublicPeer{
			// DisplayName: peer.DNSHostname,
			DisplayName: fmt.Sprintf("Global VPN Device [%s] (%s)", peer.Comments, peer.DNSHostname),
			PublicKey:   peer.PublicKey,
			AllowedIPs:  peer.AllowedIPs,
			Endpoint:    *peer.PublicEndpoint,
		})
	}

	privatePeers := make([]wgutils.PrivatePeer, 0, len(privPeers))
	for _, peer := range privPeers {
		if peer.IP != nil {
			deviceHosts[peer.DNSHostname] = *peer.IP
		}
		if peer.DNSHostname == fmt.Sprintf("%s.device.local", gvpnDevice.Name) {
			continue
		}
		privatePeers = append(privatePeers, wgutils.PrivatePeer{
			DisplayName: fmt.Sprintf("Global VPN Device [%s] (%s)", peer.Comments, peer.DNSHostname),
			// DisplayName: peer.DNSHostname,
			PublicKey:  peer.PublicKey,
			AllowedIPs: peer.AllowedIPs,
		})
	}

	return &wgutils.WgConfigParams{
		IPAddr:       gvpnDevice.IPAddr,
		PrivateKey:   gvpnDevice.PrivateKey,
		PublicPeers:  publicPeers,
		PrivatePeers: privatePeers,
	}, deviceHosts, nil
}

func (d *domain) getGlobalVPNDeviceWgConfig(ctx InfraContext, gvpn string, gvpnDevice string, postUp []string) (string, error) {
	device, err := d.findGlobalVPNDevice(ctx, gvpn, gvpnDevice)
	if err != nil {
		return "", err
	}

	gv, err := d.findGlobalVPN(ctx, gvpn)
	if err != nil {
		return "", err
	}

	klDevice, err := d.findGlobalVPNDevice(ctx, gvpn, gv.KloudliteGatewayDevice.Name)
	if err != nil {
		return "", err
	}

	if klDevice.PublicEndpoint == nil {
		return "", errors.New("kloudlite device public endpoint is nil, please wait for some time")
	}

	config, err := wgutils.GenerateWireguardConfig(wgutils.WgConfigParams{
		IPAddr:     device.IPAddr,
		PrivateKey: device.PrivateKey,
		DNS:        gv.KloudliteGatewayDevice.IPAddr,
		PostUp:     postUp,
		PublicPeers: []wgutils.PublicPeer{
			{
				PublicKey:  klDevice.PublicKey,
				Endpoint:   *klDevice.PublicEndpoint,
				AllowedIPs: []string{gv.CIDR},
			},
		},
	})
	if err != nil {
		return "", err
	}

	return config, nil
}

func (d *domain) findGlobalVPNDevice(ctx InfraContext, gvpn string, gvpnDevice string) (*entities.GlobalVPNDevice, error) {
	device, err := d.gvpnDevicesRepo.FindOne(ctx, repos.Filter{
		fc.AccountName:                  ctx.AccountName,
		fc.GlobalVPNDeviceGlobalVPNName: gvpn,
		fc.MetadataName:                 gvpnDevice,
	})
	if err != nil {
		return nil, err
	}

	if device == nil {
		return nil, errors.ErrNotFound{Message: fmt.Sprintf("no global vpn device with name=%s", gvpnDevice)}
	}
	return device, nil
}
