package domain

import (
	"fmt"

	"encoding/json"

	opHarbor "github.com/kloudlite/operator/pkg/harbor"
	"github.com/kloudlite/operator/pkg/kubectl"
	"go.uber.org/fx"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/apps/container-registry/internal/env"
	"kloudlite.io/constants"
	"kloudlite.io/pkg/harbor"
	"kloudlite.io/pkg/k8s"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

type Impl struct {
	harborCli            *harbor.Client
	k8sExtendedClient    k8s.ExtendedK8sClient
	harborProjectRepo    repos.DbRepo[*entities.HarborProject]
	harborRobotUsersRepo repos.DbRepo[*entities.HarborRobotUser]
	k8sYamlCli           *kubectl.YAMLClient
}

func (i *Impl) findHarborRobot(ctx RegistryContext, name string) (*entities.HarborRobotUser, error) {
	hru, err := i.harborRobotUsersRepo.FindOne(ctx, repos.Filter{"metadata.name": name, "spec.accountName": ctx.accountName})
	if err != nil {
		return nil, err
	}
	if hru == nil {
		return nil, fmt.Errorf("no robot user account found for (name: %q, accountName: %q)", name, ctx.accountName)
	}

	return hru, nil
}

// ReSyncHarborRobot implements Domain
func (d *Impl) ReSyncHarborRobot(ctx RegistryContext, name string) error {
	hru, err := d.findHarborRobot(ctx, name)
	if err != nil {
		return err
	}

	b, err := json.Marshal(hru.HarborUserAccount)
	if err != nil {
		return err
	}

	_, err = d.k8sYamlCli.ApplyYAML(ctx, b)
	return err
}

// UpdateHarborRobot implements Domain
func (d *Impl) UpdateHarborRobot(ctx RegistryContext, name string, permissions []opHarbor.Permission) (*entities.HarborRobotUser, error) {
	hru, err := d.findHarborRobot(ctx, name)
	if err != nil {
		return nil, err
	}

	hru.Spec.Permissions = permissions
	upHru, err := d.harborRobotUsersRepo.UpdateById(ctx, hru.Id, hru)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(hru.HarborUserAccount)
	if err != nil {
		return nil, err
	}

	_, err = d.k8sYamlCli.ApplyYAML(ctx, b)
	if err != nil {
		return nil, err
	}

	return upHru, nil
}

func (d *Impl) GetHarborCredentials(ctx RegistryContext) (*entities.HarborCredentials, error) {
	one, err := d.harborProjectRepo.FindOne(ctx, repos.Filter{"account_name": ctx.GetAccountName()})
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}
	return &one.Credentials, nil
}

// func (d *Impl) CreateHarborProject(ctx RegistryContext) (*entities.HarborProject, error) {
// 	project, err := d.harborCli.CreateProject(ctx, ctx.GetAccountName())
// 	if err != nil {
// 		return nil, err
// 	}
// 	robot, err := d.harborCli.CreateRobot(ctx, ctx.GetAccountName(), "svc-account", func() *string {
// 		s := "Service account for kloudlite"
// 		return &s
// 	}(), true)
// 	create, err := d.harborProjectRepo.Create(ctx, &entities.HarborProject{
// 		AccountName: ctx.GetAccountName(),
// 		ProjectId:   project.ProjectId,
// 		Credentials: entities.HarborCredentials{
// 			Username: robot.Name,
// 			Password: robot.Secret,
// 		},
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return create, nil
// }

func (d *Impl) GetRepoArtifacts(ctx RegistryContext, repoName string) ([]harbor.Artifact, error) {
	return d.harborCli.ListArtifacts(ctx, ctx.GetAccountName(), repoName, harbor.ListTagsOpts{
		WithImmutable: false,
		WithSignature: false,
		ListOptions: harbor.ListOptions{
			Page:     1,
			PageSize: 100,
		},
	})
}

func (d *Impl) CreateHarborRobot(ctx RegistryContext, hru *entities.HarborRobotUser) (*entities.HarborRobotUser, error) {
	hru.EnsureGVK()
	hru.Namespace = constants.NamespaceCore
	hru.Spec.AccountName = ctx.accountName

	if err := d.k8sExtendedClient.ValidateStruct(ctx, &hru.HarborUserAccount); err != nil {
		return nil, err
	}

	hru.SyncStatus = t.GetSyncStatusForCreation()
	nHru, err := d.harborRobotUsersRepo.Create(ctx, hru)
	if err != nil {
		if d.harborRobotUsersRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("harbor robot user with name %q, already exists", hru.Name)
		}
		return nil, err
	}

	b, err := json.Marshal(nHru.HarborUserAccount)
	if err != nil {
		return nil, err
	}

	if _, err := d.k8sYamlCli.ApplyYAML(ctx, b); err != nil {
		return nil, err
	}

	return nHru, nil
}

func (d *Impl) GetHarborImages(ctx RegistryContext) ([]harbor.Repository, error) {
	return d.harborCli.SearchRepositories(ctx, ctx.GetAccountName(), "", harbor.ListOptions{
		PageSize: 100,
		Page:     1,
	})
}

func (d *Impl) ListHarborRobots(ctx RegistryContext) ([]*entities.HarborRobotUser, error) {
	return d.harborRobotUsersRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{"spec.accountName": ctx.accountName},
	})

	// if one == nil {
	// 	return nil, errors.New(fmt.Sprintf("project for account %s not found", accountName))
	// }

	// return d.harborCli.GetRobots(ctx, one.ProjectId, harbor.ListOptions{
	// 	PageSize: 100,
	// 	Page:     1,
	// })
}

func (d *Impl) DeleteHarborRobot(ctx RegistryContext, robotId int) error {
	return d.harborCli.DeleteRobot(ctx, ctx.GetAccountName(), robotId)
}

var Module = fx.Module(
	"domain",
	fx.Provide(
		func(e *env.Env,
			projectRepo repos.DbRepo[*entities.HarborProject],
			hruRepo repos.DbRepo[*entities.HarborRobotUser],
			k8sYamlCli *kubectl.YAMLClient,
			k8sExtendedClient k8s.ExtendedK8sClient,
		) (Domain, error) {
			client, err := harbor.NewClient(harbor.Args{
				HarborAdminUsername: e.HarborAdminUsername,
				HarborAdminPassword: e.HarborAdminPassword,
				HarborRegistryHost:  e.HarborRegistryHost,
			})
			if err != nil {
				return nil, err
			}
			return &Impl{
				harborCli:            client,
				harborProjectRepo:    projectRepo,
				harborRobotUsersRepo: hruRepo,
				k8sYamlCli:           k8sYamlCli,
				k8sExtendedClient:    k8sExtendedClient,
			}, nil
		}),
)
