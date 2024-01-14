package domain

import (
	"time"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
)

func (d *domain) updateVPNDevice(ctx InfraContext, clusterName string, deviceIn entities.VPNDevice, ensureSelfManaged bool) (*entities.VPNDevice, error) {
	if err := d.canPerformActionInDevice(ctx, iamT.UpdateVPNDevice, deviceIn.Name); err != nil {
		return nil, errors.NewE(err)
	}

	deviceIn.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &deviceIn.Device); err != nil {
		return nil, errors.NewE(err)
	}

	currDevice, err := d.findVPNDevice(ctx, clusterName, deviceIn.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if currDevice.ManagingByDev != nil && ensureSelfManaged {
		return nil, errors.Newf("device is not self managed, cannot update")
	}

	currDevice.SyncStatus = t.GenSyncStatus(t.SyncActionApply, currDevice.RecordVersion)

	nDevice, err := d.vpnDeviceRepo.PatchById(ctx, currDevice.Id, repos.Document{
		"metadata.labels":      deviceIn.Labels,
		"metadata.annotations": deviceIn.Annotations,
		"displayName":          deviceIn.DisplayName,
		"recordVersion":        currDevice.RecordVersion + 1,
		"spec.Ports":           deviceIn.Spec.Ports,
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
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishVpnDeviceEvent(nDevice, PublishUpdate)

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

func (d *domain) deleteVPNDevice(ctx InfraContext, clusterName string, name string) error {
	if err := d.canPerformActionInDevice(ctx, iamT.UpdateVPNDevice, name); err != nil {
		return errors.NewE(err)
	}

	device, err := d.findVPNDevice(ctx, clusterName, name)
	if err != nil {
		return errors.NewE(err)
	}
	if device.IsMarkedForDeletion() {
		return errors.Newf("vpnDevice %q (clusterName=%q) is already marked for deletion", name, clusterName)
	}

	if _, err := d.vpnDeviceRepo.PatchById(ctx, device.Id, repos.Document{
		"markedForDeletion": fn.New(true),
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
		"syncStatus.lastSyncedAt": time.Now(),
		"syncStatus.action":       t.SyncActionDelete,
		"syncStatus.state":        t.SyncStateInQueue,
	}); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishVpnDeviceEvent(device, PublishUpdate)
	return d.resDispatcher.DeleteFromTargetCluster(ctx, clusterName, &device.Device)

}

func (d *domain) createVPNDevice(ctx InfraContext, clusterName string, device entities.VPNDevice) (*entities.VPNDevice, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateVPNDevice); err != nil {
		return nil, errors.NewE(err)
	}

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

	if _, err := d.iamClient.AddMembership(ctx, &iam.AddMembershipIn{
		UserId:       string(ctx.UserId),
		ResourceType: string(iamT.ResourceVPNDevice),
		ResourceRef:  iamT.NewResourceRef(ctx.AccountName, iamT.ResourceVPNDevice, device.Name),
		Role:         string(iamT.RoleResourceOwner),
	}); err != nil {
		return nil, errors.NewE(err)
	}

	nDevice, err := d.vpnDeviceRepo.Create(ctx, &device)
	if err != nil {
		if d.vpnDeviceRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishVpnDeviceEvent(&device, PublishAdd)

	if err := d.resDispatcher.ApplyToTargetCluster(ctx, clusterName, &nDevice.Device, nDevice.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}
	return nDevice, nil
}
