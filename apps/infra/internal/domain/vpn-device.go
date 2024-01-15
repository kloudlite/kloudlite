package domain

import (
	"time"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) UpdateVpnDeviceNs(ctx InfraContext, clusterName string, devName string, namespace string) error {
	if err := d.canPerformActionInDevice(ctx, iamT.UpdateVPNDevice, devName); err != nil {
		return errors.NewE(err)
	}

	currDevice, err := d.findVPNDevice(ctx, clusterName, devName)
	if err != nil {
		return errors.NewE(err)
	}

	if currDevice.ManagingByDev != nil {
		return errors.Newf("device is not self managed, cannot update")
	}

	currDevice.SyncStatus = t.GenSyncStatus(t.SyncActionApply, currDevice.RecordVersion)

	nDevice, err := d.vpnDeviceRepo.PatchById(ctx, currDevice.Id, repos.Document{
		"spec.deviceNamespace": namespace,
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
		"syncStatus.lastSyncedAt": time.Now(),
		"syncStatus.action":       t.SyncActionApply,
		"syncStatus.state":        t.SyncStateInQueue,
	})
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishVpnDeviceEvent(nDevice, PublishUpdate)

	if err := d.resDispatcher.ApplyToTargetCluster(ctx, clusterName, &nDevice.Device, nDevice.RecordVersion); err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) ListVPNDevices(ctx InfraContext, accountName string, clusterName *string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.VPNDevice], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateVPNDevice); err != nil {
		return nil, errors.NewE(err)
	}

	filter := repos.Filter{"accountName": accountName, "managingByDev": nil}
	if clusterName != nil {
		filter["clusterName"] = *clusterName
	}
	return d.vpnDeviceRepo.FindPaginated(ctx, d.vpnDeviceRepo.MergeMatchFilters(filter, search), pagination)
}

func (d *domain) GetVPNDevice(ctx InfraContext, clusterName string, deviceName string) (*entities.VPNDevice, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetVPNDevice); err != nil {
		return nil, errors.NewE(err)
	}
	return d.findVPNDevice(ctx, clusterName, deviceName)
}

func (d *domain) UpdateVPNDevice(ctx InfraContext, clusterName string, deviceIn entities.VPNDevice) (*entities.VPNDevice, error) {
	return d.updateVPNDevice(ctx, clusterName, deviceIn, true)
}

func (d *domain) UpdateVpnDevicePorts(ctx InfraContext, clusterName string, devName string, ports []*wgv1.Port) error {
	if err := d.canPerformActionInDevice(ctx, iamT.UpdateVPNDevice, devName); err != nil {
		return errors.NewE(err)
	}

	currDevice, err := d.findVPNDevice(ctx, clusterName, devName)
	if err != nil {
		return errors.NewE(err)
	}

	if currDevice.ManagingByDev != nil {
		return errors.Newf("device is not self managed, cannot update")
	}

	currDevice.SyncStatus = t.GenSyncStatus(t.SyncActionApply, currDevice.RecordVersion)

	nDevice, err := d.vpnDeviceRepo.PatchById(ctx, currDevice.Id, repos.Document{
		"spec.ports": ports,
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
		"syncStatus.lastSyncedAt": time.Now(),

		"syncStatus.action": t.SyncActionApply,
		"syncStatus.state":  t.SyncStateInQueue,
	})
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishVpnDeviceEvent(nDevice, PublishUpdate)

	if err := d.resDispatcher.ApplyToTargetCluster(ctx, clusterName, &nDevice.Device, nDevice.RecordVersion); err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) UpsertManagedVPNDevice(ctx InfraContext, clusterName string, deviceIn entities.VPNDevice, managedDeviceId repos.ID) (*entities.VPNDevice, error) {
	if managedDeviceId == "" {
		return nil, errors.Newf("managedDeviceId cannot be empty")
	}

	// checking if any device is already running
	existingManagingDevice, err := d.vpnDeviceRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"managingByDev": managedDeviceId,
		"spec.disabled": false,
	})

	if existingManagingDevice != nil {
		existingManagingDevice.Spec.Disabled = true

		// if any device already running disable it
		_, _ = d.updateVPNDevice(ctx, existingManagingDevice.ClusterName, *existingManagingDevice, false)
	}

	device, err := d.vpnDeviceRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"clusterName":   clusterName,
		"metadata.name": deviceIn.Name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if device == nil {
		deviceIn.ManagingByDev = &managedDeviceId
		device, err = d.createVPNDevice(ctx, clusterName, deviceIn)
		if err != nil {
			return nil, errors.NewE(err)
		}

		return device, nil
	}

	device.ManagingByDev = &managedDeviceId
	device.Spec = deviceIn.Spec
	return d.updateVPNDevice(ctx, clusterName, *device, false)
}

func (d *domain) DeleteManagedVPNDevice(ctx InfraContext, managedDeviceId string) error {
	devices, err := d.vpnDeviceRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"accountName":   ctx.AccountName,
			"managingByDev": repos.ID(managedDeviceId),
		},
	})
	if err != nil {
		return errors.NewE(err)
	}
	for _, device := range devices {
		if err := d.deleteVPNDevice(ctx, device.ClusterName, device.Name); err != nil {
			return errors.NewE(err)
		}
	}
	return nil
}

func (d *domain) DeleteVPNDevice(ctx InfraContext, clusterName string, name string) error {
	return d.deleteVPNDevice(ctx, clusterName, name)
}

func (d *domain) OnVPNDeviceApplyError(ctx InfraContext, clusterName string, name string, errMsg string, opts UpdateAndDeleteOpts) error {
	currentDevice, err := d.findVPNDevice(ctx, clusterName, name)
	if err != nil {
		return errors.NewE(err)
	}

	_, err = d.vpnDeviceRepo.PatchById(ctx, currentDevice.Id, repos.Document{
		"syncStatus.state":        t.SyncStateErroredAtAgent,
		"syncStatus.lastSyncedAt": opts.MessageTimestamp,
		"syncStatus.error":        &errMsg,
	})
	d.resourceEventPublisher.PublishVpnDeviceEvent(currentDevice, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) CreateVPNDevice(ctx InfraContext, clusterName string, device entities.VPNDevice) (*entities.VPNDevice, error) {
	return d.createVPNDevice(ctx, clusterName, device)
}

func (d *domain) OnVPNDeviceUpdateMessage(ctx InfraContext, clusterName string, device entities.VPNDevice, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	currentDevice, err := d.findVPNDevice(ctx, clusterName, device.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.matchRecordVersion(device.Annotations, currentDevice.RecordVersion); err != nil {
		return d.resyncToTargetCluster(ctx, currentDevice.SyncStatus.Action, clusterName, &currentDevice.Device, currentDevice.RecordVersion)
	}

	if _, err = d.vpnDeviceRepo.PatchById(ctx, currentDevice.Id, repos.Document{
		"metadata.labels":            device.Labels,
		"metadata.annotations":       device.Annotations,
		"metadata.generation":        device.Generation,
		"metadata.creationTimestamp": device.CreationTimestamp,
		"wireguardConfig":            device.WireguardConfig,
		"status":                     device.Status,
		"syncStatus": t.SyncStatus{
			LastSyncedAt:  opts.MessageTimestamp,
			Error:         nil,
			Action:        t.SyncActionApply,
			RecordVersion: currentDevice.RecordVersion,
			State: func() t.SyncState {
				if status == types.ResourceStatusDeleting {
					return t.SyncStateDeletingAtAgent
				}
				return t.SyncStateUpdatedAtAgent
			}(),
		},
	}); err != nil {
		return err
	}

	d.resourceEventPublisher.PublishVpnDeviceEvent(currentDevice, PublishUpdate)
	return nil
}

func (d *domain) OnVPNDeviceDeleteMessage(ctx InfraContext, clusterName string, device entities.VPNDevice) error {
	currDevice, err := d.findVPNDevice(ctx, clusterName, device.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err = d.vpnDeviceRepo.DeleteById(ctx, currDevice.Id); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishVpnDeviceEvent(currDevice, PublishUpdate)
	return nil
}
