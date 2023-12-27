package domain

import (
	"github.com/kloudlite/api/pkg/errors"
	"time"

	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
)

func (d *domain) ListManagedServices(ctx ConsoleContext, namespace string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.ManagedService], error) {
	if err := d.canReadResourcesInWorkspace(ctx, namespace); err != nil {
		return nil, errors.NewE(err)
	}
	filter := repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
	}

	return d.msvcRepo.FindPaginated(ctx, d.msvcRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) findMSvc(ctx ConsoleContext, namespace string, name string) (*entities.ManagedService, error) {
	msvc, err := d.msvcRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
		"metadata.name":      name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if msvc == nil {
		return nil, errors.Newf("no secret with name=%q,namespace=%q found", name, namespace)
	}
	return msvc, nil
}

func (d *domain) GetManagedService(ctx ConsoleContext, namespace string, name string) (*entities.ManagedService, error) {
	if err := d.canReadResourcesInWorkspace(ctx, namespace); err != nil {
		return nil, errors.NewE(err)
	}
	return d.findMSvc(ctx, namespace, name)
}

// mutations

func (d *domain) CreateManagedService(ctx ConsoleContext, msvc entities.ManagedService) (*entities.ManagedService, error) {
	if err := d.canMutateResourcesInWorkspace(ctx, msvc.Namespace); err != nil {
		return nil, errors.NewE(err)
	}

	msvc.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &msvc.ManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	msvc.IncrementRecordVersion()

	msvc.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	msvc.LastUpdatedBy = msvc.CreatedBy

	msvc.AccountName = ctx.AccountName
	msvc.ClusterName = ctx.ClusterName
	msvc.SyncStatus = t.GenSyncStatus(t.SyncActionApply, msvc.RecordVersion)

	m, err := d.msvcRepo.Create(ctx, &msvc)
	if err != nil {
		if d.msvcRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishMsvcEvent(&msvc, PublishAdd)

	if err := d.applyK8sResource(ctx, &m.ManagedService, m.RecordVersion); err != nil {
		return m, errors.NewE(err)
	}

	return m, nil
}

func (d *domain) UpdateManagedService(ctx ConsoleContext, msvc entities.ManagedService) (*entities.ManagedService, error) {
	if err := d.canMutateResourcesInWorkspace(ctx, msvc.Namespace); err != nil {
		return nil, errors.NewE(err)
	}

	msvc.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &msvc.ManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	m, err := d.findMSvc(ctx, msvc.Namespace, msvc.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	m.IncrementRecordVersion()
	m.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	m.DisplayName = msvc.DisplayName

	m.Annotations = msvc.Annotations
	m.Labels = msvc.Labels

	m.Spec = msvc.Spec
	m.SyncStatus = t.GenSyncStatus(t.SyncActionApply, m.RecordVersion)

	upMSvc, err := d.msvcRepo.UpdateById(ctx, m.Id, m)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishMsvcEvent(upMSvc, PublishUpdate)

	if err := d.applyK8sResource(ctx, &upMSvc.ManagedService, upMSvc.RecordVersion); err != nil {
		return upMSvc, errors.NewE(err)
	}

	return upMSvc, nil
}

func (d *domain) DeleteManagedService(ctx ConsoleContext, namespace string, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return errors.NewE(err)
	}
	m, err := d.findMSvc(ctx, namespace, name)
	if err != nil {
		return errors.NewE(err)
	}

	m.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, m.RecordVersion)
	if _, err := d.msvcRepo.UpdateById(ctx, m.Id, m); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishMsvcEvent(m, PublishUpdate)

	return d.deleteK8sResource(ctx, &m.ManagedService)
}

func (d *domain) OnDeleteManagedServiceMessage(ctx ConsoleContext, msvc entities.ManagedService) error {
	exMsvc, err := d.findMSvc(ctx, msvc.Namespace, msvc.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(msvc.Annotations, exMsvc.RecordVersion); err != nil {
		return d.resyncK8sResource(ctx, exMsvc.SyncStatus.Action, &exMsvc.ManagedService, exMsvc.RecordVersion)
	}

	err = d.msvcRepo.DeleteById(ctx, exMsvc.Id)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishMsvcEvent(exMsvc, PublishDelete)
	return nil
}

func (d *domain) OnUpdateManagedServiceMessage(ctx ConsoleContext, msvc entities.ManagedService) error {
	exMsvc, err := d.findMSvc(ctx, msvc.Namespace, msvc.Name)
	if err != nil {
		return errors.NewE(err)
	}

	annotatedVersion, err := d.parseRecordVersionFromAnnotations(msvc.Annotations)
	if err != nil {
		return d.resyncK8sResource(ctx, exMsvc.SyncStatus.Action, &exMsvc.ManagedService, exMsvc.RecordVersion)
	}

	if annotatedVersion != exMsvc.RecordVersion {
		return d.resyncK8sResource(ctx, exMsvc.SyncStatus.Action, &exMsvc.ManagedService, exMsvc.RecordVersion)
	}

	exMsvc.CreationTimestamp = msvc.CreationTimestamp
	exMsvc.Labels = msvc.Labels
	exMsvc.Annotations = msvc.Annotations
	exMsvc.Generation = msvc.Generation

	exMsvc.Status = msvc.Status

	exMsvc.SyncStatus.State = t.SyncStateReceivedUpdateFromAgent
	exMsvc.SyncStatus.RecordVersion = annotatedVersion
	exMsvc.SyncStatus.Error = nil
	exMsvc.SyncStatus.LastSyncedAt = time.Now()

	_, err = d.msvcRepo.UpdateById(ctx, exMsvc.Id, exMsvc)
	d.resourceEventPublisher.PublishMsvcEvent(exMsvc, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnApplyManagedServiceError(ctx ConsoleContext, errMsg string, namespace string, name string) error {
	m, err2 := d.findMSvc(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	m.SyncStatus.State = t.SyncStateErroredAtAgent
	m.SyncStatus.LastSyncedAt = time.Now()
	m.SyncStatus.Error = &errMsg
	_, err := d.msvcRepo.UpdateById(ctx, m.Id, m)
	return errors.NewE(err)
}

func (d *domain) ResyncManagedService(ctx ConsoleContext, namespace, name string) error {
	if err := d.canMutateResourcesInWorkspace(ctx, namespace); err != nil {
		return errors.NewE(err)
	}

	c, err := d.findMSvc(ctx, namespace, name)
	if err != nil {
		return errors.NewE(err)
	}

	return d.resyncK8sResource(ctx, c.SyncStatus.Action, &c.ManagedService, c.RecordVersion)
}
