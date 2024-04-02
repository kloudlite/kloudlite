package domain

import (
	"github.com/kloudlite/api/apps/iot-console/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) findApp(ctx IotResourceContext, deviceBlueprintName string, name string) (*entities.IOTApp, error) {
	filter := ctx.IOTConsoleDBFilters()
	filter.Add("deviceBlueprintName", deviceBlueprintName)
	filter.Add("name", name)
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
	filter.Add("deviceBlueprintName", deviceBlueprintName)
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
	//TODO implement me
	panic("implement me")
}

func (d *domain) DeleteApp(ctx IotResourceContext, deviceBlueprintName string, name string) error {
	filter := ctx.IOTConsoleDBFilters()
	filter.Add("deviceBlueprintName", deviceBlueprintName)
	filter.Add("name", name)
	err := d.iotAppRepo.DeleteOne(
		ctx,
		filter,
	)
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}
