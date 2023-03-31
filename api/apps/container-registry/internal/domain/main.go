package domain

import (
	"fmt"
	"go.uber.org/fx"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/apps/container-registry/internal/env"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/harbor"
	"kloudlite.io/pkg/repos"
)

type Impl struct {
	harborCli   *harbor.Client
	projectRepo repos.DbRepo[*entities.HarborProject]
}

func (d *Impl) GetHarborCredentials(ctx RegistryContext) (*entities.HarborCredentials, error) {
	one, err := d.projectRepo.FindOne(ctx, repos.Filter{"account_name": ctx.GetAccountName()})
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}
	return &one.Credentials, nil
}

func (d *Impl) CreateHarborProject(ctx RegistryContext) (*entities.HarborProject, error) {
	project, err := d.harborCli.CreateProject(ctx, ctx.GetAccountName())
	if err != nil {
		return nil, err
	}
	robot, err := d.harborCli.CreateRobot(ctx, ctx.GetAccountName(), "svc-account", func() *string {
		s := "Service account for kloudlite"
		return &s
	}(), true)
	create, err := d.projectRepo.Create(ctx, &entities.HarborProject{
		AccountName: ctx.GetAccountName(),
		ProjectId:   project.ProjectId,
		Credentials: entities.HarborCredentials{
			Username: robot.Name,
			Password: robot.Secret,
		},
	})
	if err != nil {
		return nil, err
	}
	return create, nil
}

func (d *Impl) GetImageTags(ctx RegistryContext, repoName string) ([]harbor.ImageTag, error) {
	return d.harborCli.ListTags(ctx, ctx.GetAccountName(), repoName, harbor.ListTagsOpts{
		WithImmutable: false,
		WithSignature: false,
		ListOptions: harbor.ListOptions{
			Page:     1,
			PageSize: 100,
		},
	})
}

func (d *Impl) CreateHarborRobot(ctx RegistryContext, name string, description *string, readOnly bool) (*harbor.Robot, error) {
	return d.harborCli.CreateRobot(ctx, ctx.GetAccountName(), name, description, readOnly)
}

func (d *Impl) GetHarborImages(ctx RegistryContext) ([]harbor.Repository, error) {
	return d.harborCli.SearchRepositories(ctx, ctx.GetAccountName(), "", harbor.ListOptions{
		PageSize: 100,
		Page:     1,
	})
}

func (d *Impl) GetHarborRobots(ctx RegistryContext) ([]harbor.Robot, error) {
	accountName := ctx.GetAccountName()
	one, err := d.projectRepo.FindOne(ctx, repos.Filter{"account_name": accountName})
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, errors.New(fmt.Sprintf("project for account %s not found", accountName))
	}
	return d.harborCli.GetRobots(ctx, one.ProjectId, harbor.ListOptions{
		PageSize: 100,
		Page:     1,
	})
}

func (d *Impl) DeleteHarborRobot(ctx RegistryContext, robotId int) error {
	return d.harborCli.DeleteRobot(ctx, ctx.GetAccountName(), robotId)
}

var Module = fx.Module(
	"domain",
	fx.Provide(func(
		e *env.Env,
		projectRepo repos.DbRepo[*entities.HarborProject],
	) (Domain, error) {
		client, err := harbor.NewClient(harbor.Args{
			HarborAdminUsername: e.HarborAdminUsername,
			HarborAdminPassword: e.HarborAdminPassword,
			HarborRegistryHost:  e.HarborRegistryHost,
			HarborApiVersion:    nil,
		})
		if err != nil {
			return nil, err
		}
		return &Impl{
			harborCli:   client,
			projectRepo: projectRepo,
		}, nil
	}),
)
