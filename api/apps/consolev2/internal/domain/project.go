package domain

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/common"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateProject(ctx context.Context, project entities.Project) (*entities.Project, error) {
	isAvailable, err := d.isNameAvailable(ctx, common.ResourceProject, project.Namespace, project.Name)
	if err != nil {
		return nil, err
	}
	if !isAvailable {
		return nil, fmt.Errorf("project with (name=%s) already exists", project.Name)
	}

	prj, err := d.projectRepo.Create(ctx, &project)
	if err != nil {
		return nil, err
	}

	clusterId, err := d.getClusterForProject(ctx, prj.Name)
	if err != nil {
		return nil, err
	}

	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(prj.Id), &corev1.Namespace{
		TypeMeta:   metav1.TypeMeta{Kind: "Namespace", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: prj.Name},
	}); err != nil {
		return nil, err
	}

	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(prj.Id), prj.Project); err != nil {
		return nil, err
	}

	return prj, nil
}

func (d *domain) UpdateProject(ctx context.Context, project entities.Project) (*entities.Project, error) {
	prj, err := d.projectRepo.FindOne(ctx, repos.Filter{"metadata.name": project.Name})
	if err != nil {
		return nil, errors.Newf("project %s not found", project.Name)
	}

	prj.Project = project.Project
	nPrj, err := d.projectRepo.UpdateById(ctx, prj.Id, prj)
	if err != nil {
		return nil, err
	}

	clusterId, err := d.getClusterForProject(ctx, prj.Name)
	if err != nil {
		return nil, err
	}

	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(nPrj.Id), nPrj); err != nil {
		return nil, err
	}

	return nPrj, nil
}

func (d *domain) DeleteProject(ctx context.Context, name string) (bool, error) {
	prj, err := d.projectRepo.FindOne(ctx, repos.Filter{"metadata.name": name})
	if err != nil {
		return false, err
	}
	clusterId, err := d.getClusterForProject(ctx, prj.Name)

	if err := d.workloadMessenger.SendAction(ActionDelete, d.getDispatchKafkaTopic(clusterId), string(prj.Id), prj.Project); err != nil {
		return false, err
	}
	if err := d.projectRepo.DeleteOne(ctx, repos.Filter{"metadata.name": name}); err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) GetAccountProjects(ctx context.Context, accountId repos.ID) ([]*entities.Project, error) {
	return d.projectRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{"spec.accountId": accountId},
	})
}

func (d *domain) GetProjectWithID(ctx context.Context, projectId repos.ID) (*entities.Project, error) {
	return d.projectRepo.FindById(ctx, projectId)
}

func (d *domain) GetProjectWithName(ctx context.Context, projectName string) (*entities.Project, error) {
	return d.projectRepo.FindOne(ctx, repos.Filter{"metadata.name": projectName})
}
