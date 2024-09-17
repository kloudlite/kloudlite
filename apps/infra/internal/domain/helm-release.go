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
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *domain) findHelmRelease(ctx InfraContext, clusterName string, hrName string) (*entities.HelmRelease, error) {
	cluster, err := d.helmReleaseRepo.FindOne(ctx, repos.Filter{
		fields.ClusterName:  clusterName,
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: hrName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cluster == nil {
		return nil, errors.Newf("helm release with name %q not found", hrName)
	}
	return cluster, nil
}

func (d *domain) upsertHelmRelease(ctx InfraContext, clusterName string, hr *entities.HelmRelease) (*entities.HelmRelease, error) {
	cluster, err := d.helmReleaseRepo.Upsert(ctx, repos.Filter{
		fields.ClusterName:  clusterName,
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: hr.Name,
	}, hr)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cluster == nil {
		return nil, errors.Newf("could not upsert helm release %s", hr.Name)
	}
	return cluster, nil
}

func (d *domain) ListHelmReleases(ctx InfraContext, clusterName string, mf map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.HelmRelease], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListHelmReleases); err != nil {
		return nil, errors.NewE(err)
	}

	f := repos.Filter{
		fields.ClusterName: clusterName,
		fields.AccountName: ctx.AccountName,
	}

	pr, err := d.helmReleaseRepo.FindPaginated(ctx, d.helmReleaseRepo.MergeMatchFilters(f, mf), pagination)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return pr, nil
}

func (d *domain) GetHelmRelease(ctx InfraContext, clusterName string, hrName string) (*entities.HelmRelease, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetHelmRelease); err != nil {
		return nil, errors.NewE(err)
	}

	c, err := d.GetHelmRelease(ctx, clusterName, hrName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return c, nil
}

func (d *domain) applyHelmRelease(ctx InfraContext, hr *entities.HelmRelease) error {
	addTrackingId(&hr.HelmChart, hr.Id)
	return d.resDispatcher.ApplyToTargetCluster(ctx, hr.DispatchAddr, &hr.HelmChart, hr.RecordVersion)
}

func (d *domain) CreateHelmRelease(ctx InfraContext, clusterName string, hr entities.HelmRelease) (*entities.HelmRelease, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateHelmRelease); err != nil {
		return nil, errors.NewE(err)
	}
	hr.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &hr.HelmChart); err != nil {
		return nil, errors.NewE(err)
	}

	if hr.DispatchAddr == nil {
		hr.DispatchAddr = &entities.DispatchAddr{AccountName: ctx.AccountName, ClusterName: clusterName}
	}

	hr.IncrementRecordVersion()
	hr.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	hr.LastUpdatedBy = hr.CreatedBy

	existing, err := d.helmReleaseRepo.FindOne(
		ctx,
		repos.Filter{
			fields.ClusterName:  clusterName,
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: hr.Name,
		},
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if existing != nil {
		return nil, errors.Newf("helm release with name %q already exists", hr.Name)
	}

	hr.AccountName = ctx.AccountName
	hr.ClusterName = clusterName
	hr.SyncStatus = t.GenSyncStatus(t.SyncActionApply, hr.RecordVersion)

	nhr, err := d.helmReleaseRepo.Create(ctx, &hr)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, nhr.ClusterName, ResourceTypeHelmRelease, nhr.Name, PublishAdd)

	if err = d.resDispatcher.ApplyToTargetCluster(ctx, hr.DispatchAddr, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: hr.Namespace,
		},
	}, hr.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyHelmRelease(ctx, nhr); err != nil {
		return nil, errors.NewE(err)
	}

	return nhr, nil
}

func (d *domain) UpdateHelmRelease(ctx InfraContext, clusterName string, hrIn entities.HelmRelease) (*entities.HelmRelease, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateHelmRelease); err != nil {
		return nil, errors.NewE(err)
	}

	hrIn.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &hrIn); err != nil {
		return nil, errors.NewE(err)
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&hrIn,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.HelmReleaseSpecChartVersion: hrIn.Spec.ChartVersion,
				fc.HelmReleaseSpecValues:       hrIn.Spec.Values,
			},
		})

	uphr, err := d.helmReleaseRepo.Patch(
		ctx,
		repos.Filter{
			fields.ClusterName:  clusterName,
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: hrIn.Name,
		},
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, uphr.ClusterName, ResourceTypeHelmRelease, uphr.Name, PublishUpdate)
	if err := d.applyHelmRelease(ctx, uphr); err != nil {
		return nil, errors.NewE(err)
	}
	return uphr, nil
}

func (d *domain) DeleteHelmRelease(ctx InfraContext, clusterName string, name string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteHelmRelease); err != nil {
		return errors.NewE(err)
	}

	uphr, err := d.helmReleaseRepo.Patch(
		ctx,
		repos.Filter{
			fields.ClusterName:  clusterName,
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: name,
		},
		common.PatchForMarkDeletion(),
	)
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, uphr.ClusterName, ResourceTypeHelmRelease, uphr.Name, PublishUpdate)

	return d.resDispatcher.DeleteFromTargetCluster(ctx, uphr.DispatchAddr, &uphr.HelmChart)
}

func (d *domain) OnHelmReleaseApplyError(ctx InfraContext, clusterName string, name string, errMsg string, opts UpdateAndDeleteOpts) error {
	uphr, err := d.helmReleaseRepo.Patch(
		ctx,
		repos.Filter{
			fields.ClusterName:  clusterName,
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: name,
		},
		common.PatchForErrorFromAgent(
			errMsg,
			common.PatchOpts{
				MessageTimestamp: opts.MessageTimestamp,
			},
		),
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, uphr.ClusterName, ResourceTypeHelmRelease, uphr.Name, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnHelmReleaseDeleteMessage(ctx InfraContext, clusterName string, hr entities.HelmRelease) error {
	err := d.helmReleaseRepo.DeleteOne(
		ctx,
		repos.Filter{
			fields.ClusterName:  clusterName,
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: hr.Name,
		},
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeHelmRelease, hr.Name, PublishDelete)
	return err
}

func (d *domain) OnHelmReleaseUpdateMessage(ctx InfraContext, clusterName string, hr entities.HelmRelease, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xhr, err := d.findHelmRelease(ctx, clusterName, hr.Name)
	if err != nil {
		return errors.NewE(err)
	}

	recordVersion, err := d.matchRecordVersion(hr.Annotations, xhr.RecordVersion)
	if err != nil {
		return d.resyncToTargetCluster(ctx, xhr.SyncStatus.Action, xhr.DispatchAddr, xhr, xhr.RecordVersion)
	}

	uphr, err := d.helmReleaseRepo.PatchById(
		ctx,
		xhr.Id,
		common.PatchForSyncFromAgent(&hr, recordVersion, status, common.PatchOpts{
			MessageTimestamp: opts.MessageTimestamp,
		}))
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, uphr.ClusterName, ResourceTypeHelmRelease, uphr.GetName(), PublishUpdate)
	return nil
}
