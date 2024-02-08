package domain

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"

	"github.com/kloudlite/api/common/fields"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/operator/operators/resource-watcher/types"

	"github.com/kloudlite/api/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
)

func (d *domain) findEnvironment(ctx ConsoleContext, projectName string, name string) (*entities.Environment, error) {
	env, err := d.environmentRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.ProjectName:  projectName,
		fields.MetadataName: name,
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
		if d.consoleCacheStore.ErrKeyNotFound(err) {
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
		fields.AccountName:            ctx.AccountName,
		fc.EnvironmentSpecProjectName: projectName,
	}

	return d.environmentRepo.FindPaginated(ctx, d.environmentRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) findEnvironmentByTargetNs(ctx ConsoleContext, targetNs string) (*entities.Environment, error) {
	w, err := d.environmentRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:                ctx.AccountName,
		fc.EnvironmentSpecTargetNamespace: targetNs,
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
	env.Namespace = project.Spec.TargetNamespace

	env.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &env.Environment); err != nil {
		return nil, errors.NewE(err)
	}

	env.IncrementRecordVersion()

	if env.Spec.TargetNamespace == "" {
		env.Spec.TargetNamespace = d.getEnvironmentTargetNamespace(projectName, env.Name)
	}

	if env.Spec.Routing == nil {
		env.Spec.Routing = &crdsv1.EnvironmentRouting{
			Mode: crdsv1.EnvironmentRoutingModePrivate,
		}
	}

	env.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}
	env.LastUpdatedBy = env.CreatedBy

	env.AccountName = ctx.AccountName
	env.SyncStatus = t.GenSyncStatus(t.SyncActionApply, env.RecordVersion)

	if _, err := d.upsertEnvironmentResourceMapping(ResourceContext{ConsoleContext: ctx, ProjectName: env.ProjectName, EnvironmentName: env.Name}, &env); err != nil {
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

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeEnvironment, nenv.Name, PublishAdd)

	if _, err := d.iamClient.AddMembership(ctx, &iam.AddMembershipIn{
		UserId:       string(ctx.UserId),
		ResourceType: string(iamT.ResourceEnvironment),
		ResourceRef:  iamT.NewResourceRef(ctx.AccountName, iamT.ResourceEnvironment, nenv.Name),
		Role:         string(iamT.RoleResourceOwner),
	}); err != nil {
		d.logger.Errorf(err, "error while adding membership")
	}

	if err := d.applyK8sResource(ctx, nenv.ProjectName, &nenv.Environment, nenv.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return nenv, nil
}

func (d *domain) CloneEnvironment(ctx ConsoleContext, projectName string, sourceEnvName string, destinationEnvName string, displayName string, envRoutingMode crdsv1.EnvironmentRoutingMode) (*entities.Environment, error) {
	if err := d.canMutateResourcesInProject(ctx, projectName); err != nil {
		return nil, errors.NewE(err)
	}

	sourceEnv, err := d.findEnvironment(ctx, projectName, sourceEnvName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	destEnv := &entities.Environment{
		Environment: crdsv1.Environment{
			TypeMeta: sourceEnv.TypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:      destinationEnvName,
				Namespace: sourceEnv.Namespace,
			},
			Spec: crdsv1.EnvironmentSpec{
				ProjectName:     projectName,
				TargetNamespace: d.getEnvironmentTargetNamespace(projectName, destinationEnvName),
				Routing: &crdsv1.EnvironmentRouting{
					Mode: envRoutingMode,
				},
			},
		},
		AccountName: ctx.AccountName,
		ProjectName: projectName,
		ResourceMetadata: common.ResourceMetadata{
			DisplayName: displayName,
			CreatedBy: common.CreatedOrUpdatedBy{
				UserId:    ctx.UserId,
				UserName:  ctx.UserName,
				UserEmail: ctx.UserEmail,
			},
			LastUpdatedBy: common.CreatedOrUpdatedBy{
				UserId:    ctx.UserId,
				UserName:  ctx.UserName,
				UserEmail: ctx.UserEmail,
			},
		},
		SyncStatus: t.GenSyncStatus(t.SyncActionApply, 0),
	}

	if err := d.k8sClient.ValidateObject(ctx, &destEnv.Environment); err != nil {
		return nil, errors.NewE(err)
	}

	if _, err := d.iamClient.AddMembership(ctx, &iam.AddMembershipIn{
		UserId:       string(ctx.UserId),
		ResourceType: string(iamT.ResourceEnvironment),
		ResourceRef:  iamT.NewResourceRef(ctx.AccountName, iamT.ResourceEnvironment, destEnv.Spec.TargetNamespace),
		Role:         string(iamT.RoleResourceOwner),
	}); err != nil {
		d.logger.Errorf(err, "error while adding membership")
	}

	destEnv, err = d.environmentRepo.Create(ctx, destEnv)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if _, err := d.upsertEnvironmentResourceMapping(ResourceContext{ConsoleContext: ctx, ProjectName: sourceEnv.ProjectName, EnvironmentName: sourceEnv.Name}, sourceEnv); err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, sourceEnv.ProjectName, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: destEnv.Spec.TargetNamespace,
		},
	}, destEnv.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, sourceEnv.ProjectName, &destEnv.Environment, destEnv.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	resCtx := ResourceContext{
		ConsoleContext:  ctx,
		ProjectName:     sourceEnv.ProjectName,
		EnvironmentName: destEnv.Name,
	}

	filters := repos.Filter{
		fields.AccountName:     resCtx.AccountName,
		fields.ProjectName:     resCtx.ProjectName,
		fields.EnvironmentName: sourceEnvName,
	}

	apps, err := d.appRepo.Find(ctx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	secrets, err := d.secretRepo.Find(ctx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	configs, err := d.configRepo.Find(ctx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	routers, err := d.routerRepo.Find(ctx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	managedResources, err := d.mresRepo.Find(ctx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	resourceMetadata := func(dn string) common.ResourceMetadata {
		return common.ResourceMetadata{
			DisplayName: dn,
			CreatedBy: common.CreatedOrUpdatedBy{
				UserId:    ctx.UserId,
				UserName:  ctx.UserName,
				UserEmail: ctx.UserEmail,
			},
			LastUpdatedBy: common.CreatedOrUpdatedBy{
				UserId:    ctx.UserId,
				UserName:  ctx.UserName,
				UserEmail: ctx.UserEmail,
			},
		}
	}

	objectMeta := func(sourceMeta metav1.ObjectMeta, namespace string) metav1.ObjectMeta {
		sourceMeta.Namespace = namespace
		return sourceMeta
	}

	for i := range apps {
		if _, err := d.createAndApplyApp(resCtx, &entities.App{
			App: crdsv1.App{
				TypeMeta:   apps[i].TypeMeta,
				ObjectMeta: objectMeta(apps[i].ObjectMeta, destEnv.Spec.TargetNamespace),
				Spec:       apps[i].Spec,
			},
			AccountName:      ctx.AccountName,
			ProjectName:      projectName,
			EnvironmentName:  destEnv.Name,
			ResourceMetadata: resourceMetadata(apps[i].DisplayName),
			SyncStatus:       t.GenSyncStatus(t.SyncActionApply, 0),
		}); err != nil {
			return nil, err
		}
	}

	for i := range secrets {
		if _, err := d.createAndApplySecret(resCtx, &entities.Secret{
			Secret: corev1.Secret{
				TypeMeta:   secrets[i].TypeMeta,
				ObjectMeta: objectMeta(secrets[i].ObjectMeta, destEnv.Spec.TargetNamespace),
				Immutable:  secrets[i].Immutable,
				Data:       secrets[i].Data,
				StringData: secrets[i].StringData,
				Type:       secrets[i].Type,
			},
			AccountName:      ctx.AccountName,
			ProjectName:      projectName,
			EnvironmentName:  destEnv.Name,
			ResourceMetadata: resourceMetadata(secrets[i].DisplayName),
		}); err != nil {
			return nil, err
		}
	}

	for i := range configs {
		if _, err := d.createAndApplyConfig(resCtx, &entities.Config{
			ConfigMap: corev1.ConfigMap{
				TypeMeta:   configs[i].TypeMeta,
				ObjectMeta: objectMeta(configs[i].ObjectMeta, destEnv.Spec.TargetNamespace),
				Immutable:  configs[i].Immutable,
				Data:       configs[i].Data,
				BinaryData: configs[i].BinaryData,
			},
			AccountName:      ctx.AccountName,
			ProjectName:      projectName,
			EnvironmentName:  destEnv.Name,
			ResourceMetadata: resourceMetadata(configs[i].DisplayName),
		}); err != nil {
			return nil, err
		}
	}

	for i := range routers {
		if _, err := d.createAndApplyRouter(resCtx, &entities.Router{
			Router: crdsv1.Router{
				TypeMeta:   routers[i].TypeMeta,
				ObjectMeta: objectMeta(routers[i].ObjectMeta, destEnv.Spec.TargetNamespace),
				Spec:       routers[i].Spec,
				Enabled:    routers[i].Enabled,
			},
			AccountName:      ctx.AccountName,
			ProjectName:      projectName,
			EnvironmentName:  destEnv.Name,
			ResourceMetadata: resourceMetadata(routers[i].DisplayName),
		}); err != nil {
			return nil, err
		}
	}

	for i := range managedResources {
		spec := managedResources[i].Spec
		spec.ResourceName = genMresResourceName(destEnv.Name, managedResources[i].Name)
		if _, err := d.createAndApplyManagedResource(resCtx, &entities.ManagedResource{
			ManagedResource: crdsv1.ManagedResource{
				TypeMeta:   managedResources[i].TypeMeta,
				ObjectMeta: objectMeta(managedResources[i].ObjectMeta, destEnv.Spec.TargetNamespace),
				Spec:       spec,
				Enabled:    managedResources[i].Enabled,
			},
			AccountName:      ctx.AccountName,
			ProjectName:      projectName,
			EnvironmentName:  destEnv.Name,
			ResourceMetadata: resourceMetadata(managedResources[i].DisplayName),
		}); err != nil {
			return nil, err
		}
	}

	return destEnv, nil
}

func (d *domain) getEnvironmentTargetNamespace(projectName string, envName string) string {
	envNamespace := fmt.Sprintf("env-%s-%s", projectName, envName)
	hash := md5.Sum([]byte(envNamespace))
	return fmt.Sprintf("env-%s", hex.EncodeToString(hash[:]))
}

func (d *domain) UpdateEnvironment(ctx ConsoleContext, projectName string, env entities.Environment) (*entities.Environment, error) {
	if err := d.canMutateResourcesInProject(ctx, projectName); err != nil {
		return nil, errors.NewE(err)
	}

	env.Namespace = "trest"

	env.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &env.Environment); err != nil {
		return nil, errors.NewE(err)
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&env,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.EnvironmentSpec: env.Spec,
			},
		},
	)

	upEnv, err := d.environmentRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: env.Name,
		},
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}
	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeEnvironment, upEnv.Name, PublishUpdate)

	if err := d.applyK8sResource(ctx, upEnv.ProjectName, &upEnv.Environment, upEnv.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return upEnv, nil
}

func (d *domain) DeleteEnvironment(ctx ConsoleContext, projectName string, name string) error {
	if err := d.canMutateResourcesInProject(ctx, projectName); err != nil {
		return errors.NewE(err)
	}

	uenv, err := d.environmentRepo.Patch(
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

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeEnvironment, uenv.Name, PublishUpdate)

	if err := d.deleteK8sResource(ctx, uenv.ProjectName, &uenv.Environment); err != nil {
		if errors.Is(err, ErrNoClusterAttached) {
			return d.appRepo.DeleteById(ctx, uenv.Id)
		}
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) OnEnvironmentApplyError(ctx ConsoleContext, errMsg, namespace, name string, opts UpdateAndDeleteOpts) error {
	uenv, err := d.environmentRepo.Patch(
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

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeEnvironment, uenv.Name, PublishDelete)

	return errors.NewE(err)
}

func (d *domain) OnEnvironmentDeleteMessage(ctx ConsoleContext, env entities.Environment) error {
	err := d.environmentRepo.DeleteOne(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: env.Name,
		},
	)
	if err != nil {
		return errors.NewE(err)
	}

	if _, err = d.iamClient.RemoveResource(ctx, &iam.RemoveResourceIn{
		ResourceRef: iamT.NewResourceRef(ctx.AccountName, iamT.ResourceEnvironment, env.Name),
	}); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeEnvironment, env.Name, PublishDelete)
	return nil
}

func (d *domain) OnEnvironmentUpdateMessage(ctx ConsoleContext, env entities.Environment, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xenv, err := d.findEnvironment(ctx, env.Spec.ProjectName, env.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xenv == nil {
		return errors.Newf("no environment found")
	}

	recordVersion, err := d.MatchRecordVersion(env.Annotations, xenv.RecordVersion)
	if err != nil {
		return d.resyncK8sResource(ctx, xenv.ProjectName, xenv.SyncStatus.Action, &xenv.Environment, xenv.RecordVersion)
	}

	uenv, err := d.environmentRepo.PatchById(
		ctx,
		xenv.Id,
		common.PatchForSyncFromAgent(
			&env,
			recordVersion,
			status,
			common.PatchOpts{
				MessageTimestamp: opts.MessageTimestamp,
			}))
	if err != nil {
		return err
	}

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeEnvironment, uenv.Name, PublishUpdate)
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

	return d.resyncK8sResource(ctx, e.ProjectName, e.SyncStatus.Action, &e.Environment, e.RecordVersion)
}
