package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
)

func (d *domain) ListProjectManagedServices(ctx ConsoleContext, projectName string, mf map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ProjectManagedService], error) {
	if err := d.canReadResourcesInProject(ctx, projectName); err != nil {
		return nil, errors.NewE(err)
	}

	f := repos.Filter{
		"projectName": projectName,
		"accountName": ctx.AccountName,
	}

	pr, err := d.pmsRepo.FindPaginated(ctx, d.secretRepo.MergeMatchFilters(f, mf), pagination)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return pr, nil
}

func (d *domain) findProjectManagedService(ctx ConsoleContext, projectName string, svcName string) (*entities.ProjectManagedService, error) {
	pmsvc, err := d.pmsRepo.FindOne(ctx, repos.Filter{
		"projectName":   projectName,
		"accountName":   ctx.AccountName,
		"metadata.name": svcName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if pmsvc == nil {
		return nil, errors.Newf("cmsvc with name %q not found", projectName)
	}
	return pmsvc, nil
}

func (d *domain) GetProjectManagedService(ctx ConsoleContext, projectName string, serviceName string) (*entities.ProjectManagedService, error) {
	if err := d.canReadResourcesInProject(ctx, projectName); err != nil {
		return nil, errors.NewE(err)
	}

	c, err := d.findProjectManagedService(ctx, projectName, serviceName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return c, nil
}

func (d *domain) applyProjectManagedService(ctx ConsoleContext, pmsvc *entities.ProjectManagedService) error {
	addTrackingId(&pmsvc.ProjectManagedService, pmsvc.Id)
	return d.applyK8sResource(ctx, pmsvc.ProjectName, &pmsvc.ProjectManagedService, pmsvc.RecordVersion)
}

func (d *domain) CreateProjectManagedService(ctx ConsoleContext, projectName string, service entities.ProjectManagedService) (*entities.ProjectManagedService, error) {
	if err := d.canMutateResourcesInProject(ctx, projectName); err != nil {
		return nil, errors.NewE(err)
	}

	service.IncrementRecordVersion()

	service.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	service.LastUpdatedBy = service.CreatedBy

	existing, err := d.pmsRepo.FindOne(ctx, repos.Filter{
		"projectName":   projectName,
		"accountName":   ctx.AccountName,
		"metadata.name": service.Name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if existing != nil {
		return nil, errors.Newf("project managed service with name %q already exists", projectName)
	}

	service.AccountName = ctx.AccountName
	service.ProjectName = projectName
	service.SyncStatus = t.GenSyncStatus(t.SyncActionApply, service.RecordVersion)
	service.Namespace = d.getProjectNamespace(projectName)

	service.EnsureGVK()

	if err := d.k8sClient.ValidateObject(ctx, &service.ProjectManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	pms, err := d.pmsRepo.Create(ctx, &service)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyProjectManagedService(ctx, pms); err != nil {
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishProjectManagedServiceEvent(&service, PublishAdd)
	return pms, nil
}

func (d *domain) UpdateProjectManagedService(ctx ConsoleContext, projectName string, service entities.ProjectManagedService) (*entities.ProjectManagedService, error) {
	if err := d.canMutateResourcesInProject(ctx, projectName); err != nil {
		return nil, errors.NewE(err)
	}

	service.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &service); err != nil {
		return nil, errors.NewE(err)
	}

	pmsvc, err := d.findProjectManagedService(ctx, projectName, service.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if pmsvc.IsMarkedForDeletion() {
		return nil, errors.Newf("cluster managed service %q (projectName=%q) is marked for deletion", service.Name, projectName)
	}

	pmsvc.IncrementRecordVersion()
	pmsvc.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	pmsvc.Labels = service.Labels
	pmsvc.Annotations = service.Annotations

	pmsvc.SyncStatus = t.GenSyncStatus(t.SyncActionApply, pmsvc.RecordVersion)

	upmsvc, err := d.pmsRepo.UpdateById(ctx, pmsvc.Id, pmsvc)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyProjectManagedService(ctx, upmsvc); err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishProjectManagedServiceEvent(upmsvc, PublishUpdate)

	return upmsvc, nil
}

func (d *domain) DeleteProjectManagedService(ctx ConsoleContext, projectName string, name string) error {
	if err := d.canMutateResourcesInProject(ctx, projectName); err != nil {
		return errors.NewE(err)
	}

	pmsvc, err := d.findProjectManagedService(ctx, projectName, name)
	if err != nil {
		return errors.NewE(err)
	}

	if pmsvc.IsMarkedForDeletion() {
		return errors.Newf("project managed service %q (projectName=%q) is already marked for deletion", name, projectName)
	}

	pmsvc.MarkedForDeletion = fn.New(true)
	pmsvc.SyncStatus = t.GetSyncStatusForDeletion(pmsvc.Generation)
	upC, err := d.pmsRepo.UpdateById(ctx, pmsvc.Id, pmsvc)
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishProjectManagedServiceEvent(upC, PublishUpdate)

	return d.deleteK8sResource(ctx, projectName, &upC.ProjectManagedService)
}

func (d *domain) OnProjectManagedServiceApplyError(ctx ConsoleContext, projectName, name, errMsg string, opts UpdateAndDeleteOpts) error {
	svc, err := d.findProjectManagedService(ctx, projectName, name)
	if err != nil {
		return errors.NewE(err)
	}

	svc.SyncStatus.State = t.SyncStateErroredAtAgent
	svc.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	svc.SyncStatus.Error = &errMsg

	_, err = d.pmsRepo.UpdateById(ctx, svc.Id, svc)
	d.resourceEventPublisher.PublishProjectManagedServiceEvent(svc, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) OnProjectManagedServiceDeleteMessage(ctx ConsoleContext, projectName string, service entities.ProjectManagedService) error {
	svc, err := d.findProjectManagedService(ctx, projectName, service.Name)
	if err != nil {
		return err
	}

	if svc == nil {
		// does not exist, (maybe already deleted)
		return nil
	}

	if err := d.MatchRecordVersion(service.Annotations, svc.RecordVersion); err != nil {
		return d.ResyncProjectManagedService(ctx, service.ProjectName, service.Name)
	}

	if err := d.pmsRepo.DeleteById(ctx, svc.Id); err != nil {
		return err
	}
	d.resourceEventPublisher.PublishProjectManagedServiceEvent(svc, PublishDelete)
	return err
}

func (d *domain) OnProjectManagedServiceUpdateMessage(ctx ConsoleContext, projectName string, service entities.ProjectManagedService, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	svc, err := d.findProjectManagedService(ctx, projectName, service.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(service.Annotations, svc.RecordVersion); err != nil {
		return d.ResyncProjectManagedService(ctx, service.ProjectName, service.Name)
	}

	svc.Status = service.Status

	svc.SyncStatus.State = func() t.SyncState {
		if status == types.ResourceStatusDeleting {
			return t.SyncStateDeletingAtAgent
		}
		return t.SyncStateUpdatedAtAgent
	}()
	svc.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	svc.SyncStatus.Error = nil
	svc.SyncStatus.RecordVersion = svc.RecordVersion

	if _, err := d.pmsRepo.UpdateById(ctx, svc.Id, svc); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishProjectManagedServiceEvent(svc, PublishUpdate)
	return nil
}

func (d *domain) ResyncProjectManagedService(ctx ConsoleContext, projectName, name string) error {
	if err := d.canMutateResourcesInProject(ctx, projectName); err != nil {
		return errors.NewE(err)
	}

	a, err := d.findProjectManagedService(ctx, projectName, name)
	if err != nil {
		return errors.NewE(err)
	}

	return d.resyncK8sResource(ctx, a.ProjectName, a.SyncStatus.Action, &a.ProjectManagedService, a.RecordVersion)
}

