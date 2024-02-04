package domain

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/common/fields"

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
		fields.ProjectName: projectName,
		fields.AccountName: ctx.AccountName,
	}

	pr, err := d.pmsRepo.FindPaginated(ctx, d.secretRepo.MergeMatchFilters(f, mf), pagination)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return pr, nil
}

func (d *domain) findProjectManagedService(ctx ConsoleContext, projectName string, svcName string) (*entities.ProjectManagedService, error) {
	pmsvc, err := d.pmsRepo.FindOne(ctx, repos.Filter{
		fields.ProjectName:  projectName,
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: svcName,
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

	service.Namespace = d.getProjectTargetNamespace(projectName)
	service.IncrementRecordVersion()

	if service.Spec.TargetNamespace == "" {
		service.Spec.TargetNamespace = d.getPMSTargetNamespace(projectName, service.Name, entities.ResourceTypeProjectManagedService)
	}

	service.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	service.LastUpdatedBy = service.CreatedBy

	existing, err := d.pmsRepo.FindOne(ctx, repos.Filter{
		fields.ProjectName:  projectName,
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: service.Name,
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

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeProjectManagedService, pms.Name, PublishAdd)

	return pms, nil
}

func (d *domain) getPMSTargetNamespace(projectName string, msvcName string, msvcType entities.ResourceType) string {
	msvcNamespace := fmt.Sprintf("pmsvc-%s-%s-%s", projectName, msvcName, msvcType)
	hash := md5.Sum([]byte(msvcNamespace))
	return fmt.Sprintf("pmsvc-%s", hex.EncodeToString(hash[:]))
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

	service.Namespace = "trest"
	service.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &service.ProjectManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	if pmsvc.IsMarkedForDeletion() {
		return nil, errors.Newf("cluster managed service %q (projectName=%q) is marked for deletion", service.Name, projectName)
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&service,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.ProjectManagedServiceSpecMsvcSpecServiceTemplateSpec: service.Spec.MSVCSpec.ServiceTemplate.Spec,
			},
		})

	upmsvc, err := d.pmsRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: service.Name,
		},
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeProjectManagedService, service.Name, PublishUpdate)

	if err := d.applyProjectManagedService(ctx, upmsvc); err != nil {
		return nil, errors.NewE(err)
	}

	return upmsvc, nil
}

func (d *domain) DeleteProjectManagedService(ctx ConsoleContext, projectName string, name string) error {
	if err := d.canMutateResourcesInProject(ctx, projectName); err != nil {
		return errors.NewE(err)
	}

	upmsvc, err := d.pmsRepo.Patch(
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

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeProjectManagedService, name, PublishUpdate)

	return d.deleteK8sResource(ctx, projectName, &upmsvc.ProjectManagedService)
}

// RestartProjectManagedService implements Domain.
func (d *domain) RestartProjectManagedService(ctx ConsoleContext, projectName string, name string) error {
	if err := d.canMutateResourcesInProject(ctx, projectName); err != nil {
		return errors.NewE(err)
	}

	pms, err := d.findProjectManagedService(ctx, projectName, name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.restartK8sResource(ctx, projectName, pms.Spec.TargetNamespace, pms.GetEnsuredLabels()); err != nil {
		return err
	}

	return nil
}

func (d *domain) OnProjectManagedServiceApplyError(ctx ConsoleContext, projectName, name, errMsg string, opts UpdateAndDeleteOpts) error {
	upmsvc, err := d.pmsRepo.Patch(
		ctx,
		repos.Filter{
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

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeProjectManagedService, upmsvc.Name, PublishDelete)

	return errors.NewE(err)
}

func (d *domain) OnProjectManagedServiceDeleteMessage(ctx ConsoleContext, projectName string, service entities.ProjectManagedService) error {
	err := d.pmsRepo.DeleteOne(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: service.Name,
		},
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeProjectManagedService, service.Name, PublishDelete)
	return nil
}

func (d *domain) OnProjectManagedServiceUpdateMessage(ctx ConsoleContext, projectName string, service entities.ProjectManagedService, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	svc, err := d.findProjectManagedService(ctx, projectName, service.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if svc == nil {
		return errors.Newf("no project manage service found")
	}

	recordVersion, err := d.MatchRecordVersion(service.Annotations, svc.RecordVersion)
	if err != nil {
		return d.ResyncProjectManagedService(ctx, service.ProjectName, service.Name)
	}

	upmsvc, err := d.pmsRepo.PatchById(
		ctx,
		svc.Id,
		common.PatchForSyncFromAgent(
			&service,
			recordVersion,
			status,
			common.PatchOpts{
				MessageTimestamp: opts.MessageTimestamp,
				XPatch: repos.Document{
					fc.ProjectManagedServiceSyncedOutputSecretRef: service.SyncedOutputSecretRef,
				},
			}))
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeProjectManagedService, upmsvc.Name, PublishUpdate)

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
