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

			ip, err := iputils.GetIPAddrInARange(gvpn.CIDR, gvpn.NumAllocatedDevices+1, gvpn.NumReservedIPsForNonClusterUse)
			if err != nil {
				return "", err
			}

			if _, err := d.freeDeviceIpRepo.Create(ctx, &entities.FreeDeviceIP{
				AccountName:   ctx.AccountName,
				GlobalVPNName: gvpnName,
				IPAddr:        ip,
			}); err != nil {
				continue
			}

			if _, err := d.gvpnRepo.PatchById(ctx, gvpn.Id, repos.Document{"$inc": map[string]any{fc.GlobalVPNNumAllocatedDevices: 1}}); err != nil {
				continue
			}

			continue
		}

		ipAddr := freeIp.IPAddr

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

			d.logger.Warnf("ip addr already claimed (err: %s), retrying again", err.Error())
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

	if err := d.reconGlobalVPNConnections(ctx, device.GlobalVPNName); err != nil {
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

	gv, err := d.gvpnDevicesRepo.Create(ctx, &gvpnDevice)
	if err != nil {
		return nil, err
	}

	if err := d.reconGlobalVPNConnections(ctx, gvpnDevice.GlobalVPNName); err != nil {
		return nil, err
	}

	return gv, nil
}

func (d *domain) buildPeerFromGlobalVPNDevice(ctx InfraContext, gvpn *entities.GlobalVPN, device *entities.GlobalVPNDevice) *networkingv1.Peer {
	allowedIPs := []string{fmt.Sprintf("%s/32", device.IPAddr)}

	if device.CreationMethod == kloudliteGlobalVPNDevice {
		allowedIPs = append(allowedIPs, gvpn.NonClusterUseAllowedIPs...)
	}

	return &networkingv1.Peer{
		DisplayName:    fmt.Sprintf("device/%s/%s", device.GlobalVPNName, device.Name),
		PublicKey:      device.PublicKey,
		PublicEndpoint: device.PublicEndpoint,
		IP:             device.IPAddr,
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
			if devices[i].CreationMethod == kloudliteGlobalVPNDevice {
				allowedIPs = append(allowedIPs, gv.NonClusterUseAllowedIPs...)
			}

			publicPeers = append(publicPeers, networkingv1.Peer{
				DisplayName:    fmt.Sprintf("device/%s/%s", devices[i].GlobalVPNName, devices[i].Name),
				PublicKey:      devices[i].PublicKey,
				PublicEndpoint: devices[i].PublicEndpoint,
				IP:             devices[i].IPAddr,
				DNSSuffix:      nil,
				AllowedIPs:     allowedIPs,
			})
			continue
		}

		privatePeers = append(privatePeers, networkingv1.Peer{
			DisplayName: fmt.Sprintf("device/%s/%s", devices[i].GlobalVPNName, devices[i].Name),
			PublicKey:   devices[i].PublicKey,
			IP:          devices[i].IPAddr,
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

func (d *domain) buildGlobalVPNDeviceWgBaseParams(ctx InfraContext, gvpnConns []*entities.GlobalVPNConnection, gvpnDevice *entities.GlobalVPNDevice) (*wgutils.WgConfigParams, error) {
	gvpnConnPeers, err := d.getGlobalVPNConnectionPeers(ctx, gvpnConns)
	if err != nil {
		return nil, err
	}

	pubPeers, privPeers, err := d.buildPeersFromGlobalVPNDevices(ctx, gvpnDevice.GlobalVPNName)
	if err != nil {
		return nil, err
	}

	pubPeers = append(gvpnConnPeers, pubPeers...)

	publicPeers := make([]wgutils.PublicPeer, 0, len(pubPeers))
	for _, peer := range pubPeers {
		if peer.DisplayName == fmt.Sprintf("device/%s/%s", gvpnDevice.GlobalVPNName, gvpnDevice.Name) || peer.DisplayName == fmt.Sprintf("gateway/%s/%s", gvpnDevice.GlobalVPNName, gvpnDevice.Name) {
			continue
		}
		if peer.PublicEndpoint == nil {
			continue
		}
		publicPeers = append(publicPeers, wgutils.PublicPeer{
			DisplayName: peer.DisplayName,
			PublicKey:   peer.PublicKey,
			AllowedIPs:  peer.AllowedIPs,
			Endpoint:    *peer.PublicEndpoint,
		})
	}

	privatePeers := make([]wgutils.PrivatePeer, 0, len(privPeers))
	for _, peer := range privPeers {
		if peer.DisplayName == fmt.Sprintf("%s/%s", gvpnDevice.GlobalVPNName, gvpnDevice.Name) {
			continue
		}
		privatePeers = append(privatePeers, wgutils.PrivatePeer{
			DisplayName: peer.DisplayName,
			PublicKey:   peer.PublicKey,
			AllowedIPs:  peer.AllowedIPs,
		})
	}

	return &wgutils.WgConfigParams{
		IPAddr:       gvpnDevice.IPAddr,
		PrivateKey:   gvpnDevice.PrivateKey,
		PublicPeers:  publicPeers,
		PrivatePeers: privatePeers,
	}, nil
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

	klDevice, err := d.findGlobalVPNDevice(ctx, gvpn, gv.KloudliteDevice.Name)
	if err != nil {
		return "", err
	}

	if klDevice.PublicEndpoint == nil {
		return "", errors.New("kloudlite device public endpoint is nil, please wait for some time")
	}

	config, err := wgutils.GenerateWireguardConfig(wgutils.WgConfigParams{
		IPAddr:     device.IPAddr,
		PrivateKey: device.PrivateKey,
		DNS:        gv.KloudliteDevice.IPAddr,
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
