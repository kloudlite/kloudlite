package domain

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/common"
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
		ObjectMeta: metav1.ObjectMeta{Name: prj.Name},
	}); err != nil {
		return nil, err
	}

	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(prj.Id), prj.Project); err != nil {
		return nil, err
	}

	return prj, nil
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
