package domain

import (
	"context"
	"github.com/kloudlite/api/pkg/errors"
	"time"

	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
)

func (d *domain) ListVPNDevices(ctx context.Context, accountName string, clusterName *string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.VPNDevice], error) {
	filter := repos.Filter{"accountName": accountName}
	if clusterName != nil {
		filter["clusterName"] = *clusterName
	}
	return d.vpnDeviceRepo.FindPaginated(ctx, d.vpnDeviceRepo.MergeMatchFilters(filter, search), pagination)
}

func (d *domain) GetVPNDevice(ctx InfraContext, clusterName string, deviceName string) (*entities.VPNDevice, error) {
	return d.findVPNDevice(ctx, clusterName, deviceName)
}

func (d *domain) CreateVPNDevice(ctx InfraContext, clusterName string, device entities.VPNDevice) (*entities.VPNDevice, error) {
	device.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &device.Device); err != nil {
		return nil, errors.NewE(err)
	}

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
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}
	if err:=d.natCli.Conn.Publish(d.vpnDeviceResUpdateSubject(nDevice), []byte("Deleted")); err != nil {
		d.logger.Errorf(err, "failed to publish message to subject %q", d.vpnDeviceResUpdateSubject(nDevice))
	}

	if err := d.resDispatcher.ApplyToTargetCluster(ctx, clusterName, &nDevice.Device, nDevice.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}
	return nDevice, nil
}

func (d *domain) UpdateVPNDevice(ctx InfraContext, clusterName string, device entities.VPNDevice) (*entities.VPNDevice, error) {
	device.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &device.Device); err != nil {
		return nil, errors.NewE(err)
	}

	currDevice, err := d.findVPNDevice(ctx, clusterName, device.Name)
	if err != nil {
		return nil, errors.NewE(err)
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
		return nil, errors.NewE(err)
	}
	if err:=d.natCli.Conn.Publish(d.vpnDeviceResUpdateSubject(nDevice), []byte("Updated")); err != nil {
		d.logger.Errorf(err, "failed to publish message to subject %q", d.vpnDeviceResUpdateSubject(nDevice))
	}

	if err := d.resDispatcher.ApplyToTargetCluster(ctx, clusterName, &nDevice.Device, nDevice.RecordVersion); err != nil {
		return nil, errors.NewE(err)
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
		return nil, errors.NewE(err)
	}

	if device == nil {
		return nil, errors.Newf("no vpn device with name=%q found", name)
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
		return errors.NewE(err)
	}

	device.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, device.RecordVersion)
	if _, err := d.vpnDeviceRepo.UpdateById(ctx, device.Id, device); err != nil {
		return errors.NewE(err)
	}
	if err:=d.natCli.Conn.Publish(d.vpnDeviceResUpdateSubject(device), []byte("Updated")); err != nil {
		d.logger.Errorf(err, "failed to publish message to subject %q", d.vpnDeviceResUpdateSubject(device))
	}
	return d.resDispatcher.DeleteFromTargetCluster(ctx, clusterName, &device.Device)
}

func (d *domain) OnVPNDeviceApplyError(ctx InfraContext, clusterName string, name string, errMsg string) error {
	currDevice, err := d.findVPNDevice(ctx, clusterName, name)
	if err != nil {
		return errors.NewE(err)
	}

	currDevice.SyncStatus.State = t.SyncStateErroredAtAgent
	currDevice.SyncStatus.LastSyncedAt = time.Now()
	currDevice.SyncStatus.Error = &errMsg

	_, err = d.vpnDeviceRepo.UpdateById(ctx, currDevice.Id, currDevice)
	if err:=d.natCli.Conn.Publish(d.vpnDeviceResUpdateSubject(currDevice), []byte("Updated")); err != nil {
		d.logger.Errorf(err, "failed to publish message to subject %q", d.vpnDeviceResUpdateSubject(currDevice))
	}
	return errors.NewE(err)
}

func (d *domain) OnVPNDeviceUpdateMessage(ctx InfraContext, clusterName string, device entities.VPNDevice) error {
	currDevice, err := d.findVPNDevice(ctx, clusterName, device.Name)
	if err != nil {
		return errors.NewE(err)
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
	if err:=d.natCli.Conn.Publish(d.vpnDeviceResUpdateSubject(currDevice), []byte("Updated")); err != nil {
		d.logger.Errorf(err, "failed to publish message to subject %q", d.vpnDeviceResUpdateSubject(currDevice))
	}
	return errors.NewE(err)
}

func (d *domain) OnVPNDeviceDeleteMessage(ctx InfraContext, clusterName string, device entities.VPNDevice) error {
	currDevice, err := d.findVPNDevice(ctx, clusterName, device.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.matchRecordVersion(device.Annotations, currDevice.RecordVersion); err != nil {
		return errors.NewE(err)
	}

	if err = d.vpnDeviceRepo.DeleteById(ctx, currDevice.Id); err != nil {
		return errors.NewE(err)
	}
	if err:=d.natCli.Conn.Publish(d.vpnDeviceResUpdateSubject(currDevice), []byte("Updated")); err != nil {
		d.logger.Errorf(err, "failed to publish message to subject %q", d.vpnDeviceResUpdateSubject(currDevice))
	}
	return nil
}
