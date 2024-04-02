package domain

import (
	"github.com/kloudlite/api/apps/iot-console/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) findProject(ctx IotConsoleContext, name string) (*entities.IOTProject, error) {
	prj, err := d.iotProjectRepo.FindOne(ctx, repos.Filter{
		"name": name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if prj == nil {
		return nil, errors.Newf("no project with name=%q found", name)
	}
	return prj, nil
}

func (d *domain) ListProjects(ctx IotConsoleContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.IOTProject], error) {
	filter := repos.Filter{
		fields.AccountName: ctx.AccountName,
	}
	return d.iotProjectRepo.FindPaginated(ctx, d.iotDeploymentRepo.MergeMatchFilters(filter, search), pagination)
}

func (d *domain) GetProject(ctx IotConsoleContext, name string) (*entities.IOTProject, error) {
	return d.findProject(ctx, name)
}

func (d *domain) CreateProject(ctx IotConsoleContext, project entities.IOTProject) (*entities.IOTProject, error) {
	project.AccountName = ctx.AccountName
	project.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	project.LastUpdatedBy = project.CreatedBy

	prj, err := d.iotProjectRepo.Create(ctx, &project)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return prj, nil
}

func (d *domain) UpdateProject(ctx IotConsoleContext, project entities.IOTProject) (*entities.IOTProject, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domain) DeleteProject(ctx IotConsoleContext, name string) error {
	err := d.iotProjectRepo.DeleteOne(
		ctx,
		repos.Filter{
			fields.AccountName: ctx.AccountName,
			"name":             name,
		},
	)
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}
