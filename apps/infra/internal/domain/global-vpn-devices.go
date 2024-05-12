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
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
)

func (d *domain) claimNextFreeDeviceIP(ctx InfraContext, deviceName string, gvpnName string) (string, error) {
	for {
		freeIp, err := d.freeDeviceIpRepo.FindOne(ctx, repos.Filter{
			fields.AccountName:           ctx.AccountName,
			fc.FreeDeviceIPGlobalVPNName: gvpnName,
		})
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
	filter := d.gvpnDevicesRepo.MergeMatchFilters(repos.Filter{
		fc.AccountName:                  ctx.AccountName,
		fc.GlobalVPNDeviceGlobalVPNName: gvpn,
	}, search)
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

func (d *domain) buildPeersFromGlobalVPNDevices(ctx InfraContext, gvpn string) (publicPeers []wgv1.Peer, privatePeers []wgv1.Peer, err error) {
	devices, err := d.gvpnDevicesRepo.Find(ctx, repos.Query{
		Filter: map[string]any{
			fc.AccountName:                   ctx.AccountName,
			fc.GlobalVPNDeviceGlobalVPNName:  gvpn,
			fc.GlobalVPNDeviceCreationMethod: map[string]any{"$ne": gvpnConnectionDeviceMethod},
		},
	})
	if err != nil {
		return nil, nil, err
	}

	publicPeers = make([]wgv1.Peer, 0, 10) // 10 is just a random low number
	privatePeers = make([]wgv1.Peer, 0, len(devices))
	for i := range devices {
		if devices[i].PublicEndpoint != nil {
			publicPeers = append(publicPeers, wgv1.Peer{
				PublicKey:  devices[i].PublicKey,
				Endpoint:   *devices[i].PublicEndpoint,
				IP:         devices[i].IPAddr,
				DeviceName: devices[i].Name,
				AllowedIPs: []string{fmt.Sprintf("%s/32", devices[i].IPAddr)},
			})
			continue
		}

		privatePeers = append(privatePeers, wgv1.Peer{
			PublicKey:  devices[i].PublicKey,
			IP:         devices[i].IPAddr,
			DeviceName: devices[i].Name,
			AllowedIPs: []string{fmt.Sprintf("%s/32", devices[i].IPAddr)},
		})
	}

	return publicPeers, privatePeers, nil
}

func (d *domain) GetGlobalVPNDevice(ctx InfraContext, gvpn string, gvpnDevice string) (*entities.GlobalVPNDevice, error) {
	return d.findGlobalVPNDevice(ctx, gvpn, gvpnDevice)
}

func (d *domain) GetGlobalVPNDeviceWgConfig(ctx InfraContext, gvpn string, gvpnDevice string) (string, error) {
	return d.getGlobalVPNDeviceWgConfig(ctx, gvpn, gvpnDevice)
}

func (d *domain) getGlobalVPNDeviceWgConfig(ctx InfraContext, gvpn string, gvpnDevice string) (string, error) {
	device, err := d.findGlobalVPNDevice(ctx, gvpn, gvpnDevice)
	if err != nil {
		return "", err
	}

	gvpnConns, err := d.listGlobalVPNConnections(ctx, gvpn)
	if err != nil {
		return "", err
	}

	gvpnConnPeers, err := d.getGlobalVPNConnectionPeers(gvpnConns)
	if err != nil {
		return "", err
	}

	pubPeers, privPeers, err := d.buildPeersFromGlobalVPNDevices(ctx, gvpn)
	if err != nil {
		return "", err
	}

	pubPeers = append(gvpnConnPeers, pubPeers...)

	publicPeers := make([]wgutils.PublicPeer, 0, len(pubPeers))
	for _, peer := range pubPeers {
		publicPeers = append(publicPeers, wgutils.PublicPeer{
			PublicKey:  peer.PublicKey,
			AllowedIPs: peer.AllowedIPs,
			Endpoint:   peer.Endpoint,
			IPAddr:     peer.IP,
		})
	}

	privatePeers := make([]wgutils.PrivatePeer, 0, len(privPeers))
	for _, peer := range privatePeers {
		privatePeers = append(privatePeers, wgutils.PrivatePeer{
			PublicKey:  peer.PublicKey,
			AllowedIPs: peer.AllowedIPs,
		})
	}

	dnsServer := ""
	for i := range gvpnConns {
		if gvpnConns[i].ParsedWgParams != nil && gvpnConns[i].ParsedWgParams.DNSServer != nil {
			dnsServer = *gvpnConns[i].ParsedWgParams.DNSServer
		}
	}

	if dnsServer == "" {
		return "", errors.Newf("no DNS server found for global VPN device %s", gvpn)
	}

	config, err := wgutils.GenerateWireguardConfig(wgutils.WgConfigParams{
		IPAddr:       device.IPAddr,
		PrivateKey:   device.PrivateKey,
		DNS:          dnsServer,
		PublicPeers:  publicPeers,
		PrivatePeers: privatePeers,
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
		return nil, errors.Newf("no global vpn device (name=%s) found", gvpnDevice)
	}
	return device, nil
}
