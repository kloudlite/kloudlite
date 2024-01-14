package domain

import (
	"time"

	"github.com/kloudlite/api/apps/console/internal/entities"

	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/pkg/errors"
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

	if _, err := d.upsertProjectResourceMapping(ctx, projectName, pms); err != nil {
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

	patch := repos.Document{
		"recordVersion": pmsvc.RecordVersion + 1,
		"lastUpdatedBy": common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},

		"metadata.labels":      service.Labels,
		"metadata.annotations": service.Annotations,

		"syncStatus.syncScheduledAt": time.Now(),
		"syncStatus.action":          t.SyncActionApply,
		"syncStatus.state":           t.SyncStateInQueue,
	}

	upmsvc, err := d.pmsRepo.PatchById(ctx, pmsvc.Id, patch)
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

	if !pmsvc.IsMarkedForDeletion() {
		patch := repos.Document{
			"markedForDeletion":          true,
			"syncStatus.action":          t.SyncActionDelete,
			"syncStatus.syncScheduledAt": time.Now(),
			"syncStatus.state":           t.SyncStateInQueue,
		}

		upC, err := d.pmsRepo.PatchById(ctx, pmsvc.Id, patch)
		if err != nil {
			return errors.NewE(err)
		}

		d.resourceEventPublisher.PublishProjectManagedServiceEvent(upC, PublishUpdate)
	}

	return d.deleteK8sResource(ctx, projectName, &pmsvc.ProjectManagedService)
}

func (d *domain) OnProjectManagedServiceApplyError(ctx ConsoleContext, projectName, name, errMsg string, opts UpdateAndDeleteOpts) error {
	svc, err := d.findProjectManagedService(ctx, projectName, name)
	if err != nil {
		return errors.NewE(err)
	}

	patch := repos.Document{
		"syncStatus.state":        t.SyncStateErroredAtAgent,
		"syncStatus.lastSyncedAt": opts.MessageTimestamp,
		"syncStatus.error":        errMsg,
	}

	usvc, err := d.pmsRepo.PatchById(ctx, svc.Id, patch)
	if err != nil {
		return err
	}
	d.resourceEventPublisher.PublishProjectManagedServiceEvent(usvc, PublishUpdate)
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

	patch := repos.Document{
		"metadata.annotations":       service.Annotations,
		"metadata.labels":            service.Annotations,
		"metadata.generation":        service.Generation,
		"metadata.creationTimestamp": service.CreationTimestamp,

		"syncedOutputSecretRef": service.SyncedOutputSecretRef,

		"status": service.Status,
		"syncStatus.state": func() t.SyncState {
			if status == types.ResourceStatusDeleting {
				return t.SyncStateDeletingAtAgent
			}
			return t.SyncStateUpdatedAtAgent
		}(),
		"syncStatus.lastSyncedAt":  opts.MessageTimestamp,
		"syncStatus.error":         nil,
		"syncStatus.recordVersion": svc.RecordVersion,
	}

	if _, err := d.pmsRepo.PatchById(ctx, svc.Id, patch); err != nil {
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
