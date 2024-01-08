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

	service.EnsureGVK()

	if err := d.k8sClient.ValidateObject(ctx, &service.ProjectManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, service.ProjectName, &service.ProjectManagedService, 1); err != nil {
		return nil, errors.NewE(err)
	}

	pms, err := d.pmsRepo.Create(ctx, &service)
	if err != nil {
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

	cms, err := d.findProjectManagedService(ctx, projectName, service.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cms.IsMarkedForDeletion() {
		return nil, errors.Newf("cluster managed service %q (projectName=%q) is marked for deletion", service.Name, projectName)
	}

	cms.IncrementRecordVersion()
	cms.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	cms.Labels = service.Labels
	cms.Annotations = service.Annotations

	cms.SyncStatus = t.GenSyncStatus(t.SyncActionApply, cms.RecordVersion)

	unp, err := d.pmsRepo.UpdateById(ctx, cms.Id, cms)
	if err != nil {
		return nil, errors.NewE(err)
	}


	if err := d.applyK8sResource(ctx, projectName, &unp.ProjectManagedService, unp.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishProjectManagedServiceEvent(unp, PublishUpdate)

	return unp, nil
}

func (d *domain) DeleteProjectManagedService(ctx ConsoleContext, projectName string, name string) error {
	if err := d.canMutateResourcesInProject(ctx, projectName); err != nil {
		return errors.NewE(err)
	}

	svc, err := d.findProjectManagedService(ctx, projectName, name)
	if err != nil {
		return errors.NewE(err)
	}

	if svc.IsMarkedForDeletion() {
		return errors.Newf("project managed service %q (projectName=%q) is already marked for deletion", name, projectName)
	}

	svc.MarkedForDeletion = fn.New(true)
	svc.SyncStatus = t.GetSyncStatusForDeletion(svc.Generation)
	upC, err := d.pmsRepo.UpdateById(ctx, svc.Id, svc)
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

	a, err := d.findProjectManagedService(ctx, projectName,name)
	if err != nil {
		return errors.NewE(err)
	}

	return d.resyncK8sResource(ctx, a.ProjectName, a.SyncStatus.Action, &a.ProjectManagedService, a.RecordVersion)
}