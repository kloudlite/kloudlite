package domain

import (
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"time"
)

func (d *domain) UpdateVpnDeviceNs(ctx InfraContext, clusterName string, devName string, namespace string) error {
	if err := d.canPerformActionInDevice(ctx, iamT.UpdateVPNDevice, devName); err != nil {
		return errors.NewE(err)
	}

	currDevice, err := d.findVPNDevice(ctx, clusterName, devName)
	if err != nil {
		return errors.NewE(err)
	}

	if currDevice.ManagingByDev!=nil {
		return errors.Newf("deviceIn is not self managed, cannot update")
	}

	currDevice.SyncStatus = t.GenSyncStatus(t.SyncActionApply, currDevice.RecordVersion)

	nDevice, err := d.vpnDeviceRepo.PatchById(ctx, currDevice.Id, repos.Document{
		"spec.deviceNamespace":			 namespace,
		"lastUpdatedBy":common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
		"syncStatus.lastSyncedAt": time.Now(),
		"syncStatus.action":        t.SyncActionApply,
		"syncStatus.state":         t.SyncStateInQueue,
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

	filter := repos.Filter{"accountName": accountName}
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

func (d *domain) CreateVPNDevice(ctx InfraContext, clusterName string, device entities.VPNDevice) (*entities.VPNDevice, error) {
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

func (d *domain) UpdateVPNDevice(ctx InfraContext, clusterName string, deviceIn entities.VPNDevice) (*entities.VPNDevice, error) {
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

	if currDevice.ManagingByDev!=nil {
		return nil, errors.Newf("deviceIn is not self managed, cannot update")
	}

	currDevice.SyncStatus = t.GenSyncStatus(t.SyncActionApply, currDevice.RecordVersion)

	nDevice, err := d.vpnDeviceRepo.PatchById(ctx, currDevice.Id, repos.Document{
		"metadata.labels":      deviceIn.Labels,
		"metadata.annotations": deviceIn.Annotations,
		"displayName":		  deviceIn.DisplayName,
		"recordVersion":    currDevice.RecordVersion+1,
		"spec.Ports":			 deviceIn.Spec.Ports,
		"lastUpdatedBy":common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
		"syncStatus.lastSyncedAt": time.Now(),
		"syncStatus.action":        t.SyncActionApply,
		"syncStatus.state":         t.SyncStateInQueue,
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

func (d *domain) UpdateVpnDevicePorts(ctx InfraContext, clusterName string, devName string, ports []*wgv1.Port) error {
	if err := d.canPerformActionInDevice(ctx, iamT.UpdateVPNDevice, devName); err != nil {
		return errors.NewE(err)
	}

	currDevice, err := d.findVPNDevice(ctx, clusterName, devName)
	if err != nil {
		return errors.NewE(err)
	}

	if currDevice.ManagingByDev!=nil {
		return errors.Newf("deviceIn is not self managed, cannot update")
	}

	currDevice.SyncStatus = t.GenSyncStatus(t.SyncActionApply, currDevice.RecordVersion)

	nDevice, err := d.vpnDeviceRepo.PatchById(ctx, currDevice.Id, repos.Document{
		"spec.Ports":			 ports,
		"lastUpdatedBy":common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
		"syncStatus.lastSyncedAt": time.Now(),
		"syncStatus.action":        t.SyncActionApply,
		"syncStatus.state":         t.SyncStateInQueue,
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


func (d *domain) DeleteVPNDevice(ctx InfraContext, clusterName string, name string) error {
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
		"lastUpdatedBy":common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
		"syncStatus.lastSyncedAt": time.Now(),
		"syncStatus.action":        t.SyncActionDelete,
		"syncStatus.state":         t.SyncStateInQueue,
	}); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishVpnDeviceEvent(device, PublishUpdate)
	return d.resDispatcher.DeleteFromTargetCluster(ctx, clusterName, &device.Device)
}

func (d *domain) OnVPNDeviceApplyError(ctx InfraContext, clusterName string, name string, errMsg string, opts UpdateAndDeleteOpts) error {
	currentDevice, err := d.findVPNDevice(ctx, clusterName, name)
	if err != nil {
		return errors.NewE(err)
	}

	_, err = d.vpnDeviceRepo.PatchById(ctx, currentDevice.Id, repos.Document{
		"syncStatus.state":         t.SyncStateErroredAtAgent,
		"syncStatus.lastSyncedAt": opts.MessageTimestamp,
		"syncStatus.error":         &errMsg,
	})
	d.resourceEventPublisher.PublishVpnDeviceEvent(currentDevice, PublishUpdate)
	return errors.NewE(err)
}

func(d *domain) UpsertManagedVPNDevice(ctx InfraContext, clusterName string, deviceIn entities.VPNDevice, managedDeviceId repos.ID) (*entities.VPNDevice, error){
	existingManagingDevice, err := d.vpnDeviceRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"managingByDev": managedDeviceId,
		"spec.disabled": false,
	})
	if existingManagingDevice!= nil {
		existingManagingDevice.Spec.Disabled = true
		_, _ = d.UpdateVPNDevice(ctx, existingManagingDevice.ClusterName, *existingManagingDevice)
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
		device, err = d.CreateVPNDevice(ctx, clusterName, deviceIn)
		if err != nil {
			return nil, errors.NewE(err)
		}
	}
	device.ManagingByDev = &managedDeviceId
	return d.UpdateVPNDevice(ctx, clusterName, *device)
}

func (d *domain)DeleteManagedVPNDevice(ctx InfraContext, managedDeviceId string) error{
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
		if err := d.DeleteVPNDevice(ctx, device.ClusterName, device.Name); err != nil {
			return errors.NewE(err)
		}
	}
	return nil
}

func (d *domain) OnVPNDeviceUpdateMessage(ctx InfraContext, clusterName string, device entities.VPNDevice, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	currentDevice, err := d.findVPNDevice(ctx, clusterName, device.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.matchRecordVersion(device.Annotations, currentDevice.RecordVersion ); err != nil {
		return d.resyncToTargetCluster(ctx, currentDevice.SyncStatus.Action, clusterName, &currentDevice.Device, currentDevice.RecordVersion)
	}

	annVersion, _ := d.parseRecordVersionFromAnnotations(device.Annotations)
	_, err = d.vpnDeviceRepo.PatchById(ctx, currentDevice.Id, repos.Document{
		"metadata.labels":      device.Labels,
		"metadata.annotations": device.Annotations,
		"metadata.generation":  device.Generation,
		"metadata.creationTimestamp":  device.CreationTimestamp,
		"wireguardConfig":	  device.WireguardConfig,
		"status":      device.Status,
		"syncStatus":  t.SyncStatus{
			LastSyncedAt: opts.MessageTimestamp,
			Error: 	  nil,
			Action:       t.SyncActionApply,
			RecordVersion: annVersion,
			State: func() t.SyncState {
				if status == types.ResourceStatusDeleting {
					return t.SyncStateDeletingAtAgent
				}
				return t.SyncStateUpdatedAtAgent
			}(),
		},
	})

	d.resourceEventPublisher.PublishVpnDeviceEvent(currentDevice, PublishUpdate)
	return nil
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
	d.resourceEventPublisher.PublishVpnDeviceEvent(currDevice, PublishUpdate)
	return nil
}
