package domain

import (
	"github.com/kloudlite/api/apps/iot-console/internal/entities"
	fc "github.com/kloudlite/api/apps/iot-console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *domain) findDeployment(ctx IotResourceContext, name string) (*entities.IOTDeployment, error) {
	prj, err := d.iotDeploymentRepo.FindOne(ctx, ctx.IOTConsoleDBFilters().Add(fc.IOTDeploymentName, name))
	if err != nil {
		return nil, errors.NewE(err)
	}
	if prj == nil {
		return nil, errors.Newf("no deployment with name=%q found", name)
	}
	return prj, nil
}

func (d domain) ListDeployments(ctx IotResourceContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.IOTDeployment], error) {
	filter := ctx.IOTConsoleDBFilters()
	return d.iotDeploymentRepo.FindPaginated(ctx, d.iotDeploymentRepo.MergeMatchFilters(filter, search), pagination)
}

func (d domain) GetDeployment(ctx IotResourceContext, name string) (*entities.IOTDeployment, error) {
	return d.findDeployment(ctx, name)
}

func (d domain) CreateDeployment(ctx IotResourceContext, deployment entities.IOTDeployment) (*entities.IOTDeployment, error) {
	deployment.AccountName = ctx.AccountName
	deployment.ProjectName = ctx.ProjectName
	deployment.EnvironmentName = ctx.EnvironmentName
	deployment.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	deployment.LastUpdatedBy = deployment.CreatedBy

	dep, err := d.iotDeploymentRepo.Create(ctx, &deployment)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return dep, nil
}

func (d domain) UpdateDeployment(ctx IotResourceContext, deployment entities.IOTDeployment) (*entities.IOTDeployment, error) {
	patchForUpdate := repos.Document{
		fields.DisplayName: deployment.DisplayName,
		fields.LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.GetUserId(),
			UserName:  ctx.GetUserName(),
			UserEmail: ctx.GetUserEmail(),
		},
	}

	patchFilter := ctx.IOTConsoleDBFilters().Add(fc.IOTDeploymentName, deployment.Name)

	upDep, err := d.iotDeploymentRepo.Patch(
		ctx,
		patchFilter,
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return upDep, nil
}

func (d domain) DeleteDeployment(ctx IotResourceContext, name string) error {
	err := d.iotDeploymentRepo.DeleteOne(
		ctx,
		ctx.IOTConsoleDBFilters().Add(fc.IOTDeploymentName, name),
	)
	if err != nil {
		return errors.NewE(err)
	}
	return nil
}
