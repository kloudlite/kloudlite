package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) findVPNDevice(ctx ConsoleContext, name string) (*entities.ConsoleVPNDevice, error) {
	device, err := d.vpnDeviceRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
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

func (d *domain) ListVPNDevices(ctx ConsoleContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ConsoleVPNDevice], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListVPNDevices); err != nil {
		return nil, errors.NewE(err)
	}

	filter := repos.Filter{"accountName": ctx.AccountName}
	return d.vpnDeviceRepo.FindPaginated(ctx, d.vpnDeviceRepo.MergeMatchFilters(filter, search), pagination)
}

func (d *domain) ListVPNDevicesForUser(ctx ConsoleContext) ([]*entities.ConsoleVPNDevice, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListVPNDevices); err != nil {
		return nil, errors.NewE(err)
	}

	return d.vpnDeviceRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"createdBy.userId": ctx.UserId,
		},
	})
}

func (d *domain) getClusterFromDevice(ctx ConsoleContext, device *entities.ConsoleVPNDevice) (string, error) {
	if device == nil {
		return "", errors.Newf("device is nil")
	}

	if device.ProjectName == nil {
		return "", errors.NewE(errors.Newf("project name is nil"))
	}

	cluster, err := d.getClusterAttachedToProject(ctx, *device.ProjectName)
	if err != nil {
		return "", errors.NewE(err)
	}
	if cluster == nil {
		return "", errors.NewE(errors.Newf("no cluster attached to project %s", *device.ProjectName))
	}
	return *cluster, nil
}

func (d *domain) GetVPNDevice(ctx ConsoleContext, name string) (*entities.ConsoleVPNDevice, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetVPNDevice); err != nil {
		return nil, errors.NewE(err)
	}

	return d.findVPNDevice(ctx, name)
}

func (d *domain) CreateVPNDevice(ctx ConsoleContext, device entities.ConsoleVPNDevice) (*entities.ConsoleVPNDevice, error) {
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

	device.SyncStatus = t.GenSyncStatus(t.SyncActionApply, device.RecordVersion)

	device.Namespace = d.envVars.DeviceNamespace

	if device.ProjectName != nil && device.EnvironmentName != nil {
		s, err := d.envTargetNamespace(ctx, *device.ProjectName, *device.EnvironmentName)
		if err != nil {
			return nil, errors.NewE(err)
		}

		device.Spec.ActiveNamespace = &s
	}

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

	if device.ProjectName != nil && device.EnvironmentName != nil {
		d.applyK8sResource(ctx, *device.ProjectName, &device.Device, device.RecordVersion)
	}

	return nDevice, nil
}

func (d *domain) updateVpnOnCluster(ctx ConsoleContext, ndev, xdev *entities.ConsoleVPNDevice) error {
	if ndev.ProjectName != nil && ndev.EnvironmentName != nil {
		if err := d.applyK8sResource(ctx, *ndev.ProjectName, &ndev.Device, ndev.RecordVersion); err != nil {
			return errors.NewE(err)
		}
	}

	if (xdev.ProjectName != nil) && (*xdev.ProjectName != *ndev.ProjectName) {
		xdev.Spec.Disabled = true
		if err := d.applyK8sResource(ctx, *xdev.ProjectName, &xdev.Device, xdev.RecordVersion); err != nil {
			return errors.NewE(err)
		}
	}

	return nil
}

func (d *domain) UpdateVPNDevice(ctx ConsoleContext, device entities.ConsoleVPNDevice) (*entities.ConsoleVPNDevice, error) {
	if err := d.canPerformActionInDevice(ctx, iamT.UpdateVPNDevice, device.Name); err != nil {
		return nil, errors.NewE(err)
	}

	xdevice, err := d.findVPNDevice(ctx, device.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if device.ProjectName != nil && device.EnvironmentName != nil {
		s, err := d.envTargetNamespace(ctx, *device.ProjectName, *device.EnvironmentName)
		if err != nil {
			return nil, errors.NewE(err)
		}

		device.Spec.ActiveNamespace = &s
	}

	patch := repos.Document{
		"displayName":     device.DisplayName,
		"spec":            device.Spec,
		"projectName":     device.ProjectName,
		"environmentName": device.EnvironmentName,
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	}

	cv, err := d.vpnDeviceRepo.PatchById(ctx, xdevice.Id, patch)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishVpnDeviceEvent(cv, PublishUpdate)

	if err := d.updateVpnOnCluster(ctx, cv, xdevice); err != nil {
		return nil, errors.NewE(err)
	}

	return cv, nil
}

func (d *domain) DeleteVPNDevice(ctx ConsoleContext, name string) error {
	if err := d.canPerformActionInDevice(ctx, iamT.DeleteVPNDevice, name); err != nil {
		return errors.NewE(err)
	}

	device, err := d.findVPNDevice(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishVpnDeviceEvent(device, PublishDelete)

	if err := d.deleteK8sResource(ctx, *device.ProjectName, &device.Device); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) UpdateVpnDevicePorts(ctx ConsoleContext, devName string, ports []*wgv1.Port) error {

	if err := d.canPerformActionInDevice(ctx, iamT.UpdateVPNDevice, devName); err != nil {
		return errors.NewE(err)
	}

	xdevice, err := d.findVPNDevice(ctx, devName)
	if err != nil {
		return errors.NewE(err)
	}

	var prt []wgv1.Port
	for _, p := range ports {
		if p != nil {
			prt = append(prt, *p)
		}
	}

	patch := repos.Document{
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
		"spec.ports": prt,
	}

	nDevice, err := d.vpnDeviceRepo.PatchById(ctx, xdevice.Id, patch)
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishVpnDeviceEvent(nDevice, PublishUpdate)

	if err := d.updateVpnOnCluster(ctx, nDevice, xdevice); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) UpdateVpnDeviceEnvironment(ctx ConsoleContext, devName string, projectName string, envName string) error {
	xdevice, err := d.findVPNDevice(ctx, devName)
	if err != nil {
		return errors.NewE(err)
	}

	envNamesapce, err := d.envTargetNamespace(ctx, projectName, envName)
	if err != nil {
		return errors.NewE(err)
	}

	patch := repos.Document{
		"projectName":          projectName,
		"environmentName":      envName,
		"spec.activeNamespace": envNamesapce,
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	}

	ndevice, err := d.vpnDeviceRepo.PatchById(ctx, xdevice.Id, patch)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishVpnDeviceEvent(ndevice, PublishUpdate)

	if err := d.updateVpnOnCluster(ctx, ndevice, xdevice); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) OnVPNDeviceUpdateMessage(ctx ConsoleContext, device entities.ConsoleVPNDevice, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xdevice, err := d.findVPNDevice(ctx, device.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(device.Annotations, xdevice.RecordVersion); err != nil {
		if xdevice.ProjectName != nil {
			return d.resyncK8sResource(ctx, *xdevice.ProjectName, xdevice.SyncStatus.Action, &xdevice.Device, xdevice.RecordVersion)
		}
	}

	if _, err = d.vpnDeviceRepo.PatchById(ctx, xdevice.Id, repos.Document{
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
			RecordVersion: xdevice.RecordVersion,
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

	d.resourceEventPublisher.PublishVpnDeviceEvent(xdevice, PublishUpdate)
	return nil
}

func (d *domain) OnVPNDeviceDeleteMessage(ctx ConsoleContext, device entities.ConsoleVPNDevice) error {
	xdevice, err := d.findVPNDevice(ctx, device.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err = d.vpnDeviceRepo.DeleteById(ctx, xdevice.Id); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishVpnDeviceEvent(xdevice, PublishUpdate)
	return nil
}

func (d *domain) OnVPNDeviceApplyError(ctx ConsoleContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	panic("h")
}
