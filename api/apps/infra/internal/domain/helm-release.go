package domain

import (
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"time"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *domain) findHelmRelease(ctx InfraContext, clusterName string, hrName string) (*entities.HelmRelease, error) {
	cluster, err := d.helmReleaseRepo.FindOne(ctx, repos.Filter{
		"clusterName":   clusterName,
		"accountName":   ctx.AccountName,
		"metadata.name": hrName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cluster == nil {
		return nil, errors.Newf("helm release with name %q not found", hrName)
	}
	return cluster, nil
}

func (d *domain) ListHelmReleases(ctx InfraContext, clusterName string, mf map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.HelmRelease], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListHelmReleases); err != nil {
		return nil, errors.NewE(err)
	}

	f := repos.Filter{
		"clusterName": clusterName,
		"accountName": ctx.AccountName,
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
	return d.resDispatcher.ApplyToTargetCluster(ctx, hr.ClusterName, &hr.HelmChart, hr.RecordVersion)
}

func (d *domain) CreateHelmRelease(ctx InfraContext, clusterName string, hr entities.HelmRelease) (*entities.HelmRelease, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateHelmRelease); err != nil {
		return nil, errors.NewE(err)
	}
	hr.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &hr.HelmChart); err != nil {
		return nil, errors.NewE(err)
	}

	hr.IncrementRecordVersion()
	hr.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	hr.LastUpdatedBy = hr.CreatedBy

	existing, err := d.helmReleaseRepo.FindOne(ctx, repos.Filter{
		"clusterName":   clusterName,
		"accountName":   ctx.AccountName,
		"metadata.name": hr.Name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if existing != nil {
		return nil, errors.Newf("helm release with name %q already exists", hr.Name)
	}

	hr.AccountName = ctx.AccountName
	hr.ClusterName = clusterName
	hr.SyncStatus = t.GenSyncStatus(t.SyncActionApply, hr.RecordVersion)

	cms, err := d.helmReleaseRepo.Create(ctx, &hr)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishHelmReleaseEvent(&hr, PublishAdd)

	if err = d.resDispatcher.ApplyToTargetCluster(ctx, clusterName, &corev1.Namespace{
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

	if err := d.applyHelmRelease(ctx, &hr); err != nil {
		return nil, errors.NewE(err)
	}

	return cms, nil
}

func (d *domain) UpdateHelmRelease(ctx InfraContext, clusterName string, hrIn entities.HelmRelease) (*entities.HelmRelease, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateHelmRelease); err != nil {
		return nil, errors.NewE(err)
	}

	hrIn.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &hrIn); err != nil {
		return nil, errors.NewE(err)
	}

	cms, err := d.findHelmRelease(ctx, clusterName, hrIn.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cms.IsMarkedForDeletion() {
		return nil, errors.Newf("helm release with name %q is marked for deletion", hrIn.Name)
	}

	unp, err := d.helmReleaseRepo.PatchById(ctx, cms.Id, repos.Document{
		"metadata.labels":      hrIn.Labels,
		"metadata.annotations": hrIn.Annotations,
		"displayName":          hrIn.DisplayName,
		"recordVersion":        hrIn.RecordVersion + 1,
		"spec.chartVersion":    hrIn.Spec.ChartVersion,
		"spec.values":          hrIn.Spec.Values,
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
		"syncStatus.lastSyncedAt": time.Now(),
		"syncStatus.action":       t.SyncActionApply,
		"syncStatus.state":        t.SyncStateInQueue,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishHelmReleaseEvent(unp, PublishUpdate)
	if err := d.applyHelmRelease(ctx, unp); err != nil {
		return nil, errors.NewE(err)
	}
	return unp, nil
}

func (d *domain) DeleteHelmRelease(ctx InfraContext, clusterName string, name string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteHelmRelease); err != nil {
		return errors.NewE(err)
	}

	svc, err := d.findHelmRelease(ctx, clusterName, name)
	if err != nil {
		return errors.NewE(err)
	}

	if svc.IsMarkedForDeletion() {
		return errors.Newf("helm release with name %q is marked for deletion", name)
	}

	upC, err := d.helmReleaseRepo.PatchById(ctx, svc.Id, repos.Document{
		"markedForDeletion": true,
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
		"syncStatus.lastSyncedAt": time.Now(),
		"syncStatus.action":       t.SyncActionDelete,
		"syncStatus.state":        t.SyncStateInQueue,
	})
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishHelmReleaseEvent(upC, PublishUpdate)

	return d.resDispatcher.DeleteFromTargetCluster(ctx, clusterName, &upC.HelmChart)
}

func (d *domain) OnHelmReleaseApplyError(ctx InfraContext, clusterName string, name string, errMsg string, opts UpdateAndDeleteOpts) error {
	svc, err := d.findHelmRelease(ctx, clusterName, name)
	if err != nil {
		return errors.NewE(err)
	}

	_, err = d.helmReleaseRepo.PatchById(ctx, svc.Id, repos.Document{
		"syncStatus.state":        t.SyncStateErroredAtAgent,
		"syncStatus.lastSyncedAt": opts.MessageTimestamp,
		"syncStatus.error":        &errMsg,
	})
	d.resourceEventPublisher.PublishHelmReleaseEvent(svc, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnHelmReleaseDeleteMessage(ctx InfraContext, clusterName string, hr entities.HelmRelease) error {
	xhr, err := d.findHelmRelease(ctx, clusterName, hr.Name)
	if err != nil {
		return err
	}
	if xhr == nil {
		// does not exist, (maybe already deleted)
		return nil
	}

	if err = d.helmReleaseRepo.DeleteById(ctx, xhr.Id); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishHelmReleaseEvent(xhr, PublishDelete)
	return err
}

func (d *domain) OnHelmReleaseUpdateMessage(ctx InfraContext, clusterName string, hr entities.HelmRelease, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xhr, err := d.findHelmRelease(ctx, clusterName, hr.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.matchRecordVersion(hr.Annotations, xhr.RecordVersion); err != nil {
		return d.resyncToTargetCluster(ctx, xhr.SyncStatus.Action, clusterName, xhr, xhr.RecordVersion)
	}

	// Ignore error if annotation don't have record version
	annVersion, _ := d.parseRecordVersionFromAnnotations(hr.Annotations)

	if _, err := d.helmReleaseRepo.PatchById(ctx, xhr.Id, repos.Document{
		"metadata.labels":            hr.Labels,
		"metadata.annotations":       hr.Annotations,
		"metadata.generation":        hr.Generation,
		"metadata.creationTimestamp": hr.CreationTimestamp,
		"status":                     hr.Status,
		"syncStatus": t.SyncStatus{
			LastSyncedAt:  opts.MessageTimestamp,
			Error:         nil,
			Action:        t.SyncActionApply,
			RecordVersion: annVersion,
			State: func() t.SyncState {
				if status == types.ResourceStatusDeleting {
					return t.SyncStateDeletingAtAgent
				}
				return t.SyncStateUpdatedAtAgent
			}(),
		},
	}); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishHelmReleaseEvent(xhr, PublishUpdate)
	return nil
}
