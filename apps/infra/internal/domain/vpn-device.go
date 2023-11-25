package domain

import (
	"context"
	"fmt"
	"time"

	"kloudlite.io/apps/infra/internal/entities"
	"kloudlite.io/common"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

func (d *domain) ListVPNDevices(ctx context.Context, accountName string, clusterName *string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.VPNDevice], error) {
	filter := repos.Filter{
		"accountName": accountName,
	}
	if clusterName != nil {
		filter["clusterName"] = *clusterName
	}

	return d.vpnDeviceRepo.FindPaginated(ctx, d.vpnDeviceRepo.MergeMatchFilters(filter, search), pagination)
}

func (d *domain) GetVPNDevice(ctx InfraContext, clusterName string, deviceName string) (*entities.VPNDevice, error) {
	return d.findVPNDevice(ctx, clusterName, deviceName)
}

func (d *domain) validateVPNDeviceLimits(ctx context.Context, accountName string, clusterName string) error {
	count, err := d.vpnDeviceRepo.Count(ctx, repos.Filter{
		"accountName": accountName,
		"clusterName": clusterName,
	})
	if err != nil {
		return err
	}

	if count > d.env.VPNDevicesMaxOffset {
		return fmt.Errorf("max vpn devices limit reached (max. %d devices)", d.env.VPNDevicesMaxOffset)
	}

	return nil
}

// func (d *domain) lookupNextOffset(ctx context.Context, accountName string, clusterName string) (int, error) {
// 	vpnDevices, err := d.vpnDeviceRepo.Find(ctx, repos.Query{Filter: repos.Filter{
// 		"accountName": accountName,
// 		"clusterName": clusterName,
// 	}, Sort: map[string]any{"spec.offset": 1}})
// 	if err != nil {
// 		return 0, err
// 	}
//
// 	for i, j := d.env.VPNDevicesOffsetStart, 0; i <= int(d.env.VPNDevicesMaxOffset); i, j = i+1, j+1 {
// 		if j >= len(vpnDevices) {
// 			return i, nil
// 		}
//
// 		if vpnDevices[j].Spec.Offset != nil && *vpnDevices[j].Spec.Offset != i {
// 			return i, nil
// 		}
// 	}
//
// 	return 0, fmt.Errorf("max vpn devices limit reached (max. %d devices)", d.env.VPNDevicesMaxOffset)
// }

func (d *domain) CreateVPNDevice(ctx InfraContext, clusterName string, device entities.VPNDevice) (*entities.VPNDevice, error) {
	device.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &device.Device); err != nil {
		return nil, err
	}

	if err := d.validateVPNDeviceLimits(ctx, ctx.AccountName, clusterName); err != nil {
		return nil, err
	}

	// offset, err := d.lookupNextOffset(ctx, ctx.AccountName, clusterName)
	// if err != nil {
	// 	return nil, err
	// }
	//
	// device.Spec.Offset = fn.New(offset)

	device.IncrementRecordVersion()
	device.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	device.LastUpdatedBy = device.CreatedBy

	device.AccountName = ctx.AccountName
	device.ClusterName = clusterName
	device.SyncStatus = t.GenSyncStatus(t.SyncActionApply, device.RecordVersion)

	nDevice, err := d.vpnDeviceRepo.Create(ctx, &device)
	if err != nil {
		if d.vpnDeviceRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, err
		}
		return nil, err
	}

	if err := d.applyToTargetCluster(ctx, clusterName, &nDevice.Device, nDevice.RecordVersion); err != nil {
		return nil, err
	}
	return nDevice, nil
}

func (d *domain) UpdateVPNDevice(ctx InfraContext, clusterName string, device entities.VPNDevice) (*entities.VPNDevice, error) {
	device.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &device.Device); err != nil {
		return nil, err
	}

	currDevice, err := d.findVPNDevice(ctx, clusterName, device.Name)
	if err != nil {
		return nil, err
	}

	currDevice.IncrementRecordVersion()
	currDevice.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	currDevice.DisplayName = device.DisplayName

	currDevice.Labels = device.Labels
	currDevice.Annotations = device.Annotations

	currDevice.Spec.Ports = device.Spec.Ports

	currDevice.SyncStatus = t.GenSyncStatus(t.SyncActionApply, currDevice.RecordVersion)

	nDevice, err := d.vpnDeviceRepo.UpdateById(ctx, device.Id, &device)
	if err != nil {
		return nil, err
	}

	if err := d.applyToTargetCluster(ctx, clusterName, &nDevice.Device, nDevice.RecordVersion); err != nil {
		return nil, err
	}
	return nDevice, nil
}

func (d *domain) findVPNDevice(ctx InfraContext, clusterName string, name string) (*entities.VPNDevice, error) {
	device, err := d.vpnDeviceRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"clusterName":   clusterName,
		"metadata.name": name,
	})
	if err != nil {
		return nil, err
	}

	if device == nil {
		return nil, fmt.Errorf("no vpn device with name=%q found", name)
	}

	return device, nil
}

func (d *domain) GetWgConfigForDevice(ctx InfraContext, clusterName string, deviceName string) (*string, error) {
	// TOOD (nxtcoder17): go to the target cluster, and fetch that secret
	panic("not implemented yet")
}

func (d *domain) DeleteVPNDevice(ctx InfraContext, clusterName string, name string) error {
	device, err := d.findVPNDevice(ctx, clusterName, name)
	if err != nil {
		return err
	}

	device.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, device.RecordVersion)
	if _, err := d.vpnDeviceRepo.UpdateById(ctx, device.Id, device); err != nil {
		return err
	}
	return d.deleteFromTargetCluster(ctx, clusterName, &device.Device)
}

func (d *domain) OnVPNDeviceApplyError(ctx InfraContext, clusterName string, errMsg string, name string) error {
	currDevice, err := d.findVPNDevice(ctx, clusterName, name)
	if err != nil {
		return err
	}

	currDevice.SyncStatus.State = t.SyncStateErroredAtAgent
	currDevice.SyncStatus.LastSyncedAt = time.Now()
	currDevice.SyncStatus.Error = &errMsg

	_, err = d.vpnDeviceRepo.UpdateById(ctx, currDevice.Id, currDevice)
	return err
}

func (d *domain) OnVPNDeviceUpdateMessage(ctx InfraContext, clusterName string, device entities.VPNDevice) error {
	currDevice, err := d.findVPNDevice(ctx, clusterName, device.Name)
	if err != nil {
		return err
	}

	if err := d.matchRecordVersion(device.Annotations, currDevice.RecordVersion); err != nil {
		return d.resyncToTargetCluster(ctx, currDevice.SyncStatus.Action, clusterName, &currDevice.Device, currDevice.RecordVersion)
	}

	currDevice.CreationTimestamp = device.CreationTimestamp
	currDevice.Labels = device.Labels
	currDevice.Annotations = device.Annotations
	currDevice.Generation = device.Generation

	currDevice.Status = device.Status

	currDevice.WireguardConfig = device.WireguardConfig

	currDevice.SyncStatus.State = t.SyncStateReceivedUpdateFromAgent
	currDevice.SyncStatus.RecordVersion = currDevice.RecordVersion
	currDevice.SyncStatus.Error = nil
	currDevice.SyncStatus.LastSyncedAt = time.Now()

	_, err = d.vpnDeviceRepo.UpdateById(ctx, currDevice.Id, currDevice)
	return err
}

func (d *domain) OnVPNDeviceDeleteMessage(ctx InfraContext, clusterName string, device entities.VPNDevice) error {
	currDevice, err := d.findVPNDevice(ctx, clusterName, device.Name)
	if err != nil {
		return err
	}

	if err := d.matchRecordVersion(device.Annotations, currDevice.RecordVersion); err != nil {
		return err
	}

	return d.vpnDeviceRepo.DeleteById(ctx, currDevice.Id)
}
