package domain

import (
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) findVPNDevice(ctx InfraContext, clusterName string, name string) (*entities.VPNDevice, error) {
	device, err := d.vpnDeviceRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
		fields.MetadataName: name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if device == nil {
		return nil, errors.Newf("no vpn device with name=%q found", name)
	}

	return device, nil
}

func (d *domain) UpdateVpnDeviceNs(ctx InfraContext, clusterName string, devName string, namespace string) error {
	if err := d.canPerformActionInDevice(ctx, iamT.UpdateVPNDevice, devName); err != nil {
		return errors.NewE(err)
	}

	xDevice, err := d.findVPNDevice(ctx, clusterName, devName)
	if err != nil {
		return errors.NewE(err)
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		xDevice,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.VPNDeviceSpecActiveNamespace: namespace,
			},
		})

	upDevice, err := d.vpnDeviceRepo.PatchById(
		ctx,
		xDevice.Id,
		patchForUpdate,
	)

	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeVpnDevice, upDevice.Name, PublishUpdate)

	if err := d.resDispatcher.ApplyToTargetCluster(ctx, clusterName, &upDevice.Device, upDevice.RecordVersion); err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) ListVPNDevices(ctx InfraContext, accountName string, clusterName *string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.VPNDevice], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateVPNDevice); err != nil {
		return nil, errors.NewE(err)
	}

	filter := repos.Filter{fields.AccountName: accountName}
	if clusterName != nil {
		filter[fields.ClusterName] = *clusterName
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
	if err := d.canPerformActionInDevice(ctx, iamT.UpdateVPNDevice, deviceIn.Name); err != nil {
		return nil, errors.NewE(err)
	}

	deviceIn.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &deviceIn.Device); err != nil {
		return nil, errors.NewE(err)
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&deviceIn,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.VPNDeviceSpec:                deviceIn.Spec,
				fc.VPNDeviceSpecPorts:           deviceIn.Spec.Ports,
				fc.VPNDeviceSpecActiveNamespace: deviceIn.Spec.ActiveNamespace,
			},
		})

	upDevice, err := d.vpnDeviceRepo.Patch(
		ctx,
		repos.Filter{
			fields.ClusterName:  clusterName,
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: deviceIn.Name,
		},
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeVpnDevice, upDevice.Name, PublishUpdate)

	if err := d.resDispatcher.ApplyToTargetCluster(ctx, clusterName, &upDevice.Device, upDevice.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}
	return upDevice, nil
}

func (d *domain) UpdateVpnDevicePorts(ctx InfraContext, clusterName string, devName string, ports []*wgv1.Port) error {
	if err := d.canPerformActionInDevice(ctx, iamT.UpdateVPNDevice, devName); err != nil {
		return errors.NewE(err)
	}

	xdevice, err := d.findVPNDevice(ctx, clusterName, devName)
	if err != nil {
		return errors.NewE(err)
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		xdevice,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.VPNDeviceSpecPorts: ports,
			},
		})

	upDevice, err := d.vpnDeviceRepo.Patch(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.ClusterName:  clusterName,
		fields.MetadataName: devName,
	}, patchForUpdate)

	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeVpnDevice, devName, PublishUpdate)

	if err := d.resDispatcher.ApplyToTargetCluster(ctx, clusterName, &upDevice.Device, upDevice.RecordVersion); err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) DeleteVPNDevice(ctx InfraContext, clusterName string, name string) error {
	if err := d.canPerformActionInDevice(ctx, iamT.UpdateVPNDevice, name); err != nil {
		return errors.NewE(err)
	}

	upDevice, err := d.vpnDeviceRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.ClusterName:  clusterName,
			fields.MetadataName: name,
		},
		common.PatchForMarkDeletion(),
	)

	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeVpnDevice, name, PublishUpdate)
	return d.resDispatcher.DeleteFromTargetCluster(ctx, clusterName, &upDevice.Device)
}

func (d *domain) OnVPNDeviceApplyError(ctx InfraContext, clusterName string, name string, errMsg string, opts UpdateAndDeleteOpts) error {
	udevice, err := d.vpnDeviceRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.ClusterName:  clusterName,
			fields.MetadataName: name,
		},
		common.PatchForErrorFromAgent(
			errMsg,
			common.PatchOpts{
				MessageTimestamp: opts.MessageTimestamp,
			},
		),
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeVpnDevice, udevice.Name, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) CreateVPNDevice(ctx InfraContext, clusterName string, device entities.VPNDevice) (*entities.VPNDevice, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateVPNDevice); err != nil {
		return nil, errors.NewE(err)
	}

	device.Namespace = d.env.DeviceNamespace
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
	device.Spec.NoExternalService = true

	if _, err := d.iamClient.AddMembership(ctx, &iam.AddMembershipIn{
		UserId:       string(ctx.UserId),
		ResourceType: string(iamT.ResourceInfraVPNDevice),
		ResourceRef:  iamT.NewResourceRef(ctx.AccountName, iamT.ResourceInfraVPNDevice, device.Name),
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
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeVpnDevice, nDevice.Name, PublishAdd)

	if err := d.resDispatcher.ApplyToTargetCluster(ctx, clusterName, &nDevice.Device, nDevice.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}
	return nDevice, nil
}

func (d *domain) OnVPNDeviceUpdateMessage(ctx InfraContext, clusterName string, device entities.VPNDevice, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xdevice, err := d.findVPNDevice(ctx, clusterName, device.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if _, err := d.matchRecordVersion(device.Annotations, xdevice.RecordVersion); err != nil {
		return d.resyncToTargetCluster(ctx, xdevice.SyncStatus.Action, clusterName, &xdevice.Device, xdevice.RecordVersion)
	}

	recordVersion, err := d.matchRecordVersion(device.Annotations, xdevice.RecordVersion)
	if err != nil {
		return errors.NewE(err)
	}

	upDevice, err := d.vpnDeviceRepo.PatchById(
		ctx,
		xdevice.Id,
		common.PatchForSyncFromAgent(
			&device,
			recordVersion,
			status,
			common.PatchOpts{
				MessageTimestamp: opts.MessageTimestamp,
				XPatch: repos.Document{
					fc.VPNDeviceWireguardConfig: device.WireguardConfig,
				},
			}))

	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeVpnDevice, upDevice.Name, PublishUpdate)
	return nil
}

func (d *domain) OnVPNDeviceDeleteMessage(ctx InfraContext, clusterName string, device entities.VPNDevice) error {
	err := d.vpnDeviceRepo.DeleteOne(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.ClusterName:  clusterName,
			fields.MetadataName: device.Name,
		},
	)
	if err != nil {
		return errors.NewE(err)
	}

	if _, err = d.iamClient.RemoveMembership(ctx, &iam.RemoveMembershipIn{
		UserId:      string(ctx.UserId),
		ResourceRef: iamT.NewResourceRef(ctx.AccountName, iamT.ResourceInfraVPNDevice, device.Name),
	}); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeVpnDevice, device.Name, PublishDelete)
	return nil
}
