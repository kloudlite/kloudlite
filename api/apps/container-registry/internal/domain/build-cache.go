package domain

// import (
// 	"github.com/kloudlite/api/apps/container-registry/internal/domain/entities"
// 	fc "github.com/kloudlite/api/apps/container-registry/internal/domain/entities/field-constants"
// 	iamT "github.com/kloudlite/api/apps/iam/types"
// 	"github.com/kloudlite/api/common/fields"
// 	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
// 	"github.com/kloudlite/api/pkg/errors"
// 	"github.com/kloudlite/api/pkg/repos"
// )
//
// func (d *Impl) AddBuildCache(ctx RegistryContext, buildCache entities.BuildCacheKey) (*entities.BuildCacheKey, error) {
// 	co, err := d.iamClient.Can(ctx, &iam.CanIn{
// 		UserId: string(ctx.UserId),
// 		ResourceRefs: []string{
// 			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
// 		},
// 		Action: string(iamT.UpdateAccount),
// 	})
// 	if err != nil {
// 		return nil, errors.NewE(err)
// 	}
//
// 	if !co.Status {
// 		return nil, errors.Newf("unauthorized to add build cache")
// 	}
//
// 	buildCache.AccountName = ctx.AccountName
// 	buildCacheRepoCreated, err := d.buildCacheRepo.Create(ctx, &buildCache)
// 	if err != nil {
// 		return nil, errors.NewE(err)
// 	}
// 	d.resourceEventPublisher.PublishBuildCacheEvent(buildCacheRepoCreated, PublishAdd)
// 	return buildCacheRepoCreated, nil
// }
//
// func (d *Impl) UpdateBuildCache(ctx RegistryContext, id repos.ID, buildCache entities.BuildCacheKey) (*entities.BuildCacheKey, error) {
// 	co, err := d.iamClient.Can(ctx, &iam.CanIn{
// 		UserId: string(ctx.UserId),
// 		ResourceRefs: []string{
// 			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
// 		},
// 		Action: string(iamT.UpdateAccount),
// 	})
// 	if err != nil {
// 		return nil, errors.NewE(err)
// 	}
//
// 	if !co.Status {
// 		return nil, errors.Newf("unauthorized to update build cache")
// 	}
//
// 	buildCacheRepoUpdated, err := d.buildCacheRepo.PatchById(
// 		ctx,
// 		id,
// 		repos.Document{
// 			fc.BuildCacheKeyVolumeSizeInGB: buildCache.VolumeSize,
// 			fields.DisplayName:             buildCache.DisplayName,
// 			fields.AccountName:             ctx.AccountName,
// 		},
// 	)
//
// 	d.resourceEventPublisher.PublishBuildCacheEvent(buildCacheRepoUpdated, PublishUpdate)
// 	return buildCacheRepoUpdated, nil
// }
//
// func (d *Impl) DeleteBuildCache(ctx RegistryContext, id repos.ID) error {
// 	co, err := d.iamClient.Can(ctx, &iam.CanIn{
// 		UserId: string(ctx.UserId),
// 		ResourceRefs: []string{
// 			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
// 		},
// 		Action: string(iamT.UpdateAccount),
// 	})
// 	if err != nil {
// 		return errors.NewE(err)
// 	}
//
// 	if !co.Status {
// 		return errors.Newf("unauthorized to delete build cache")
// 	}
//
// 	back, err := d.buildCacheRepo.FindOne(ctx, repos.Filter{
// 		fields.AccountName: ctx.AccountName,
// 		fields.Id:          id,
// 	})
//
// 	if err != nil {
// 		return errors.NewE(err)
// 	}
//
// 	i, err := d.buildRepo.Count(ctx, repos.Filter{
// 		fc.BuildSpecAccountName:  ctx.AccountName,
// 		fc.BuildSpecCacheKeyName: back.Name,
// 	})
// 	if err != nil {
// 		return errors.NewE(err)
// 	}
//
// 	if i > 0 {
// 		return errors.Newf("build cache is in use, please delete all builds that use this cache first")
// 	}
//
// 	err = d.buildCacheRepo.DeleteOne(ctx, repos.Filter{"accountName": ctx.AccountName, "id": id})
// 	if err != nil {
// 		return errors.NewE(err)
// 	}
// 	d.resourceEventPublisher.PublishBuildCacheEvent(back, PublishDelete)
// 	return nil
// }
//
// func (d *Impl) ListBuildCaches(ctx RegistryContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.BuildCacheKey], error) {
// 	co, err := d.iamClient.Can(ctx, &iam.CanIn{
// 		UserId: string(ctx.UserId),
// 		ResourceRefs: []string{
// 			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
// 		},
// 		Action: string(iamT.GetAccount),
// 	})
// 	if err != nil {
// 		return nil, errors.NewE(err)
// 	}
//
// 	if !co.Status {
// 		return nil, errors.Newf("unauthorized to list build caches")
// 	}
//
// 	return d.buildCacheRepo.FindPaginated(ctx, d.buildCacheRepo.MergeMatchFilters(repos.Filter{"accountName": ctx.AccountName}, search), pagination)
// }
