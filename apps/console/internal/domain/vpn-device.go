package domain

import (
	"encoding/json"

	"github.com/kloudlite/api/apps/console/internal/entities"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/infra"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
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

	device, err := d.findVPNDevice(ctx, name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.canPerformActionInDevice(ctx, iamT.GetVPNDeviceConnectConfig, device.Name); err != nil {
		d.logger.Infof("user (%s) does not have to access to VPNDevice connect config (%s)", ctx.UserId, device.Name)
		return nil, err
	}

	cluster, err := d.getClusterFromDevice(ctx, device)
	if err != nil {
		return device, err
	}

	gco, err := d.infraClient.GetVpnDevice(ctx, &infra.GetVpnDeviceIn{
		AccountName: ctx.AccountName,
		DeviceName:  name,
		ClusterName: cluster,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if err := json.Unmarshal(gco.VpnDevice, &device.Device); err != nil {
		return nil, errors.NewE(err)
	}

	if err := json.Unmarshal(gco.WgConfig, &device.WireguardConfig); err != nil {
		return nil, errors.NewE(err)
	}

	return device, nil
}

func (d *domain) CreateVPNDevice(ctx ConsoleContext, device entities.ConsoleVPNDevice) (*entities.ConsoleVPNDevice, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateVPNDevice); err != nil {
		return nil, errors.NewE(err)
	}
	device.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	device.LastUpdatedBy = device.CreatedBy
	device.AccountName = ctx.AccountName

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

	d.resourceEventPublisher.PublishVpnDeviceEvent(nDevice, PublishAdd)

	return nDevice, nil
}

func (d *domain) upsertInfraDevice(ctx ConsoleContext, device *entities.ConsoleVPNDevice) (*infra.UpsertVpnDeviceOut, error) {
	clusterName, err := d.getClusterFromDevice(ctx, device)
	if err != nil {
		return nil, errors.NewE(err)
	}

	deviceBytes, err := json.Marshal(device.Device)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return d.infraClient.UpsertVpnDevice(ctx, &infra.UpsertVpnDeviceIn{
		Id:          string(device.Id),
		VpnDevice:   deviceBytes,
		AccountName: ctx.AccountName,
		ClusterName: clusterName,
	})
}

func (d *domain) attachDeviceConfig(ctx ConsoleContext, device *entities.ConsoleVPNDevice) (*entities.ConsoleVPNDevice, error) {
	infraDevOut, err := d.upsertInfraDevice(ctx, device)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := json.Unmarshal(infraDevOut.VpnDevice, &device.Device); err != nil {
		return nil, errors.NewE(err)
	}
	if err := json.Unmarshal(infraDevOut.WgConfig, &device.WireguardConfig); err != nil {
		return nil, errors.NewE(err)
	}

	return device, nil
}

func (d *domain) UpdateVPNDevice(ctx ConsoleContext, device entities.ConsoleVPNDevice) (*entities.ConsoleVPNDevice, error) {
	if err := d.canPerformActionInDevice(ctx, iamT.UpdateVPNDevice, device.Name); err != nil {
		return nil, errors.NewE(err)
	}

	if device.ProjectName == nil || device.EnvironmentName == nil {
		return nil, errors.New("device.projectName, and device.environmentName must be provided")
	}

	envTargetNs, err := d.envTargetNamespace(ctx, *device.ProjectName, *device.EnvironmentName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	xdevice, err := d.findVPNDevice(ctx, device.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	device.Spec.DeviceNamespace = &envTargetNs

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

	nDevice, err := d.vpnDeviceRepo.PatchById(ctx, xdevice.Id, patch)
	if err != nil {
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishVpnDeviceEvent(nDevice, PublishUpdate)

	return d.attachDeviceConfig(ctx, nDevice)
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

	_, err = d.infraClient.DeleteVpnDevice(ctx, &infra.DeleteVpnDeviceIn{
		AccountName: ctx.AccountName,
		Id:          string(device.Id),
	})

	if err != nil {
		return errors.NewE(err)
	}

	if err := d.vpnDeviceRepo.DeleteById(ctx, device.Id); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) UpdateVpnDevicePorts(ctx ConsoleContext, devName string, ports []*wgv1.Port) error {
	currDevice, err := d.findVPNDevice(ctx, devName)
	if err != nil {
		return errors.NewE(err)
	}

	_, err = d.getClusterFromDevice(ctx, currDevice)
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

	nDevice, err := d.vpnDeviceRepo.PatchById(ctx, currDevice.Id, patch)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishVpnDeviceEvent(nDevice, PublishUpdate)

	_, err = d.upsertInfraDevice(ctx, currDevice)
	return errors.NewE(err)
}

func (d *domain) UpdateVpnDeviceEnvironment(ctx ConsoleContext, devName string, projectName string, envName string) error {
	currDevice, err := d.findVPNDevice(ctx, devName)
	if err != nil {
		return errors.NewE(err)
	}

	environment, err := d.findEnvironment(ctx, projectName, envName)
	if err != nil {
		return errors.NewE(err)
	}

	patch := repos.Document{
		"projectName":          projectName,
		"environmentName":      envName,
		"spec.deviceNamespace": environment.Spec.TargetNamespace,
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	}

	ndevice, err := d.vpnDeviceRepo.PatchById(ctx, currDevice.Id, patch)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishVpnDeviceEvent(ndevice, PublishUpdate)

	_, err = d.upsertInfraDevice(ctx, currDevice)
	return errors.NewE(err)
}
