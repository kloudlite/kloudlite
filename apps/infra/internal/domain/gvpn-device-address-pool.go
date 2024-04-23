package domain

import (
	"fmt"
	"time"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common/fields"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/iputils"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) getNextDeviceAddress(ctx InfraContext, gvpnName string) (string, error) {
	uuid := fn.CleanerNanoidOrDie(40)

	var ipClaim *entities.IPClaim
	var err error

	var addrPool *entities.GlobalVPNDeviceAddressPool
	var freeIp *entities.FreeIP

	fromAddressPool := false
	fromFreeIP := false
	for {
		freeIp, err = d.freeIpRepo.FindOne(ctx, repos.Filter{
			fields.AccountName:     ctx.AccountName,
			fc.FreeIPGlobalVPNName: gvpnName,
		})
		if err != nil {
			return "", err
		}

		ip := ""

		if freeIp != nil {
			fromFreeIP = true
			ip = freeIp.IPAddr
		} else {
			fromFreeIP = false
			fromAddressPool = true

			// now read from addrPool
			addrPool, err = d.deviceAddressPoolRepo.FindOne(ctx, repos.Filter{
				fields.AccountName:                         ctx.AccountName,
				fc.GlobalVPNDeviceAddressPoolGlobalVPNName: gvpnName,
			})
			if err != nil {
				return "", err
			}

			if addrPool == nil {
				return "", fmt.Errorf("address pool not found")
			}

			ip, err = iputils.GetIPAddrInARange(addrPool.CIDR, addrPool.RunningOffset+1, addrPool.MaxOffset)
			if err != nil {
				return "", err
			}
		}

		ipClaim, err = d.ipClaimRepo.Create(ctx, &entities.IPClaim{
			AccountName:    ctx.AccountName,
			GlobalVPNName:  gvpnName,
			IPAddr:         ip,
			ReservationKey: uuid,
		})
		if err != nil {
			d.logger.Warnf("ip addr already claimed (err: %s), retrying again", err.Error())
			<-time.After(50 * time.Millisecond)
			continue
		}
		break
	}

	if fromFreeIP {
		if err := d.freeIpRepo.DeleteById(ctx, freeIp.Id); err != nil {
			return "", err
		}
	}

	if fromAddressPool {
		if _, err := d.deviceAddressPoolRepo.PatchById(ctx, addrPool.Id, repos.Document{"$inc": map[string]any{
			fc.GlobalVPNDeviceAddressPoolRunningOffset: 1,
		}}); err != nil {
			return "", err
		}
	}

	addr := ipClaim.IPAddr
	if err = d.ipClaimRepo.DeleteById(ctx, ipClaim.Id); err != nil {
		return "", err
	}

	return addr, nil
}

// func (d *domain) getNextDeviceAddress2(ctx InfraContext, gvpnName string) (string, error) {
// 	uuid := fn.CleanerNanoidOrDie(40)
// 	var addrPool *entities.GlobalVPNDeviceAddressPool
// 	var err error
// 	for {
// 		freeIp, err := d.freeIpRepo.FindOne(ctx, repos.Filter{
// 			fields.AccountName:     ctx.AccountName,
// 			fc.FreeIPGlobalVPNName: gvpnName,
// 		})
// 		if err != nil {
// 			return "", err
// 		}
//
// 		if freeIp == nil {
// 		}
//
// 		addrPool, err = d.deviceAddressPoolRepo.FindOne(ctx, repos.Filter{
// 			fields.AccountName: ctx.AccountName,
// 			"globalVPNName":    gvpnName,
// 		})
// 		if err != nil {
// 			return "", err
// 		}
//
// 		if addrPool == nil {
// 			return "", fmt.Errorf("address pool not found")
// 		}
//
// 		if len(addrPool.FreeAddressPool) > 0 {
// 			key := ""
// 			for k := range addrPool.FreeAddressPool {
// 				key = k
// 				break
// 			}
//
// 			addrPool.ReservedIPs[uuid] = key
// 			delete(addrPool.FreeAddressPool, key)
// 			if _, err := d.deviceAddressPoolRepo.UpdateWithVersionCheck(ctx, addrPool.Id, addrPool); err != nil {
// 				if errors.Is(err, repos.ErrRecordMismatch) {
// 					continue
// 				}
// 				return "", err
// 			}
// 			break
// 		}
//
// 		ip, err := iputils.GetIPAddrInARange(addrPool.CIDR, addrPool.RunningOffset, addrPool.MaxOffset)
// 		if err != nil {
// 			return "", err
// 		}
//
// 		addrPool.RunningOffset += 1
// 		addrPool.ReservedIPs[uuid] = ip
//
// 		if _, err := d.deviceAddressPoolRepo.UpdateWithVersionCheck(ctx, addrPool.Id, addrPool); err != nil {
// 			if errors.Is(err, repos.ErrRecordMismatch) {
// 				continue
// 			}
// 			return "", err
// 		}
// 		break
// 	}
//
// 	ip := addrPool.ReservedIPs[uuid]
// 	if _, err := d.deviceAddressPoolRepo.PatchById(ctx, addrPool.Id, repos.Document{
// 		"$unset": map[string]any{fmt.Sprintf("%s.%s", fc.GlobalVPNDeviceAddressPoolReservedIPs, uuid): 1},
// 	}); err != nil {
// 		return "", err
// 	}
//
// 	return ip, nil
// }

func (d *domain) addToFreeAddressPool(ctx InfraContext, gvpnName string, ip string) error {
	_, err := d.freeIpRepo.Create(ctx, &entities.FreeIP{
		AccountName:   ctx.AccountName,
		GlobalVPNName: gvpnName,
		IPAddr:        ip,
	})
	return err
}

func (d *domain) findDeviceAddressPool(ctx InfraContext, gvpnName string) (*entities.GlobalVPNDeviceAddressPool, error) {
	addrPool, err := d.deviceAddressPoolRepo.FindOne(ctx, repos.Filter{
		fields.AccountName: ctx.AccountName,
		"globalVPNName":    gvpnName,
	})
	if err != nil {
		return nil, err
	}

	return addrPool, nil
}

func (d *domain) createDeviceAddressPool(ctx InfraContext, pool entities.GlobalVPNDeviceAddressPool) (*entities.GlobalVPNDeviceAddressPool, error) {
	return d.deviceAddressPoolRepo.Create(ctx, &pool)
}
