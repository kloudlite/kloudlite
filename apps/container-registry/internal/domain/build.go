package domain

import (
	"fmt"

	"kloudlite.io/apps/container-registry/internal/domain/entities"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/common"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/repos"
)

func (d *Impl) AddBuild(ctx RegistryContext, build entities.Build) (*entities.Build, error) {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
		},
		Action: string(iamT.UpdateAccount),
	})

	if err != nil {
		return nil, err
	}

	if !co.Status {
		return nil, fmt.Errorf("unauthorized to add build")
	}

	return d.buildRepo.Create(ctx, &entities.Build{
		Name:        build.Name,
		AccountName: ctx.AccountName,
		Repository:  build.Repository,
		Source:      build.Source,
		Tag:         build.Tag,
		CreatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	})
}

func (d *Impl) UpdateBuild(ctx RegistryContext, id repos.ID, build entities.Build) (*entities.Build, error) {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
		},
		Action: string(iamT.UpdateAccount),
	})

	if err != nil {
		return nil, err
	}

	if !co.Status {
		return nil, fmt.Errorf("unauthorized to update build")
	}

	return d.buildRepo.UpdateById(ctx, id, &entities.Build{
		Name:        build.Name,
		AccountName: ctx.AccountName,
		Repository:  build.Repository,
		Source:      build.Source,
		Tag:         build.Tag,
		LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	})
}

func (d *Impl) ListBuilds(ctx RegistryContext, repoName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Build], error) {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
		},
		Action: string(iamT.GetAccount),
	})

	if err != nil {
		return nil, err
	}

	if !co.Status {
		return nil, fmt.Errorf("unauthorized to list builds")
	}

	filter := repos.Filter{"accountName": ctx.AccountName, "repository": repoName}

	return d.buildRepo.FindPaginated(ctx, d.buildRepo.MergeMatchFilters(filter, search), pagination)
}

func (d *Impl) GetBuild(ctx RegistryContext, buildId repos.ID) (*entities.Build, error) {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
		},
		Action: string(iamT.GetAccount),
	})

	if err != nil {
		return nil, err
	}

	if !co.Status {
		return nil, fmt.Errorf("unauthorized to get build")
	}

	b, err := d.buildRepo.FindOne(ctx, repos.Filter{"accountName": ctx.AccountName, "id": buildId})
	if err != nil {
		return nil, err
	}

	if b == nil {
		return nil, fmt.Errorf("build not found")
	}

	return b, nil
}

func (d *Impl) DeleteBuild(ctx RegistryContext, buildId repos.ID) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
		},
		Action: string(iamT.UpdateAccount),
	})

	if err != nil {
		return err
	}

	if !co.Status {
		return fmt.Errorf("unauthorized to delete build")
	}

	return d.buildRepo.DeleteOne(ctx, repos.Filter{"accountName": ctx.AccountName, "id": buildId})
}

func (d *Impl) TriggerBuild(ctx RegistryContext, buildId repos.ID) error {
	panic("implement me")
}
