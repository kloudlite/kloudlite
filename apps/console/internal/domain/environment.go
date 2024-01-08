package domain

import (
	"fmt"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/kv"
	"github.com/kloudlite/operator/operators/resource-watcher/types"

	"github.com/kloudlite/api/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/api/apps/console/internal/entities"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
)

func (d *domain) findEnvironment(ctx ConsoleContext, projectName string, name string) (*entities.Environment, error) {
	env, err := d.environmentRepo.FindOne(ctx, repos.Filter{
		"accountName":   ctx.AccountName,
		"projectName":   projectName,
		"metadata.name": name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if env == nil {
		return nil, errors.Newf("no environment with name (%s) and project (%s)", name, projectName)
	}
	return env, nil
}

func (d *domain) envTargetNamespace(ctx ConsoleContext, projectName string, envName string) (string, error) {
	key := fmt.Sprintf("environment-namespace.%s/%s/%s", ctx.AccountName, projectName, envName)
	b, err := d.consoleCacheStore.Get(ctx, key)
	if err != nil {
		if errors.Is(err, kv.ErrKeyNotFound) {
			env, err := d.findEnvironment(ctx, projectName, envName)
			if err != nil {
				return "", err
			}
			defer func() {
				if err := d.consoleCacheStore.Set(ctx, key, []byte(env.Spec.TargetNamespace)); err != nil {
					d.logger.Errorf(err, "while caching environment target namespace")
				}
			}()
			return env.Spec.TargetNamespace, nil
		}
	}

	return string(b), nil
}

func (d *domain) GetEnvironment(ctx ConsoleContext, projectName string, name string) (*entities.Environment, error) {
	if err := d.canReadResourcesInProject(ctx, projectName); err != nil {
		return nil, errors.NewE(err)
	}

	return d.findEnvironment(ctx, projectName, name)
}

func (d *domain) ListEnvironments(ctx ConsoleContext, projectName string, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Environment], error) {
	if err := d.canReadResourcesInProject(ctx, projectName); err != nil {
		return nil, errors.NewE(err)
	}

	filter := repos.Filter{
		"accountName": ctx.AccountName,
		"projectName": projectName,
	}

	return d.environmentRepo.FindPaginated(ctx, d.environmentRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) findEnvironmentByTargetNs(ctx ConsoleContext, targetNs string) (*entities.Environment, error) {
	w, err := d.environmentRepo.FindOne(ctx, repos.Filter{
		"accountName":          ctx.AccountName,
		"spec.targetNamespace": targetNs,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if w == nil {
		return nil, errors.Newf("no workspace found for target namespace %q", targetNs)
	}

	return w, nil
}

func (d *domain) CreateEnvironment(ctx ConsoleContext, projectName string, env entities.Environment) (*entities.Environment, error) {
	project, err := d.findProject(ctx, projectName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.canMutateResourcesInProject(ctx, project.Name); err != nil {
		return nil, errors.NewE(err)
	}

	env.ProjectName = project.Name

	env.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &env.Environment); err != nil {
		return nil, errors.NewE(err)
	}

	env.IncrementRecordVersion()

	if env.Spec.TargetNamespace == "" {
		env.Spec.TargetNamespace = fmt.Sprintf("env-%s", env.Name)
	}

	env.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	env.LastUpdatedBy = env.CreatedBy

	env.AccountName = ctx.AccountName
	env.SyncStatus = t.GenSyncStatus(t.SyncActionApply, env.RecordVersion)

	if _, err := d.upsertResourceMapping(ResourceContext{ConsoleContext: ctx, ProjectName: env.ProjectName, EnvironmentName: env.Name}, &env); err != nil {
		return nil, errors.NewE(err)
	}

	nenv, err := d.environmentRepo.Create(ctx, &env)
	if err != nil {
		if d.environmentRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishWorkspaceEvent(nenv, PublishAdd)

	if _, err := d.iamClient.AddMembership(ctx, &iam.AddMembershipIn{
		UserId:       string(ctx.UserId),
		ResourceType: string(iamT.ResourceEnvironment),
		ResourceRef:  iamT.NewResourceRef(ctx.AccountName, iamT.ResourceEnvironment, nenv.Name),
		Role:         string(iamT.RoleResourceOwner),
	}); err != nil {
		d.logger.Errorf(err, "error while adding membership")
	}

	if err := d.applyK8sResource(ctx, env.ProjectName, &nenv.Environment, nenv.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return nenv, nil
}

func (d *domain) CloneEnvironment(ctx ConsoleContext, projectName string, sourceEnvName string, envName string) (*entities.Environment, error) {
	if err := d.canMutateResourcesInProject(ctx, projectName); err != nil {
		return nil, errors.NewE(err)
	}

	env, err := d.findEnvironment(ctx, projectName, sourceEnvName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	env.ProjectName = projectName
	env.DisplayName = envName
	env.Name = envName
	env.Id = d.environmentRepo.NewId()
	env.PrimitiveId = ""

	env.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &env.Environment); err != nil {
		return nil, errors.NewE(err)
	}

	env.IncrementRecordVersion()

	env.Spec.TargetNamespace = fmt.Sprintf("env-%s", envName)

	env.SyncStatus = t.GenSyncStatus(t.SyncActionApply, env.RecordVersion)

	if _, err := d.upsertResourceMapping(ResourceContext{ConsoleContext: ctx, ProjectName: env.ProjectName, EnvironmentName: env.Name}, env); err != nil {
		return nil, errors.NewE(err)
	}

	nenv, err := d.environmentRepo.Create(ctx, env)
	if err != nil {
		if d.environmentRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}
	resContext := ResourceContext{
		ConsoleContext:  ctx,
		ProjectName:     env.ProjectName,
		EnvironmentName: nenv.Name,
	}
	filters := repos.Filter{
		"accountName":     resContext.AccountName,
		"projectName":     resContext.ProjectName,
		"environmentName": sourceEnvName,
	}

	apps, err := d.appRepo.Find(ctx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	secrets, err := d.secretRepo.Find(ctx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	configs, err := d.configRepo.Find(ctx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	routers, err := d.routerRepo.Find(ctx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	managedResources, err := d.mresRepo.Find(ctx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	for _, app := range apps {
		app.Namespace = env.Spec.TargetNamespace
		app.EnvironmentName = envName
		app.Id = d.appRepo.NewId()
		app.PrimitiveId = ""
		err := cloneResource(resContext, d, d.appRepo, app, &app.App)
		if err != nil {
			return nil, errors.NewE(err)
		}
	}
	for _, secret := range secrets {
		secret.Namespace = env.Spec.TargetNamespace
		secret.EnvironmentName = envName
		secret.Id = d.secretRepo.NewId()
		secret.PrimitiveId = ""
		err := cloneResource(resContext, d, d.secretRepo, secret, &secret.Secret)
		if err != nil {
			return nil, errors.NewE(err)
		}
	}
	for _, config := range configs {
		config.Namespace = env.Spec.TargetNamespace
		config.EnvironmentName = envName
		config.Id = d.configRepo.NewId()
		config.PrimitiveId = ""
		err := cloneResource(resContext, d, d.configRepo, config, &config.ConfigMap)
		if err != nil {
			return nil, errors.NewE(err)
		}
	}
	for _, router := range routers {
		router.Namespace = env.Spec.TargetNamespace
		router.EnvironmentName = envName
		router.Id = d.routerRepo.NewId()
		router.PrimitiveId = ""
		err := cloneResource(resContext, d, d.routerRepo, router, &router.Router)
		if err != nil {
			return nil, errors.NewE(err)
		}
	}
	for _, managedResource := range managedResources {
		managedResource.Namespace = env.Spec.TargetNamespace
		managedResource.EnvironmentName = envName
		managedResource.Id = d.mresRepo.NewId()
		managedResource.PrimitiveId = ""
		err := cloneResource(resContext, d, d.mresRepo, managedResource, &managedResource.ManagedResource)
		if err != nil {
			return nil, errors.NewE(err)
		}
	}

	if _, err := d.iamClient.AddMembership(ctx, &iam.AddMembershipIn{
		UserId:       string(ctx.UserId),
		ResourceType: string(iamT.ResourceEnvironment),
		ResourceRef:  iamT.NewResourceRef(ctx.AccountName, iamT.ResourceEnvironment, nenv.Name),
		Role:         string(iamT.RoleResourceOwner),
	}); err != nil {
		d.logger.Errorf(err, "error while adding membership")
	}

	if err := d.applyK8sResource(ctx, env.ProjectName, &nenv.Environment, nenv.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return nenv, nil
}

func (d *domain) UpdateEnvironment(ctx ConsoleContext, projectName string, env entities.Environment) (*entities.Environment, error) {
	if err := d.canMutateResourcesInProject(ctx, projectName); err != nil {
		return nil, errors.NewE(err)
	}

	env.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &env.Environment); err != nil {
		return nil, errors.NewE(err)
	}

	xenv, err := d.findEnvironment(ctx, projectName, env.Name)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if xenv.GetDeletionTimestamp() != nil {
		return nil, errAlreadyMarkedForDeletion("environment", env.Namespace, env.Name)
	}

	xenv.IncrementRecordVersion()
	xenv.LastUpdatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	xenv.DisplayName = env.DisplayName
	xenv.Labels = env.Labels
	xenv.Annotations = env.Annotations

	xenv.SyncStatus = t.GenSyncStatus(t.SyncActionApply, xenv.RecordVersion)

	upEnv, err := d.environmentRepo.UpdateById(ctx, xenv.Id, xenv)
	if err != nil {
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishWorkspaceEvent(upEnv, PublishUpdate)

	if err := d.applyK8sResource(ctx, xenv.ProjectName, &upEnv.Environment, upEnv.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return upEnv, nil
}

func (d *domain) DeleteEnvironment(ctx ConsoleContext, projectName string, name string) error {
	if err := d.canMutateResourcesInProject(ctx, projectName); err != nil {
		return errors.NewE(err)
	}

	env, err := d.findEnvironment(ctx, projectName, name)
	if err != nil {
		return errors.NewE(err)
	}

	env.MarkedForDeletion = fn.New(true)
	env.SyncStatus = t.GenSyncStatus(t.SyncActionDelete, env.RecordVersion)
	if _, err := d.environmentRepo.UpdateById(ctx, env.Id, env); err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishWorkspaceEvent(env, PublishUpdate)

	resContext := ResourceContext{
		ConsoleContext:  ctx,
		ProjectName:     env.ProjectName,
		EnvironmentName: name,
	}
	apps, err := d.appRepo.Find(ctx, repos.Query{
		Filter: resContext.DBFilters(),
		Sort:   nil,
	})
	secrets, err := d.secretRepo.Find(ctx, repos.Query{
		Filter: resContext.DBFilters(),
		Sort:   nil,
	})
	configs, err := d.configRepo.Find(ctx, repos.Query{
		Filter: resContext.DBFilters(),
		Sort:   nil,
	})
	routers, err := d.routerRepo.Find(ctx, repos.Query{
		Filter: resContext.DBFilters(),
		Sort:   nil,
	})
	managedResources, err := d.mresRepo.Find(ctx, repos.Query{
		Filter: resContext.DBFilters(),
		Sort:   nil,
	})

	for _, app := range apps {
		err := d.DeleteApp(resContext, app.Name)
		if err != nil {
			return errors.NewE(err)
		}
	}

	for _, secret := range secrets {
		err := d.DeleteSecret(resContext, secret.Name)
		if err != nil {
			return errors.NewE(err)
		}
	}

	for _, config := range configs {
		err := d.DeleteConfig(resContext, config.Name)
		if err != nil {
			return errors.NewE(err)
		}
	}

	for _, router := range routers {
		err := d.DeleteRouter(resContext, router.Name)
		if err != nil {
			return errors.NewE(err)
		}
	}

	for _, managedResource := range managedResources {
		err := d.DeleteManagedResource(resContext, managedResource.Name)
		if err != nil {
			return errors.NewE(err)
		}
	}

	if err := d.deleteK8sResource(ctx, env.ProjectName, &env.Environment); err != nil {
		return errors.NewE(err)
	}

	// FIXME (nxtcoder17): Should this be performed ?
	if err := d.deleteK8sResource(ctx, env.ProjectName, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: env.Spec.TargetNamespace,
		},
	}); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) OnEnvironmentApplyError(ctx ConsoleContext, errMsg, namespace, name string, opts UpdateAndDeleteOpts) error {
	ws, err2 := d.findEnvironment(ctx, namespace, name)
	if err2 != nil {
		return err2
	}

	ws.SyncStatus.State = t.SyncStateErroredAtAgent
	ws.SyncStatus.LastSyncedAt = opts.MessageTimestamp
	ws.SyncStatus.Error = &errMsg
	_, err := d.environmentRepo.UpdateById(ctx, ws.Id, ws)
	return errors.NewE(err)
}

func (d *domain) OnEnvironmentDeleteMessage(ctx ConsoleContext, ws entities.Environment) error {
	exWs, err := d.findEnvironment(ctx, ws.Namespace, ws.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(ws.Annotations, exWs.RecordVersion); err != nil {
		return errors.NewE(err)
	}

	err = d.environmentRepo.DeleteById(ctx, exWs.Id)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishWorkspaceEvent(exWs, PublishDelete)
	return nil
}

func (d *domain) OnEnvironmentUpdateMessage(ctx ConsoleContext, env entities.Environment, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xenv, err := d.findEnvironment(ctx, env.Namespace, env.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.MatchRecordVersion(env.Annotations, xenv.RecordVersion); err != nil {
		return errors.NewE(err)
	}

	xenv.CreationTimestamp = env.CreationTimestamp
	xenv.Labels = env.Labels
	xenv.Annotations = env.Annotations
	xenv.Generation = env.Generation

	xenv.Status = env.Status

	xenv.SyncStatus.State = func() t.SyncState {
		if status == types.ResourceStatusDeleting {
			return t.SyncStateDeletingAtAgent
		}
		return t.SyncStateUpdatedAtAgent
	}()
	xenv.SyncStatus.RecordVersion = xenv.RecordVersion
	xenv.SyncStatus.Error = nil
	xenv.SyncStatus.LastSyncedAt = opts.MessageTimestamp

	xenv, err = d.environmentRepo.UpdateById(ctx, xenv.Id, xenv)
	if err != nil {
		return err
	}
	d.resourceEventPublisher.PublishWorkspaceEvent(xenv, PublishUpdate)
	return nil
}

func (d *domain) ResyncEnvironment(ctx ConsoleContext, projectName string, name string) error {
	if err := d.canMutateResourcesInProject(ctx, projectName); err != nil {
		return errors.NewE(err)
	}

	e, err := d.findEnvironment(ctx, projectName, name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.resyncK8sResource(ctx, e.ProjectName, t.SyncActionApply, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: e.Spec.TargetNamespace,
			Labels: map[string]string{
				constants.EnvNameKey: e.Name,
			},
		},
	}, e.RecordVersion); err != nil {
		return errors.NewE(err)
	}

	return d.resyncK8sResource(ctx, "", e.SyncStatus.Action, &e.Environment, e.RecordVersion)
}
