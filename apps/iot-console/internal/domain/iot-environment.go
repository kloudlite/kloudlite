package domain

import (
	"github.com/kloudlite/api/apps/iot-console/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) findEnvironment(ctx IotConsoleContext, projectName string, name string) (*entities.IOTEnvironment, error) {
	env, err := d.iotEnvironmentRepo.FindOne(ctx, repos.Filter{
		fields.AccountName: ctx.AccountName,
		fields.ProjectName: projectName,
		"name":             name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if env == nil {
		return nil, errors.Newf("no environment with name (%s) and project (%s)", name, projectName)
	}
	return env, nil
}

func (d *domain) ListEnvironments(ctx IotConsoleContext, projectName string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.IOTEnvironment], error) {
	filter := repos.Filter{
		fields.AccountName: ctx.AccountName,
		fields.ProjectName: projectName,
	}
	return d.iotEnvironmentRepo.FindPaginated(ctx, d.iotEnvironmentRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) GetEnvironment(ctx IotConsoleContext, projectName string, name string) (*entities.IOTEnvironment, error) {
	return d.findEnvironment(ctx, projectName, name)
}

func (d *domain) CreateEnvironment(ctx IotConsoleContext, projectName string, env entities.IOTEnvironment) (*entities.IOTEnvironment, error) {
	env.ProjectName = projectName
	env.AccountName = ctx.AccountName
	env.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	env.LastUpdatedBy = env.CreatedBy

	nEnv, err := d.iotEnvironmentRepo.Create(ctx, &env)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return nEnv, nil
}

func (d *domain) UpdateEnvironment(ctx IotConsoleContext, projectName string, env entities.IOTEnvironment) (*entities.IOTEnvironment, error) {
	//TODO implement me
	panic("implement me")
}

func (d *domain) DeleteEnvironment(ctx IotConsoleContext, projectName string, name string) error {
	err := d.iotEnvironmentRepo.DeleteOne(
		ctx,
		repos.Filter{
			fields.AccountName: ctx.AccountName,
			fields.ProjectName: projectName,
			"name":             name,
		},
	)
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}
