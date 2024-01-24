package domain

import (
	"context"
	"fmt"

	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
)

func (d *domain) getClusterAttachedToProject(ctx K8sContext, projectName string) (*string, error) {
	cacheKey := fmt.Sprintf("account_name_%s-project_name_%s", ctx.GetAccountName(), projectName)

	clusterName, err := d.consoleCacheStore.Get(ctx, cacheKey)
	if err != nil && !d.consoleCacheStore.ErrKeyNotFound(err) {
		return nil, err
	}

	if len(clusterName) == 0 {
		proj, err := d.projectRepo.FindOne(ctx, repos.Filter{
			fields.AccountName:  ctx.GetAccountName(),
			fields.MetadataName: projectName,
		})
		if err != nil {
			return nil, errors.NewE(err)
		}
		if proj == nil {
			return nil, errors.Newf("no cluster attached to project")
		}

		defer func() {
			if err := d.consoleCacheStore.Set(ctx, cacheKey, []byte(fn.DefaultIfNil(proj.ClusterName))); err != nil {
				d.logger.Infof("failed to set project cluster map: %v", err)
			}
		}()

		return proj.ClusterName, nil
	}

	if clusterName == nil {
		return nil, nil
	}

	return fn.New(string(clusterName)), nil
}

func (d *domain) ListProjects(ctx context.Context, userId repos.ID, accountName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Project], error) {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(userId),
		ResourceRefs: []string{
			iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
		},
		Action: string(iamT.ListProjects),
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if !co.Status {
		return nil, errors.Newf("unauthorized to get project")
	}

	filter := repos.Filter{fields.AccountName: accountName}

	// return d.projectRepo.Find(ctx, repos.Query{Filter: filter})
	return d.projectRepo.FindPaginated(ctx, d.projectRepo.MergeMatchFilters(filter, search), pagination)
}

func (d *domain) findProject(ctx ConsoleContext, name string) (*entities.Project, error) {
	prj, err := d.projectRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if prj == nil {
		return nil, errors.Newf("no project with name=%q found", name)
	}
	return prj, nil
}

func (d *domain) findProjectByTargetNs(ctx ConsoleContext, targetNamespace string) (*entities.Project, error) {
	prj, err := d.projectRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:            ctx.AccountName,
		fc.ProjectSpecTargetNamespace: targetNamespace,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if prj == nil {
		return nil, errors.Newf("no project with targetNamespace=%q found", targetNamespace)
	}
	return prj, nil
}

func (d *domain) GetProject(ctx ConsoleContext, name string) (*entities.Project, error) {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, name),
		},
		Action: string(iamT.GetProject),
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if !co.Status {
		return nil, errors.Newf("unauthorized to get project")
	}

	return d.findProject(ctx, name)
}

// mutations

func (d *domain) CreateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error) {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
		},
		Action: string(iamT.CreateProject),
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if !co.Status {
		return nil, errors.Newf("unauthorized to create Project")
	}

	project.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &project.Project); err != nil {
		return nil, errors.NewE(err)
	}

	project.IncrementRecordVersion()

	project.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	project.LastUpdatedBy = project.CreatedBy

	project.AccountName = ctx.AccountName
	project.SyncStatus = t.GenSyncStatus(t.SyncActionApply, project.RecordVersion)

	project.AccountName = ctx.AccountName
	if project.Spec.TargetNamespace == "" {
		project.Spec.TargetNamespace = d.getProjectNamespace(project.Name)
	}

	prj, err := d.projectRepo.Create(ctx, &project)
	if err != nil {
		if d.projectRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeProject, prj.Name, PublishAdd)

	if err := d.applyK8sResource(ctx, prj.Name, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: prj.Spec.TargetNamespace,
			Annotations: map[string]string{
				constants.DescriptionKey: "This namespace is managed (created/updated/deleted) by kloudlite.io control plane. This namespace belongs to a project",
			},
			Labels: map[string]string{
				constants.ProjectNameKey: prj.Name,
			},
		},
	}, prj.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, prj.Name, &prj.Project, prj.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return prj, nil
}

func (d *domain) getProjectNamespace(projectName string) string {
	return fmt.Sprintf("prj-%s", projectName)
}

func (d *domain) DeleteProject(ctx ConsoleContext, name string) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
		},
		Action: string(iamT.DeleteProject),
	})
	if err != nil {
		return errors.NewE(err)
	}

	if !co.Status {
		return errors.Newf("unauthorized to delete project")
	}

	uproj, err := d.projectRepo.Patch(
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

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeProject, name, PublishUpdate)

	return d.deleteK8sResource(ctx, uproj.Name, &uproj.Project)
}

func (d *domain) UpdateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error) {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, project.Name),
		},
		Action: string(iamT.UpdateProject),
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if !co.Status {
		return nil, errors.Newf("unauthorized to update project %q", project.Name)
	}

	project.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &project.Project); err != nil {
		return nil, errors.NewE(err)
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&project,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.ProjectSpec: project.Spec,
			},
		})

	upProject, err := d.projectRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: project.Name,
		},
		patchForUpdate,
	)

	if err != nil {
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeProject, project.Name, PublishUpdate)

	if err := d.applyK8sResource(ctx, upProject.Name, &upProject.Project, upProject.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return upProject, nil
}

func (d *domain) OnProjectDeleteMessage(ctx ConsoleContext, project entities.Project) error {
	err := d.projectRepo.DeleteOne(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: project.Name,
		},
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeApp, project.Name, PublishDelete)
	return nil
}

func (d *domain) OnProjectUpdateMessage(ctx ConsoleContext, project entities.Project, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	proj, err := d.findProject(ctx, project.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if proj == nil {
		return errors.Newf("no project found")
	}

	recordVersion, err := d.MatchRecordVersion(project.Annotations, proj.RecordVersion)
	if err != nil {
		return nil
	}

	uproject, err := d.projectRepo.PatchById(
		ctx,
		proj.Id,
		common.PatchForSyncFromAgent(
			&project,
			recordVersion,
			status,
			common.PatchOpts{
				MessageTimestamp: opts.MessageTimestamp,
			}))

	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeProject, uproject.Name, PublishUpdate)

	return nil
}

func (d *domain) OnProjectApplyError(ctx ConsoleContext, errMsg string, name string, opts UpdateAndDeleteOpts) error {
	uproject, err := d.projectRepo.Patch(
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

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeApp, uproject.Name, PublishDelete)

	return errors.NewE(err)
}

func (d *domain) ResyncProject(ctx ConsoleContext, name string) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, name),
		},
		Action: string(iamT.UpdateProject),
	})
	if err != nil {
		return errors.NewE(err)
	}

	if !co.Status {
		return errors.Newf("unauthorized to update project %q", name)
	}

	project, err := d.findProject(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.resyncK8sResource(ctx, project.Name, project.SyncStatus.Action, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: project.Spec.TargetNamespace,
			Annotations: map[string]string{
				constants.DescriptionKey: "This namespace is managed (created/updated/deleted) by kloudlite.io control plane. This namespace belongs to a project",
			},
			Labels: map[string]string{
				constants.ProjectNameKey: project.Name,
			},
		},
	}, 0); err != nil {
		return errors.NewE(err)
	}

	return d.resyncK8sResource(ctx, project.Name, project.SyncStatus.Action, &project.Project, project.RecordVersion)
}
