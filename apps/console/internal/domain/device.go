package domain

import (
	"context"
	"fmt"
	"github.com/seancfoley/ipaddress-go/ipaddr"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"kloudlite.io/apps/console/internal/domain/entities"
	internal_crds "kloudlite.io/apps/console/internal/domain/op-crds/internal-crds"
	"kloudlite.io/pkg/repos"
)

func (d *domain) GetDevice(ctx context.Context, id repos.ID) (*entities.Device, error) {
	return d.deviceRepo.FindById(ctx, id)
}
func (d *domain) ListAccountDevices(ctx context.Context, accountId repos.ID) ([]*entities.Device, error) {
	q := make(repos.Filter)
	q["account_id"] = accountId
	return d.deviceRepo.Find(ctx, repos.Query{
		Filter: q,
	})
}
func (d *domain) ListUserDevices(ctx context.Context, userId repos.ID) ([]*entities.Device, error) {
	q := make(repos.Filter)
	q["user_id"] = userId
	return d.deviceRepo.Find(ctx, repos.Query{
		Filter: q,
	})
}

func (d *domain) GetDeviceConfig(ctx context.Context, deviceId repos.ID) (string, error) {
	device, err := d.deviceRepo.FindById(ctx, deviceId)
	if err != nil {
		return "", err
	}
	wgAccount, err := d.wgAccountRepo.FindOne(ctx, repos.Filter{
		"account_id": device.AccountId,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`
[Interface]
PrivateKey = %v
Address = %v/32
DNS = 10.43.0.10

[Peer]
PublicKey = %v
AllowedIPs = 10.42.0.0/16, 10.43.0.0/16, 10.13.13.0/24
Endpoint = %v:%v
`, *device.PrivateKey, device.Ip, wgAccount.WgPubKey, wgAccount.AccessDomain, wgAccount.WgPort), nil
}

func (d *domain) AddDevice(ctx context.Context, deviceName string, accountId repos.ID, userId repos.ID) (*entities.Device, error) {
	pk, e := wgtypes.GeneratePrivateKey()
	if e != nil {
		return nil, fmt.Errorf("unable to generate private key because %v", e)
	}
	e = d.ensureWgAccount(ctx, accountId)
	if e != nil {
		return nil, fmt.Errorf("unable to ensure wg account because %v", e)
	}
	devices, err := d.deviceRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"account_id": accountId,
		},
		Sort: map[string]any{
			"index": 1,
		},
	})
	if err != nil {
		return nil, err
	}
	index := -1
	count := 0
	for i, d := range devices {
		count++
		if d.Index != i {
			index = i
			break
		}
	}
	if index == -1 {
		index = count
	}
	deviceIp, e := getRemoteDeviceIp(int64(index + 2))
	ip := deviceIp.String()
	pkString := pk.String()
	pbKeyString := pk.PublicKey().String()
	device, e := d.deviceRepo.Create(ctx, &entities.Device{
		Name:       deviceName,
		AccountId:  accountId,
		UserId:     userId,
		PrivateKey: &pkString,
		PublicKey:  &pbKeyString,
		Ip:         ip,
		Status:     entities.DeviceStateSyncing,
		Index:      index,
	})
	if e != nil {
		return nil, fmt.Errorf("unable to persist in db %v", e)
	}
	err = d.workloadMessenger.SendAction("apply", string(device.Id), &internal_crds.Device{
		APIVersion: internal_crds.DeviceAPIVersion,
		Kind:       internal_crds.DeviceKind,
		Metadata: internal_crds.DeviceMetadata{
			Name: string(device.Id),
			Annotations: map[string]string{
				"kloudlite.io/account-ref": string(device.Id),
			},
		},
		Spec: internal_crds.DeviceSpec{
			Account:      string(device.AccountId),
			ActiveRegion: device.ActiveRegion,
			Offset:       device.Index,
			DeviceId:     string(device.Id),
			Ports:        device.ExposedPorts,
		},
	})
	if err != nil {
		return device, err
	}
	return device, e
}
func (d *domain) RemoveDevice(ctx context.Context, deviceId repos.ID) error {
	device, err := d.deviceRepo.FindById(ctx, deviceId)
	if err != nil {
		return err
	}
	device.Status = entities.DeviceStateSyncing
	_, err = d.deviceRepo.UpdateById(ctx, deviceId, device)
	if err != nil {
		return err
	}
	err = d.workloadMessenger.SendAction("delete", string(device.Id), &internal_crds.Device{
		APIVersion: internal_crds.DeviceAPIVersion,
		Kind:       internal_crds.DeviceKind,
		Metadata: internal_crds.DeviceMetadata{
			Name: string(device.Id),
		},
	})
	return err
}
func (d *domain) UpdateDevice(ctx context.Context, deviceId repos.ID, region string, ports []int32) (done bool, e error) {
	device, e := d.deviceRepo.FindById(ctx, deviceId)
	device.ActiveRegion = region
	device.ExposedPorts = ports
	_, err := d.deviceRepo.UpdateById(ctx, deviceId, device)
	if err != nil {
		return false, e
	}
	err = d.workloadMessenger.SendAction("apply", string(device.Id), &internal_crds.Device{
		APIVersion: internal_crds.DeviceAPIVersion,
		Kind:       internal_crds.DeviceKind,
		Metadata: internal_crds.DeviceMetadata{
			Name: string(device.Id),
			Annotations: map[string]string{
				"kloudlite.io/account-ref": string(device.Id),
			},
		},
		Spec: internal_crds.DeviceSpec{
			Account:      string(device.AccountId),
			ActiveRegion: device.ActiveRegion,
			Offset:       device.Index,
			DeviceId:     string(device.Id),
			Ports:        device.ExposedPorts,
		},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func getRemoteDeviceIp(deviceOffset int64) (*ipaddr.IPAddressString, error) {
	deviceRange := ipaddr.NewIPAddressString("10.13.0.0/16")

	if address, addressError := deviceRange.ToAddress(); addressError == nil {
		increment := address.Increment(deviceOffset + 2)
		return ipaddr.NewIPAddressString(increment.GetNetIP().String()), nil
	} else {
		return nil, addressError
	}
}

func (d *domain) ensureWgAccount(ctx context.Context, accountId repos.ID) error {
	one, err := d.wgAccountRepo.FindOne(ctx, repos.Filter{
		"account_id": accountId,
	})
	if err != nil {
		return err
	}
	if one == nil {
		pk, e := wgtypes.GeneratePrivateKey()
		if e != nil {
			return e
		}
		pkString := pk.String()
		pbKeyString := pk.PublicKey().String()
		_, err = d.wgAccountRepo.Create(ctx, &entities.WGAccount{
			AccountID:    accountId,
			WgPubKey:     pbKeyString,
			WgPrivateKey: pkString,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
