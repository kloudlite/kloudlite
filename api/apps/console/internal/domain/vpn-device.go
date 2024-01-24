package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	wgv1 "github.com/kloudlite/operator/apis/wireguard/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/strings/slices"
)

func (d *domain) findVPNDevice(ctx ConsoleContext, name string) (*entities.ConsoleVPNDevice, error) {
	device, err := d.vpnDeviceRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
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

func (d *domain) updateVpnOnCluster(ctx ConsoleContext, ndev, xdev *entities.ConsoleVPNDevice) error {
	ndev.Namespace = d.envVars.DeviceNamespace
	ndev.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &ndev.Device); err != nil {
		return errors.NewE(err)
	}

	if ndev.ProjectName != nil && ndev.EnvironmentName != nil {
		if err := d.applyVPNDevice(ctx, ndev); err != nil {
			return errors.NewE(err)
		}
	}

	if (xdev.ProjectName != nil) && (*xdev.ProjectName != *ndev.ProjectName) {
		xdev.Spec.Disabled = true
		if err := d.applyVPNDevice(ctx, xdev); err != nil {
			return errors.NewE(err)
		}
	}

	return nil
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

func (d *domain) GetVPNDevice(ctx ConsoleContext, name string) (*entities.ConsoleVPNDevice, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetVPNDevice); err != nil {
		return nil, errors.NewE(err)
	}

	return d.findVPNDevice(ctx, name)
}

func (d *domain) applyVPNDevice(ctx ConsoleContext, device *entities.ConsoleVPNDevice) error {
	if err := d.applyK8sResource(ctx, *device.ProjectName, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: d.envVars.DeviceNamespace,
			Annotations: map[string]string{
				constants.DescriptionKey: "namespace created by kloudlite platform to manage infra level VPN Devices",
			},
		},
	}, device.RecordVersion); err != nil {
		return errors.NewE(err)
	}

	if device.ProjectName != nil {
		if err := d.applyK8sResource(ctx, *device.ProjectName, &device.Device, device.RecordVersion); err != nil {
			return errors.NewE(err)
		}
	}

	return nil
}

func (d *domain) CreateVPNDevice(ctx ConsoleContext, device entities.ConsoleVPNDevice) (*entities.ConsoleVPNDevice, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateVPNDevice); err != nil {
		return nil, errors.NewE(err)
	}

	device.Namespace = d.envVars.DeviceNamespace

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
	device.LinkedClusters = []string{}

	device.SyncStatus = t.GenSyncStatus(t.SyncActionApply, device.RecordVersion)

	if device.ProjectName != nil && device.EnvironmentName != nil {
		s, err := d.envTargetNamespace(ctx, *device.ProjectName, *device.EnvironmentName)
		if err != nil {
			return nil, errors.NewE(err)
		}

		device.Spec.ActiveNamespace = &s

		clusterName, err := d.getClusterFromDevice(ctx, &device)
		if err != nil {
			return nil, errors.NewE(err)
		}

		device.LinkedClusters = append(device.LinkedClusters, clusterName)
	}

	if _, err := d.iamClient.AddMembership(ctx, &iam.AddMembershipIn{
		UserId:       string(ctx.UserId),
		ResourceType: string(iamT.ResourceConsoleVPNDevice),
		ResourceRef:  iamT.NewResourceRef(ctx.AccountName, iamT.ResourceConsoleVPNDevice, device.Name),
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

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeVPNDevice, nDevice.Name, PublishAdd)

	if device.ProjectName == nil || device.EnvironmentName == nil {
		return nDevice, nil
	}

	if err := d.applyVPNDevice(ctx, nDevice); err != nil {
		return nDevice, err
	}
	return nDevice, nil
}

func (d *domain) UpdateVPNDevice(ctx ConsoleContext, device entities.ConsoleVPNDevice) (*entities.ConsoleVPNDevice, error) {
	if err := d.canPerformActionInDevice(ctx, iamT.UpdateVPNDevice, device.Name); err != nil {
		return nil, errors.NewE(err)
	}

	xdevice, err := d.findVPNDevice(ctx, device.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	linkedClusters := xdevice.LinkedClusters

	device.Spec.ActiveNamespace = nil

	if device.ProjectName != nil && device.EnvironmentName != nil {
		activeNamespace, err := d.envTargetNamespace(ctx, *device.ProjectName, *device.EnvironmentName)
		if err != nil {
			return nil, errors.NewE(err)
		}
		device.Spec.ActiveNamespace = &activeNamespace

		clusterName, err := d.getClusterAttachedToProject(ctx, *device.ProjectName)
		if err != nil {
			return nil, errors.NewE(err)
		}

		if clusterName != nil && !slices.Contains(linkedClusters, *clusterName) {
			linkedClusters = append(linkedClusters, *clusterName)
		}
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&device,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.ConsoleVPNDeviceSpec:           device.Spec,
				fields.ProjectName:                device.ProjectName,
				fields.EnvironmentName:            device.EnvironmentName,
				fc.ConsoleVPNDeviceLinkedClusters: linkedClusters,
			},
		})

	upDevice, err := d.vpnDeviceRepo.Patch(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: device.Name,
	}, patchForUpdate)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeVPNDevice, device.Name, PublishUpdate)

	if err := d.updateVpnOnCluster(ctx, upDevice, xdevice); err != nil {
		return nil, errors.NewE(err)
	}

	return upDevice, nil
}

func (d *domain) DeleteVPNDevice(ctx ConsoleContext, name string) error {
	if err := d.canPerformActionInDevice(ctx, iamT.DeleteVPNDevice, name); err != nil {
		return errors.NewE(err)
	}

	upDevice, err := d.vpnDeviceRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: name,
		},
		common.PatchForMarkDeletion(),
	)
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeVPNDevice, name, PublishUpdate)

	for _, v := range upDevice.LinkedClusters {
		if err := d.deleteK8sResourceOfCluster(ctx, v, &upDevice.Device); err != nil {
			return errors.NewE(err)
		}
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

	var p []wgv1.Port
	for _, port := range ports {
		p = append(p, *port)
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		xdevice,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.ConsoleVPNDeviceSpecPorts: p,
			},
		})

	upDevice, err := d.vpnDeviceRepo.Patch(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: devName,
	}, patchForUpdate)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.applyVPNDevice(ctx, upDevice); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeVPNDevice, devName, PublishUpdate)
	return nil
}

func (d *domain) UpdateVpnDeviceEnvironment(ctx ConsoleContext, devName string, projectName string, envName string) error {
	xdevice, err := d.findVPNDevice(ctx, devName)
	if err != nil {
		return errors.NewE(err)
	}

	xdevice.ProjectName = &projectName
	xdevice.EnvironmentName = &envName

	_, err = d.UpdateVPNDevice(ctx, *xdevice)
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) OnVPNDeviceUpdateMessage(ctx ConsoleContext, device entities.ConsoleVPNDevice, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xdevice, err := d.findVPNDevice(ctx, device.Name)
	if err != nil {
		return errors.NewE(err)
	}

	recordVersion, err := d.MatchRecordVersion(device.Annotations, xdevice.RecordVersion)
	if err != nil {
		if xdevice.ProjectName != nil {
			return d.resyncK8sResource(ctx, *xdevice.ProjectName, xdevice.SyncStatus.Action, &xdevice.Device, xdevice.RecordVersion)
		}
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
					fc.ConsoleVPNDeviceWireguardConfig: device.WireguardConfig,
				},
			}))
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeVPNDevice, upDevice.Name, PublishUpdate)

	return nil
}

func (d *domain) OnVPNDeviceDeleteMessage(ctx ConsoleContext, device entities.ConsoleVPNDevice) error {
	xdevice, err := d.findVPNDevice(ctx, device.Name)
	if err != nil {
		return errors.NewE(err)
	}

	var linkedClusters []string
	if device.ProjectName != nil {
		clusterName, err := d.getClusterAttachedToProject(ctx, *device.ProjectName)
		if err != nil {
			return errors.NewE(err)
		}
		if clusterName == nil {
			return errors.Newf("No Cluster found")
		}
		var linkedClusters []string
		slices.Filter(linkedClusters, xdevice.LinkedClusters, func(item string) bool {
			return item != *clusterName
		})
	}

	if len(linkedClusters) == 0 {
		err := d.vpnDeviceRepo.DeleteOne(
			ctx,
			repos.Filter{
				fields.AccountName:  ctx.AccountName,
				fields.MetadataName: device.Name,
			},
		)
		if err != nil {
			return errors.NewE(err)
		}

		if _, err = d.iamClient.RemoveMembership(ctx, &iam.RemoveMembershipIn{
			UserId:      string(ctx.UserId),
			ResourceRef: iamT.NewResourceRef(ctx.AccountName, iamT.ResourceConsoleVPNDevice, device.Name),
		}); err != nil {
			return errors.NewE(err)
		}

		d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeVPNDevice, device.Name, PublishDelete)
		return nil
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		xdevice,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.ConsoleVPNDeviceLinkedClusters: linkedClusters,
			},
		})
	_, err = d.vpnDeviceRepo.PatchById(ctx, xdevice.Id, patchForUpdate)

	return errors.NewE(err)
}

func (d *domain) OnVPNDeviceApplyError(ctx ConsoleContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	udevice, err := d.vpnDeviceRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
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
	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeVPNDevice, udevice.Name, PublishDelete)
	return errors.NewE(err)
}
