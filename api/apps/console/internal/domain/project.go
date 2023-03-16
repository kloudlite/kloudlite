package domain

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

// CreateProject implements Domain
func (d *domain) CreateProject(ctx ConsoleContext, project entities.Project) (*entities.Project, error) {
	project.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &project.Project); err != nil {
		return nil, err
	}

	project.AccountName = ctx.accountName
	project.ClusterName = ctx.clusterName
	prj, err := d.projectRepo.Create(ctx, &project)
	if err != nil {
		if d.projectRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("project with name %s, already exists", project.Name)
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

	if prj.GetDeletionTimestamp() != nil {
		return errAlreadyMarkedForDeletion("app", prj.Namespace, prj.Name)
	}

	prj.SetDeletionTimestamp(&metav1.Time{Time: time.Now()})
	if _, err := d.projectRepo.UpdateById(ctx, prj.Id, prj); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &prj.Project)
}

// GetProject implements Domain
func (d *domain) GetProject(ctx ConsoleContext, name string) (*entities.Project, error) {
	return d.findProject(ctx, name)
}

// ListProjects implements Domain
// func (d *domain) ListProjects(ctx ConsoleContext) ([]*entities.Project, error) {
func (d *domain) ListProjects(ctx ConsoleContext) ([]*entities.Project, error) {
	return d.projectRepo.Find(ctx, repos.Query{Filter: repos.Filter{
		"accountName": ctx.accountName,
		"clusterName": ctx.clusterName,
	}})
}

// UpdateProject implements Domain
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

	status := exProject.Status
	exProject.Project = project.Project
	exProject.Status = status

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
		return nil, fmt.Errorf("no project with name=%s found", name)
	}
	return prj, nil
}
