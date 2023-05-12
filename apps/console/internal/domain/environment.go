package domain

import (
	"fmt"
	"time"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
	t "kloudlite.io/pkg/types"
)

// environment:query

func (d *domain) findEnvironment(ctx ConsoleContext, namespace, name string) (*entities.Environment, error) {
	env, err := d.environmentRepo.FindOne(ctx, repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
		"metadata.name":      name,
	})
	if err != nil {
		return nil, err
	}
	if env == nil {
		return nil, fmt.Errorf("no environment with name=%q found", name)
	}
	return env, nil
}

func (d *domain) GetEnvironment(ctx ConsoleContext, namespace, name string) (*entities.Environment, error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, err
	}
	return d.findEnvironment(ctx, namespace, name)
}

func (d *domain) ListEnvironments(ctx ConsoleContext, namespace string) ([]*entities.Environment, error) {
	if err := d.canReadResourcesInProject(ctx, namespace); err != nil {
		return nil, err
	}

	filter := repos.Filter{
		"accountName":        ctx.AccountName,
		"clusterName":        ctx.ClusterName,
		"metadata.namespace": namespace,
	}
	return d.environmentRepo.Find(ctx, repos.Query{Filter: filter})
}

// mutations

func (d *domain) CreateEnvironment(ctx ConsoleContext, env entities.Environment) (*entities.Environment, error) {
	env.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &env.Env); err != nil {
		return nil, err
	}

	if err := d.canMutateResourcesInProject(ctx, env.Spec.ProjectName); err != nil {
		return nil, err
	}

	env.AccountName = ctx.AccountName
	env.ClusterName = ctx.ClusterName
	env.Generation = 1
	env.SyncStatus = t.GenSyncStatus(t.SyncActionApply, env.Generation)

	nEnv, err := d.environmentRepo.Create(ctx, &env)
	if err != nil {
		if d.environmentRepo.ErrAlreadyExists(err) {
			return nil, fmt.Errorf(
				"environment with name %q, namespace=%q already exists",
				env.Name,
				env.Namespace,
			)
		}
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &env.Env); err != nil {
		return nil, err
	}

	return nEnv, nil
}

// UpdateEnvironment implements Domain
func (d *domain) UpdateEnvironment(ctx ConsoleContext, env entities.Environment) (*entities.Environment, error) {
	env.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &env.Env); err != nil {
		return nil, err
	}

	if err := d.canMutateResourcesInProject(ctx, env.Spec.ProjectName); err != nil {
		return nil, err
	}

	exEnv, err := d.findEnvironment(ctx, env.Namespace, env.Name)
	if err != nil {
		return nil, err
	}

	if exEnv.GetDeletionTimestamp() != nil {
		return nil, errAlreadyMarkedForDeletion("environment", "", env.Name)
	}

	exEnv.Spec = env.Spec
	exEnv.Generation += 1
	exEnv.SyncStatus = t.GenSyncStatus(t.SyncActionApply, exEnv.Generation)

	upEnv, err := d.environmentRepo.UpdateById(ctx, exEnv.Id, exEnv)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upEnv.Env); err != nil {
		return nil, err
	}

	return upEnv, nil
}

// DeleteEnvironment implements Domain
func (d *domain) DeleteEnvironment(ctx ConsoleContext, namespace, name string) error {
	env, err := d.findEnvironment(ctx, namespace, name)
	if err != nil {
		return err
	}

	if err := d.canMutateResourcesInProject(ctx, env.Spec.ProjectName); err != nil {
		return err
	}

	env.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, env.Generation)
	if _, err := d.environmentRepo.UpdateById(ctx, env.Id, env); err != nil {
		return err
	}

	return d.deleteK8sResource(ctx, &env.Env)
}

func (d *domain) OnApplyEnvironmentError(ctx ConsoleContext, errMsg, namespace, name string) error {
	env, err2 := d.findEnvironment(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	env.SyncStatus.Error = &errMsg
	_, err := d.environmentRepo.UpdateById(ctx, env.Id, env)
	return err
}

func (d *domain) OnDeleteEnvironmentMessage(ctx ConsoleContext, env entities.Environment) error {
	p, err := d.findEnvironment(ctx, env.Namespace, env.Name)
	if err != nil {
		return err
	}

	return d.environmentRepo.DeleteById(ctx, p.Id)
}

// OnUpdateEnvironmentMessage implements Domain
func (d *domain) OnUpdateEnvironmentMessage(ctx ConsoleContext, env entities.Environment) error {
	e, err := d.findEnvironment(ctx, env.Namespace, env.Name)
	if err != nil {
		return err
	}

	e.Status = env.Status
	e.SyncStatus.Error = nil
	e.SyncStatus.LastSyncedAt = time.Now()
	e.SyncStatus.Generation = env.Generation
	e.SyncStatus.State = t.ParseSyncState(env.Status.IsReady)

	_, err = d.environmentRepo.UpdateById(ctx, e.Id, e)
	return err
}

// ResyncEnvironment implements Domain
func (d *domain) ResyncEnvironment(ctx ConsoleContext, namespace, name string) error {
	e, err := d.findEnvironment(ctx, namespace, name)
	if err != nil {
		return err
	}

	return d.resyncK8sResource(ctx, e.SyncStatus.Action, &e.Env)
}
