package domain

import (
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
)

// TODO: needs to chain cluster in cluster group
func (d *domain) CreateClusterGroup(ctx InfraContext, cg entities.ClusterGroup) (*entities.ClusterGroup, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateCluster); err != nil {
		return nil, errors.NewE(err)
	}
	// TODO: validate name

	existing, err := d.clusterRepo.FindOne(ctx, repos.Filter{
		fields.MetadataName: cg.Name,
		fields.AccountName:  ctx.AccountName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if existing != nil {
		return nil, ErrClusterAlreadyExists{ClusterName: cg.Name, AccountName: ctx.AccountName}
	}

	cg.AccountName = ctx.AccountName

	cg.IncrementRecordVersion()
	cg.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	cg.LastUpdatedBy = cg.CreatedBy

	cg.SyncStatus = t.GenSyncStatus(t.SyncActionApply, 0)

	ncg, err := d.clusterGroupRepo.Create(ctx, &cg)
	if err != nil {
		if d.clusterRepo.ErrAlreadyExists(err) {
			return nil, errors.Newf("cluster group with name %q already exists in account %q", cg.Name, cg.AccountName)
		}
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishInfraEvent(ctx, ResourceTypeClusterGroup, ncg.Name, PublishAdd)

	return ncg, nil
}

func (d *domain) UpdateClusterGroup(ctx InfraContext, cgIn entities.ClusterGroup) (*entities.ClusterGroup, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateCluster); err != nil {
		return nil, errors.NewE(err)
	}

	ucg, err := d.clusterGroupRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: cgIn.Name,
		},
		common.PatchForUpdate(ctx, &cgIn),
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishInfraEvent(ctx, ResourceTypeClusterGroup, ucg.Name, PublishUpdate)
	return ucg, nil
}

func (d *domain) DeleteClusterGroup(ctx InfraContext, name string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteCluster); err != nil {
		return errors.NewE(err)
	}

	filter := repos.Filter{
		fields.AccountName:         ctx.AccountName,
		fc.ClusterClusterGroupName: name,
	}

	cCount, err := d.clusterRepo.Count(ctx, filter)
	if err != nil {
		return errors.NewE(err)
	}
	if cCount != 0 {
		return errors.Newf("delete clusters first, aborting cluster group deletion")
	}

	ucg, err := d.clusterGroupRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: name,
		},
		common.PatchForMarkDeletion(),
	)
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishInfraEvent(ctx, ResourceTypeClusterGroup, ucg.Name, PublishUpdate)
	return nil
}

func (d *domain) ListClustersGroup(ctx InfraContext, mf map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ClusterGroup], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListClusters); err != nil {
		return nil, errors.NewE(err)
	}

	f := repos.Filter{
		fields.AccountName: ctx.AccountName,
	}

	pr, err := d.clusterGroupRepo.FindPaginated(ctx, d.clusterGroupRepo.MergeMatchFilters(f, mf), pagination)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return pr, nil
}

func (d *domain) GetClusterGroup(ctx InfraContext, name string) (*entities.ClusterGroup, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetCluster); err != nil {
		return nil, errors.NewE(err)
	}

	c, err := d.findClusterGroup(ctx, name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return c, nil
}

func (d *domain) findClusterGroup(ctx InfraContext, cgName string) (*entities.ClusterGroup, error) {
	cg, err := d.clusterGroupRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: cgName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cg == nil {
		return nil, ErrClusterNotFound
	}
	return cg, nil
}
