package domain

import (
	"fmt"
	"time"

	"github.com/kloudlite/api/apps/console/internal/entities"

	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
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
		fc.ProjectName: projectName,
		fc.AccountName: ctx.AccountName,
	}

	pr, err := d.pmsRepo.FindPaginated(ctx, d.secretRepo.MergeMatchFilters(f, mf), pagination)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return pr, nil
}

func (d *domain) findProjectManagedService(ctx ConsoleContext, projectName string, svcName string) (*entities.ProjectManagedService, error) {
	pmsvc, err := d.pmsRepo.FindOne(ctx, repos.Filter{
		fc.ProjectName:  projectName,
		fc.AccountName:  ctx.AccountName,
		fc.MetadataName: svcName,
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

	service.Namespace = d.getProjectNamespace(projectName)
	service.IncrementRecordVersion()

	service.Spec.TargetNamespace = fmt.Sprintf("%s-pmsvc-%s", projectName, service.Name)

	service.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	service.LastUpdatedBy = service.CreatedBy

	existing, err := d.pmsRepo.FindOne(ctx, repos.Filter{
		fc.ProjectName:  projectName,
		fc.AccountName:  ctx.AccountName,
		fc.MetadataName: service.Name,
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

	pmsvc, err := d.findProjectManagedService(ctx, projectName, service.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if pmsvc == nil {
		return nil, errors.Newf("no project manage service found")
	}

	service.Namespace = pmsvc.Namespace
	service.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &service); err != nil {
		return nil, errors.NewE(err)
	}

	if pmsvc.IsMarkedForDeletion() {
		return nil, errors.Newf("cluster managed service %q (projectName=%q) is marked for deletion", service.Name, projectName)
	}

	upmsvc, err := d.pmsRepo.PatchById(ctx, pmsvc.Id, repos.Document{
		fc.RecordVersion: pmsvc.RecordVersion + 1,
		fc.LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},

		fc.MetadataLabels:      service.Labels,
		fc.MetadataAnnotations: service.Annotations,

		fc.SyncStatusSyncScheduledAt: time.Now(),
		fc.SyncStatusAction:          t.SyncActionApply,
		fc.SyncStatusState:           t.SyncStateInQueue,
	})
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

	if pmsvc == nil {
		return errors.Newf("no project manage service found")
	}

	if !pmsvc.IsMarkedForDeletion() {

		upC, err := d.pmsRepo.PatchById(ctx, pmsvc.Id, repos.Document{
			fc.MarkedForDeletion:         true,
			fc.SyncStatusAction:          t.SyncActionDelete,
			fc.SyncStatusSyncScheduledAt: time.Now(),
			fc.SyncStatusState:           t.SyncStateInQueue,
		})
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

	if svc == nil {
		return errors.Newf("no project manage service found")
	}

	usvc, err := d.pmsRepo.PatchById(ctx, svc.Id, repos.Document{
		fc.SyncStatusState:        t.SyncStateErroredAtAgent,
		fc.SyncStatusLastSyncedAt: opts.MessageTimestamp,
		fc.SyncStatusError:        errMsg,
	})
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

	if svc == nil {
		return errors.Newf("no project manage service found")
	}

	if err := d.MatchRecordVersion(service.Annotations, svc.RecordVersion); err != nil {
		return d.ResyncProjectManagedService(ctx, service.ProjectName, service.Name)
	}

	if _, err := d.pmsRepo.PatchById(ctx, svc.Id, repos.Document{
		fc.MetadataAnnotations:       service.Annotations,
		fc.MetadataLabels:            service.Annotations,
		fc.MetadataGeneration:        service.Generation,
		fc.MetadataCreationTimestamp: service.CreationTimestamp,

		fc.ProjectManagedServiceSyncedOutputSecretRef: service.SyncedOutputSecretRef,

		fc.Status: service.Status,
		fc.SyncStatusState: func() t.SyncState {
			if status == types.ResourceStatusDeleting {
				return t.SyncStateDeletingAtAgent
			}
			return t.SyncStateUpdatedAtAgent
		}(),
		fc.SyncStatusLastSyncedAt:  opts.MessageTimestamp,
		fc.SyncStatusError:         nil,
		fc.SyncStatusRecordVersion: svc.RecordVersion,
	}); err != nil {
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
