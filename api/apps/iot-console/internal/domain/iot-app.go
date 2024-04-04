package domain

import (
	"github.com/kloudlite/api/apps/iot-console/internal/entities"
	fc "github.com/kloudlite/api/apps/iot-console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) findApp(ctx IotResourceContext, deviceBlueprintName string, name string) (*entities.IOTApp, error) {
	filter := ctx.IOTConsoleDBFilters()
	filter.Add(fc.IOTAppDeviceBlueprintName, deviceBlueprintName)
	filter.Add(fields.MetadataName, name)
	app, err := d.iotAppRepo.FindOne(ctx, filter)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if app == nil {
		return nil, errors.Newf("no app with name=%q found", name)
	}
	return app, nil
}

func (d *domain) ListApps(ctx IotResourceContext, deviceBlueprintName string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.IOTApp], error) {
	filter := ctx.IOTConsoleDBFilters()
	filter.Add(fc.IOTAppDeviceBlueprintName, deviceBlueprintName)
	return d.iotAppRepo.FindPaginated(ctx, d.iotDeviceRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) GetApp(ctx IotResourceContext, deviceBlueprintName string, name string) (*entities.IOTApp, error) {
	return d.findApp(ctx, deviceBlueprintName, name)
}

func (d *domain) CreateApp(ctx IotResourceContext, deviceBlueprintName string, app entities.IOTApp) (*entities.IOTApp, error) {
	app.ProjectName = ctx.ProjectName
	app.AccountName = ctx.AccountName
	app.EnvironmentName = ctx.EnvironmentName
	app.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	app.LastUpdatedBy = app.CreatedBy
	app.DeviceBlueprintName = deviceBlueprintName

	nApp, err := d.iotAppRepo.Create(ctx, &app)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return nApp, nil
}

func (d *domain) UpdateApp(ctx IotResourceContext, deviceBlueprintName string, app entities.IOTApp) (*entities.IOTApp, error) {
	patchForUpdate := repos.Document{
		fields.DisplayName: app.DisplayName,
		fields.LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.GetUserId(),
			UserName:  ctx.GetUserName(),
			UserEmail: ctx.GetUserEmail(),
		},
		fc.IOTAppDeviceBlueprintName: deviceBlueprintName,
	}
	patchFilter := ctx.IOTConsoleDBFilters()
	patchFilter.Add(fc.IOTAppDeviceBlueprintName, deviceBlueprintName)
	patchFilter.Add(fields.MetadataName, app.Name)

	upApp, err := d.iotAppRepo.Patch(
		ctx,
		patchFilter,
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return upApp, nil
}

func (d *domain) DeleteApp(ctx IotResourceContext, deviceBlueprintName string, name string) error {
	filter := ctx.IOTConsoleDBFilters()
	filter.Add(fc.IOTAppDeviceBlueprintName, deviceBlueprintName)
	filter.Add(fields.MetadataName, name)
	err := d.iotAppRepo.DeleteOne(
		ctx,
		filter,
	)
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}
