package domain

import (
	"github.com/kloudlite/api/apps/iot-console/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) findDevice(ctx IotResourceContext, name string) (*entities.IOTDevice, error) {
	filter := ctx.IOTConsoleDBFilters()
	filter.Add("name", name)
	dev, err := d.iotDeviceRepo.FindOne(ctx, ctx.IOTConsoleDBFilters().Add("name", name))
	if err != nil {
		return nil, errors.NewE(err)
	}
	if dev == nil {
		return nil, errors.Newf("no device with name=%q found", name)
	}
	return dev, nil
}

func (d *domain) ListDevices(ctx IotResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.IOTDevice], error) {
	filter := ctx.IOTConsoleDBFilters()
	return d.iotDeviceRepo.FindPaginated(ctx, d.iotDeviceRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) GetDevice(ctx IotResourceContext, name string) (*entities.IOTDevice, error) {
	return d.findDevice(ctx, name)
}

func (d *domain) CreateDevice(ctx IotResourceContext, device entities.IOTDevice) (*entities.IOTDevice, error) {
	device.ProjectName = ctx.ProjectName
	device.AccountName = ctx.AccountName
	device.EnvironmentName = ctx.EnvironmentName
	device.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	device.LastUpdatedBy = device.CreatedBy

	nDevice, err := d.iotDeviceRepo.Create(ctx, &device)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return nDevice, nil
}

func (d *domain) UpdateDevice(ctx IotResourceContext, device entities.IOTDevice) (*entities.IOTDevice, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domain) DeleteDevice(ctx IotResourceContext, name string) error {
	err := d.iotDeviceRepo.DeleteOne(
		ctx,
		ctx.IOTConsoleDBFilters().Add("name", name),
	)
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}
