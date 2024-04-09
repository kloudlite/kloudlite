package domain

import (
	"github.com/kloudlite/api/apps/iot-console/internal/entities"
	fc "github.com/kloudlite/api/apps/iot-console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) findDevice(ctx IotResourceContext, name string, deviceBlueprintName string) (*entities.IOTDevice, error) {
	filter := ctx.IOTConsoleDBFilters()
	filter.Add(fc.IOTDeviceDeviceBlueprintName, deviceBlueprintName)
	filter.Add(fc.IOTDeviceName, name)
	dev, err := d.iotDeviceRepo.FindOne(ctx, filter)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if dev == nil {
		return nil, errors.Newf("no device with name=%q found", name)
	}
	return dev, nil
}

func (d *domain) findDeploymentDevice(ctx IotResourceContext, name string, deploymentName string) (*entities.IOTDevice, error) {
	filter := ctx.IOTConsoleDBFilters()
	filter.Add(fc.IOTDeviceDeployment, deploymentName)
	filter.Add(fc.IOTDeviceName, name)
	dev, err := d.iotDeviceRepo.FindOne(ctx, filter)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if dev == nil {
		return nil, errors.Newf("no deployment device with name=%q found", name)
	}
	return dev, nil
}

func (d *domain) ListDevices(ctx IotResourceContext, deviceBlueprintName string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.IOTDevice], error) {
	//filter := ctx.IOTConsoleDBFilters()
	filter := ctx.IOTConsoleDBFilters()
	filter.Add(fc.IOTDeviceDeviceBlueprintName, deviceBlueprintName)
	return d.iotDeviceRepo.FindPaginated(ctx, d.iotDeviceRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) ListDeploymentDevices(ctx IotResourceContext, deploymentName string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.IOTDevice], error) {
	filter := ctx.IOTConsoleDBFilters()
	filter.Add(fc.IOTDeviceDeployment, deploymentName)
	return d.iotDeviceRepo.FindPaginated(ctx, d.iotDeviceRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) GetDevice(ctx IotResourceContext, name string, deviceBlueprintName string) (*entities.IOTDevice, error) {
	return d.findDevice(ctx, name, deviceBlueprintName)
}

func (d *domain) GetDeploymentDevice(ctx IotResourceContext, name string, deploymentName string) (*entities.IOTDevice, error) {
	return d.findDeploymentDevice(ctx, name, deploymentName)
}

func (d *domain) CreateDevice(ctx IotResourceContext, deviceBlueprintName string, device entities.IOTDevice) (*entities.IOTDevice, error) {
	device.ProjectName = ctx.ProjectName
	device.AccountName = ctx.AccountName
	device.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	device.LastUpdatedBy = device.CreatedBy
	device.DeviceBlueprintName = deviceBlueprintName

	nDevice, err := d.iotDeviceRepo.Create(ctx, &device)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return nDevice, nil
}

func (d *domain) AddDeviceToDeployment(ctx IotResourceContext, deploymentName string, deviceName string, deviceBlueprintName string) (*entities.IOTDevice, error) {
	patchForUpdate := repos.Document{
		fields.LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.GetUserId(),
			UserName:  ctx.GetUserName(),
			UserEmail: ctx.GetUserEmail(),
		},
		fc.IOTDeviceDeployment: deploymentName,
	}

	patchFilter := ctx.IOTConsoleDBFilters()
	patchFilter.Add(fc.IOTDeviceDeviceBlueprintName, deviceBlueprintName)
	patchFilter.Add(fc.IOTDeviceName, deviceName)

	upDev, err := d.iotDeviceRepo.Patch(
		ctx,
		patchFilter,
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return upDev, nil
}

func (d *domain) RemoveDeviceOfDeployment(ctx IotResourceContext, deploymentName string, deviceName string, deviceBlueprintName string) (*entities.IOTDevice, error) {
	patchForUpdate := repos.Document{
		fields.LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.GetUserId(),
			UserName:  ctx.GetUserName(),
			UserEmail: ctx.GetUserEmail(),
		},
		fc.IOTDeviceDeployment: "",
	}

	patchFilter := ctx.IOTConsoleDBFilters()
	patchFilter.Add(fc.IOTDeviceDeviceBlueprintName, deviceBlueprintName)
	patchFilter.Add(fc.IOTDeviceName, deviceName)
	patchFilter.Add(fc.IOTDeviceDeployment, deploymentName)

	upDev, err := d.iotDeviceRepo.Patch(
		ctx,
		patchFilter,
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return upDev, nil
}

func (d *domain) UpdateDevice(ctx IotResourceContext, deviceBlueprintName string, device entities.IOTDevice) (*entities.IOTDevice, error) {
	patchForUpdate := repos.Document{
		fields.DisplayName: device.DisplayName,
		fields.LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.GetUserId(),
			UserName:  ctx.GetUserName(),
			UserEmail: ctx.GetUserEmail(),
		},
	}

	patchFilter := ctx.IOTConsoleDBFilters()
	patchFilter.Add(fc.IOTDeviceDeviceBlueprintName, deviceBlueprintName)
	patchFilter.Add(fc.IOTDeviceName, device.Name)

	upDev, err := d.iotDeviceRepo.Patch(
		ctx,
		patchFilter,
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return upDev, nil
}

func (d *domain) DeleteDevice(ctx IotResourceContext, deviceBlueprintName string, name string) error {
	filter := ctx.IOTConsoleDBFilters()
	filter.Add(fc.IOTDeviceDeviceBlueprintName, deviceBlueprintName)
	filter.Add(fc.IOTDeviceName, name)
	err := d.iotDeviceRepo.DeleteOne(
		ctx,
		filter,
	)
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}
