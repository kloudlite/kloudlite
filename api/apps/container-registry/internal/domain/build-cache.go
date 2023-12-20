package domain

import (
	"github.com/kloudlite/api/apps/container-registry/internal/domain/entities"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

func (d *Impl) AddBuildCache(ctx RegistryContext, buildCache entities.BuildCacheKey) (*entities.BuildCacheKey, error) {
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
		return nil, errors.Newf("unauthorized to add build cache")
	}

	buildCache.AccountName = ctx.AccountName
	return d.buildCacheRepo.Create(ctx, &buildCache)
}

func (d *Impl) UpdateBuildCache(ctx RegistryContext, id repos.ID, buildCache entities.BuildCacheKey) (*entities.BuildCacheKey, error) {
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
		return nil, errors.Newf("unauthorized to update build cache")
	}

	back, err := d.buildCacheRepo.FindOne(ctx, repos.Filter{
		"accountName": ctx.AccountName,
		"id":          id,
	})
	if err != nil {
		return nil, err
	}

	back.VolumeSize = buildCache.VolumeSize
	back.DisplayName = buildCache.DisplayName
	back.AccountName = ctx.AccountName

	return d.buildCacheRepo.UpdateById(ctx, id, back)
}

func (d *Impl) DeleteBuildCache(ctx RegistryContext, id repos.ID) error {
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
		return errors.Newf("unauthorized to delete build cache")
	}

	back, err := d.buildCacheRepo.FindOne(ctx, repos.Filter{
		"accountName": ctx.AccountName,
		"id":          id,
	})

	if err != nil {
		return err
	}

	i, err := d.buildRepo.Count(ctx, repos.Filter{
		"spec.accountName":  ctx.AccountName,
		"spec.cacheKeyName": back.Name,
	})
	if err != nil {
		return err
	}

	if i > 0 {
		return errors.Newf("build cache is in use, please delete all builds that use this cache first")
	}

	return d.buildCacheRepo.DeleteOne(ctx, repos.Filter{"accountName": ctx.AccountName, "id": id})
}

func (d *Impl) ListBuildCaches(ctx RegistryContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.BuildCacheKey], error) {
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
		return nil, errors.Newf("unauthorized to list build caches")
	}

	return d.buildCacheRepo.FindPaginated(ctx, d.buildCacheRepo.MergeMatchFilters(repos.Filter{"accountName": ctx.AccountName}, search), pagination)
}
