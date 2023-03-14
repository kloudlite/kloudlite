package domain

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

// CreateProject implements Domain
func (d *domain) CreateProject(ctx context.Context, project entities.Project) (*entities.Project, error) {
	project.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &project.Project); err != nil {
		return nil, err
	}
	prj, err := d.projectRepo.Create(ctx, &project)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &prj.Project); err != nil {
		return nil, err
	}

	return prj, nil
}

// DeleteProject implements Domain
func (d *domain) DeleteProject(ctx context.Context, name string) error {
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

	return d.k8sYamlClient.DeleteResource(ctx, &prj.Project)
}

// GetProject implements Domain
func (d *domain) GetProject(ctx context.Context, name string) (*entities.Project, error) {
	return d.findProject(ctx, name)
}

// GetProjects implements Domain
// func (d *domain) GetProjects(ctx context.Context) ([]*entities.Project, error) {
func (d *domain) GetProjects(ctx ConsoleContext) ([]*entities.Project, error) {
	return d.projectRepo.Find(ctx, repos.Query{})
}

// UpdateProject implements Domain
func (d *domain) UpdateProject(ctx context.Context, project entities.Project) (*entities.Project, error) {
	project.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &project.Project); err != nil {
		return nil, err
	}

	exProject, err := d.findProject(ctx, project.Name)
	if err != nil {
		return nil, err
	}

	if exProject.GetDeletionTimestamp() != nil {
		return nil, errAlreadyMarkedForDeletion("app", "", project.Name)
	}

	status := exProject.Status
	exProject.Project = project.Project
	exProject.Status = status

	upProject, err := d.projectRepo.UpdateById(ctx, exProject.Id, exProject)
	if err != nil {
		return nil, err
	}
	return upProject, nil
}

func (d *domain) findProject(ctx context.Context, name string) (*entities.Project, error) {
	prj, err := d.projectRepo.FindOne(ctx, repos.Filter{"metadata.name": name})
	if err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, fmt.Errorf("no project with name=%s found", name)
	}
	return prj, nil
}
