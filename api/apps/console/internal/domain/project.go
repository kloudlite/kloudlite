package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

// CreateProject implements Domain
func (d *domain) CreateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error) {
	project.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &project.Project); err != nil {
		return nil, err
	}

	project.AccountName = ctx.accountName
	project.ClusterName = ctx.clusterName
	project.SyncStatus = t.GetSyncStatusForCreation()
	prj, err := d.projectRepo.Create(ctx, &project)
	if err != nil {
		if d.projectRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("project with name %q, already exists", project.Name)
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &prj.Project); err != nil {
		return nil, err
	}

	return prj, nil
}

func (d *domain) DeleteProject(ctx ConsoleContext, name string) error {
	prj, err := d.findProject(ctx, name)
	if err != nil {
		return err
	}

	prj.SyncStatus = t.GetSyncStatusForDeletion(prj.Generation)
	if _, err := d.projectRepo.UpdateById(ctx, prj.Id, prj); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &prj.Project)
}

// GetProject implements Domain
func (d *domain) GetProject(ctx ConsoleContext, name string) (*entities.Project, error) {
	return d.findProject(ctx, name)
}

func (d *domain) ListProjects(ctx ConsoleContext) ([]*entities.Project, error) {
	return d.projectRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"accountName": ctx.accountName,
		"clusterName": ctx.clusterName,
	}})
}

func (d *domain) UpdateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error) {
	project.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &project.Project); err != nil {
		return nil, err
	}

	exProject, err := d.findProject(ctx, project.Name)
	if err != nil {
		return nil, err
	}

	if exProject.GetDeletionTimestamp() != nil {
		return nil, errAlreadyMarkedForDeletion("project", "", project.Name)
	}

	exProject.Spec = project.Spec
	exProject.SyncStatus = t.GetSyncStatusForUpdation(exProject.Generation)

	upProject, err := d.projectRepo.UpdateById(ctx, exProject.Id, exProject)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upProject.Project); err != nil {
		return nil, err
	}

	return upProject, nil
}

func (d *domain) findProject(ctx ConsoleContext, name string) (*entities.Project, error) {
	prj, err := d.projectRepo.FindOne(ctx, repos.Filter{
		"accountName":      ctx.accountName,
		"clusterName":      ctx.clusterName,
		"metadata.name":    name,
		"spec.accountName": ctx.accountName,
	})
	if err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, fmt.Errorf("no project with name=%q found", name)
	}
	return prj, nil
}

func (d *domain) OnDeleteProjectMessage(ctx ConsoleContext, project entities.Project) error {
	p, err := d.findProject(ctx, project.Name)
	if err != nil {
		return err
	}

	return d.projectRepo.DeleteById(ctx, p.Id)
}

func (d *domain) OnUpdateProjectMessage(ctx ConsoleContext, project entities.Project) error {
	p, err := d.findProject(ctx, project.Name)
	if err != nil {
		return err
	}

	p.Status = project.Status
	p.SyncStatus.LastSyncedAt = time.Now()
	p.SyncStatus.State = t.ParseSyncState(project.Status.IsReady)

	_, err = d.projectRepo.UpdateById(ctx, p.Id, p)
	return err
}

func (d *domain) OnApplyProjectError(ctx ConsoleContext, err error, name string) error {
	p, err2 := d.findProject(ctx, name)
	if err2 != nil {
		return err2
	}

	p.SyncStatus.Error = err.Error()
	_, err = d.projectRepo.UpdateById(ctx, p.Id, p)
	return err
}
