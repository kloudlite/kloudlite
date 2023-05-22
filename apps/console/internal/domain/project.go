package domain

import (
	"context"
	"fmt"
	"time"

	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kloudlite.io/apps/console/internal/domain/entities"
	iamT "kloudlite.io/apps/iam/types"
	"kloudlite.io/grpc-interfaces/kloudlite.io/rpc/iam"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

// query

func (d *domain) ListProjects(ctx context.Context, userId repos.ID, accountName string, clusterName *string) ([]*entities.Project, error) {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(userId),
		ResourceRefs: []string{
			iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
		},
		Action: string(iamT.ListProjects),
	})
	if err != nil {
		return nil, err
	}

	if !co.Status {
		return nil, fmt.Errorf("unauthorized to get project")
	}

	filter := repos.Filter{"accountName": accountName}
	if clusterName != nil {
		filter["clusterName"] = clusterName
	}
	return d.projectRepo.Find(ctx, repos.Query{Filter: filter})
}

func (d *domain) findProject(ctx ConsoleContext, name string) (*entities.Project, error) {
	prj, err := d.projectRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"clusterName":   ctx.ClusterName,
		"metadata.name": name,
	})
	if err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, fmt.Errorf("no project with name=%q found", name)
	}
	return prj, nil
}

func (d *domain) findProjectByTargetNs(ctx ConsoleContext, targetNamespace string) (*entities.Project, error) {
	prj, err := d.projectRepo.FindOne(ctx, repos.Filter{
		"accountName":          ctx.AccountName,
		"clusterName":          ctx.ClusterName,
		"spec.targetNamespace": targetNamespace,
	})

	if err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, fmt.Errorf("no project with targetNamespace=%q found", targetNamespace)
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
		return nil, err
	}

	if !co.Status {
		return nil, fmt.Errorf("unauthorized to get project")
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
		return nil, err
	}

	if !co.Status {
		return nil, fmt.Errorf("unauthorized to create Project")
	}

	project.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &project.Project); err != nil {
		return nil, err
	}

	project.AccountName = ctx.AccountName
	project.ClusterName = ctx.ClusterName
	project.Generation = 1
	project.SyncStatus = t.GenSyncStatus(t.SyncActionApply, project.Generation)

	prj, err := d.projectRepo.Create(ctx, &project)
	if err != nil {
		if d.projectRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf("project with name %q, already exists", project.Name)
		}
		return nil, err
	}

	defaultWs := entities.Workspace{
		Env: crdsv1.Env{
			ObjectMeta: metav1.ObjectMeta{
				Name:       d.envVars.DefaultProjectWorkspaceName,
				Namespace:  project.Spec.TargetNamespace,
				Generation: 1,
			},
			Spec: crdsv1.EnvSpec{
				ProjectName:     project.Name,
				TargetNamespace: fmt.Sprintf("%s-%s", project.Name, d.envVars.DefaultProjectWorkspaceName),
			},
		},
		AccountName: ctx.AccountName,
		ClusterName: ctx.ClusterName,
	}

	if _, err := d.CreateWorkspace(ctx, defaultWs); err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: prj.Spec.TargetNamespace,
			Labels: map[string]string{
				constants.ProjectNameKey: prj.Name,
			},
		},
	}); err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &prj.Project); err != nil {
		return nil, err
	}

	return prj, nil
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
		return err
	}

	if !co.Status {
		return fmt.Errorf("unauthorized to delete project")
	}

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
		return nil, err
	}

	if !co.Status {
		return nil, fmt.Errorf("unauthorized to update project %q", project.Name)
	}

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
	exProject.SyncStatus = t.GetSyncStatusForUpdation(exProject.Generation + 1)

	upProject, err := d.projectRepo.UpdateById(ctx, exProject.Id, exProject)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upProject.Project); err != nil {
		return nil, err
	}

	return upProject, nil
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
	p.SyncStatus.Error = nil
	p.SyncStatus.LastSyncedAt = time.Now()
	p.SyncStatus.Generation = project.Generation
	p.SyncStatus.State = t.ParseSyncState(project.Status.IsReady)

	_, err = d.projectRepo.UpdateById(ctx, p.Id, p)
	return err
}

func (d *domain) OnApplyProjectError(ctx ConsoleContext, errMsg string, name string) error {
	p, err2 := d.findProject(ctx, name)
	if err2 != nil {
		return err2
	}

	p.SyncStatus.Error = &errMsg
	_, err := d.projectRepo.UpdateById(ctx, p.Id, p)
	return err
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
		return err
	}

	if !co.Status {
		return fmt.Errorf("unauthorized to update project %q", name)
	}

	p, err := d.findProject(ctx, name)
	if err != nil {
		return err
	}

	if err := d.resyncK8sResource(ctx, t.SyncActionApply, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: p.Spec.TargetNamespace,
			Labels: map[string]string{
				constants.ProjectNameKey: p.Name,
			},
		},
	}); err != nil {
		return err
	}

	return d.resyncK8sResource(ctx, p.SyncStatus.Action, &p.Project)
}
