package domain

import (
	"github.com/kloudlite/api/apps/iot-console/internal/entities"
	fc "github.com/kloudlite/api/apps/iot-console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) findDeviceBlueprint(ctx IotResourceContext, name string) (*entities.IOTDeviceBlueprint, error) {
	filter := ctx.IOTConsoleDBFilters()
	filter.Add(fc.IOTDeviceBlueprintName, name)
	devBlueprint, err := d.iotDeviceBlueprintRepo.FindOne(ctx, ctx.IOTConsoleDBFilters().Add("name", name))
	if err != nil {
		return nil, errors.NewE(err)
	}
	if devBlueprint == nil {
		return nil, errors.Newf("no device Blueprint with name=%q found", name)
	}
	return devBlueprint, nil
}

func (d *domain) ListDeviceBlueprints(ctx IotResourceContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.IOTDeviceBlueprint], error) {
	filter := ctx.IOTConsoleDBFilters()
	return d.iotDeviceBlueprintRepo.FindPaginated(ctx, d.iotDeviceBlueprintRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) GetDeviceBlueprint(ctx IotResourceContext, name string) (*entities.IOTDeviceBlueprint, error) {
	return d.findDeviceBlueprint(ctx, name)
}

func (d *domain) CreateDeviceBlueprint(ctx IotResourceContext, deviceBlueprint entities.IOTDeviceBlueprint) (*entities.IOTDeviceBlueprint, error) {
	deviceBlueprint.ProjectName = ctx.ProjectName
	deviceBlueprint.AccountName = ctx.AccountName
	deviceBlueprint.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	deviceBlueprint.LastUpdatedBy = deviceBlueprint.CreatedBy

	if deviceBlueprint.BluePrintType == "singleton_blueprint" {
		deviceBlueprint.BluePrintType = entities.SingletonBlueprint
	} else {
		deviceBlueprint.BluePrintType = entities.GroupBlueprint
	}

	nDeviceBlueprint, err := d.iotDeviceBlueprintRepo.Create(ctx, &deviceBlueprint)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return nDeviceBlueprint, nil
}

func (d *domain) UpdateDeviceBlueprint(ctx IotResourceContext, deviceBlueprint entities.IOTDeviceBlueprint) (*entities.IOTDeviceBlueprint, error) {
	patchForUpdate := repos.Document{
		fields.DisplayName: deviceBlueprint.DisplayName,
		fields.LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.GetUserId(),
			UserName:  ctx.GetUserName(),
			UserEmail: ctx.GetUserEmail(),
		},
	}

	patchFilter := ctx.IOTConsoleDBFilters().Add(fc.IOTDeviceBlueprintName, deviceBlueprint.Name)

	upDevBlueprint, err := d.iotDeviceBlueprintRepo.Patch(
		ctx,
		patchFilter,
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return upDevBlueprint, nil
}

func (d *domain) DeleteDeviceBlueprint(ctx IotResourceContext, name string) error {
	err := d.iotDeviceBlueprintRepo.DeleteOne(
		ctx,
		ctx.IOTConsoleDBFilters().Add(fc.IOTDeviceBlueprintName, name),
	)
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}
