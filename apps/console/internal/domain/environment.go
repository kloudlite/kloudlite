package domain

import (
	"fmt"

	"github.com/kloudlite/api/common/fields"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	opConstants "github.com/kloudlite/operator/pkg/constants"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/operator/operators/resource-watcher/types"

	"github.com/kloudlite/api/constants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kloudlite/api/apps/console/internal/domain/ports"
	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
)

func (d *domain) cleanupEnvironment(ctx ConsoleContext, envName string) error {
	filter := repos.Filter{
		fields.AccountName:     ctx.AccountName,
		fields.EnvironmentName: envName,
	}

	if err := d.appRepo.DeleteMany(ctx, filter); err != nil {
		return errors.NewE(err)
	}

	if err := d.externalAppRepo.DeleteMany(ctx, filter); err != nil {
		return errors.NewE(err)
	}

	if err := d.secretRepo.DeleteMany(ctx, filter); err != nil {
		return errors.NewE(err)
	}

	if err := d.configRepo.DeleteMany(ctx, filter); err != nil {
		return errors.NewE(err)
	}

	if err := d.routerRepo.DeleteMany(ctx, filter); err != nil {
		return errors.NewE(err)
	}

	if err := d.importedMresRepo.DeleteMany(ctx, filter); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) findEnvironment(ctx ConsoleContext, name string) (*entities.Environment, error) {
	env, err := d.environmentRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if env == nil {
		return nil, errors.Newf("no environment with name (%s)", name)
	}
	return env, nil
}

func (d *domain) getClusterAttachedToEnvironment(ctx K8sContext, name string) (*string, error) {
	env, err := d.environmentRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.GetAccountName(),
		fields.MetadataName: name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	if env == nil {
		return nil, errors.Newf("no cluster attached to this environment")
	}

	return &env.ClusterName, nil
}

func (d *domain) envTargetNamespace(ctx ConsoleContext, envName string) (string, error) {
	env, err := d.findEnvironment(ctx, envName)
	if err != nil {
		return "", err
	}
	return env.Spec.TargetNamespace, nil
}

func (d *domain) GetEnvironment(ctx ConsoleContext, name string) (*entities.Environment, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetEnvironment); err != nil {
		return nil, errors.NewE(err)
	}

	return d.findEnvironment(ctx, name)
}

func (d *domain) ListEnvironments(ctx ConsoleContext, search map[string]repos.MatchFilter, pq repos.CursorPagination) (*repos.PaginatedRecord[*entities.Environment], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListEnvironments); err != nil {
		return nil, errors.NewE(err)
	}

	filter := repos.Filter{fields.AccountName: ctx.AccountName}

	return d.environmentRepo.FindPaginated(ctx, d.environmentRepo.MergeMatchFilters(filter, search), pq)
}

func (d *domain) listEnvironments(ctx ConsoleContext) ([]*entities.Environment, error) {
	return d.environmentRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{fields.AccountName: ctx.AccountName},
	})
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

func (d *domain) SetupDefaultEnvTemplate(ctx ConsoleContext) error {
	if d.envVars.DefaultEnvTemplateAccountName == "" && d.envVars.DefaultEnvTemplateName == "" {
		return nil
	}

	if _, err := d.CloneEnvTemplate(ctx, CloneEnvironmentTemplateArgs{
		SourceAccountName:  d.envVars.DefaultEnvTemplateAccountName,
		SourceEnvName:      d.envVars.DefaultEnvTemplateName,
		DestinationEnvName: d.envVars.DefaultEnvTemplateName,
		DisplayName:        "Default Environment",
		EnvRoutingMode:     crdsv1.EnvironmentRoutingModePublic,
	}); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) CreateEnvironment(ctx ConsoleContext, env entities.Environment) (*entities.Environment, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateEnvironment); err != nil {
		return nil, errors.NewE(err)
	}

	dEnvTempName := d.envVars.DefaultEnvTemplateName
	if dEnvTempName != "" && dEnvTempName == env.Name {
		return nil, fmt.Errorf("name already reserved by default environment template")
	}

	if env.ClusterName != "" {
		ownedBy, err := d.infraSvc.GetByokClusterOwnedBy(ctx, ports.IsClusterLabelsIn{
			UserId:      string(ctx.UserId),
			UserEmail:   ctx.UserEmail,
			UserName:    ctx.UserName,
			AccountName: ctx.AccountName,
			ClusterName: env.ClusterName,
		})
		if err != nil {
			return nil, errors.NewE(err)
		}

		if ownedBy != "" && ownedBy != string(ctx.UserId) {
			return nil, fmt.Errorf("it's owned cluster, but you are not the owner")
		}

		if env.Labels == nil {
			env.Labels = map[string]string{}
		}

		env.Labels[constants.ClusterLabelOwnedBy] = string(ctx.UserId)
	}

	env.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &env.Environment); err != nil {
		return nil, errors.NewE(err)
	}

	env.IncrementRecordVersion()

	if env.Spec.TargetNamespace == "" {
		env.Spec.TargetNamespace = d.getEnvironmentTargetNamespace(env.Name)
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

	nenv, err := d.environmentRepo.Create(ctx, &env)
	if err != nil {
		if d.environmentRepo.ErrAlreadyExists(err) {
			// TODO: better insights into error, when it is being caused by duplicated indexes
			return nil, errors.NewE(err)
		}
		return nil, errors.NewE(err)
	}

	if _, err := d.upsertEnvironmentResourceMapping(ResourceContext{ConsoleContext: ctx, EnvironmentName: env.Name}, &env); err != nil {
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

	if err := d.applyEnvironmentTargetNamespace(ctx, nenv); err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, nenv.Name, &nenv.Environment, nenv.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.syncImagePullSecretsToEnvironment(ctx, nenv.Name); err != nil {
		return nil, errors.NewE(err)
	}

	return nenv, nil
}

type CloneEnvironmentArgs struct {
	SourceEnvName      string
	DestinationEnvName string
	DisplayName        string
	EnvRoutingMode     crdsv1.EnvironmentRoutingMode
	ClusterName        string
}

func (d *domain) CloneEnvironment(ctx ConsoleContext, args CloneEnvironmentArgs) (*entities.Environment, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CloneEnvironment); err != nil {
		return nil, errors.NewE(err)
	}

	sourceEnv, err := d.findEnvironment(ctx, args.SourceEnvName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	destEnv := &entities.Environment{
		Environment: crdsv1.Environment{
			TypeMeta: sourceEnv.TypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:      args.DestinationEnvName,
				Namespace: sourceEnv.Namespace,
			},
			Spec: crdsv1.EnvironmentSpec{
				TargetNamespace: d.getEnvironmentTargetNamespace(args.DestinationEnvName),
				Routing: &crdsv1.EnvironmentRouting{
					Mode: args.EnvRoutingMode,
				},
			},
		},
		AccountName: ctx.AccountName,
		ClusterName: args.ClusterName,
		ResourceMetadata: common.ResourceMetadata{
			DisplayName: args.DisplayName,
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

	if _, err := d.upsertEnvironmentResourceMapping(ResourceContext{ConsoleContext: ctx, EnvironmentName: sourceEnv.Name}, sourceEnv); err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, destEnv.Name, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: destEnv.Spec.TargetNamespace,
			Labels: map[string]string{
				opConstants.KloudliteGatewayEnabledLabel: "true",
			},
		},
	}, destEnv.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyK8sResource(ctx, destEnv.Name, &destEnv.Environment, destEnv.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	resCtx := ResourceContext{
		ConsoleContext:  ctx,
		EnvironmentName: destEnv.Name,
	}

	filters := repos.Filter{
		fields.AccountName:     resCtx.AccountName,
		fields.EnvironmentName: args.SourceEnvName,
	}

	apps, err := d.appRepo.Find(ctx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	externalApps, err := d.externalAppRepo.Find(ctx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	secrets, err := d.secretRepo.Find(ctx, repos.Query{
		Filter: d.secretRepo.MergeMatchFilters(filters, map[string]repos.MatchFilter{
			fc.SecretFor: {
				MatchType: repos.MatchTypeExact,
				Exact:     nil,
			},
		}),
		Sort: nil,
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

	mresources, err := d.importedMresRepo.Find(ctx, repos.Query{
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
		appSpec := apps[i].Spec
		appSpec.Intercept = nil
		if _, err := d.createAndApplyApp(resCtx, &entities.App{
			App: crdsv1.App{
				TypeMeta:   apps[i].TypeMeta,
				ObjectMeta: objectMeta(apps[i].ObjectMeta, destEnv.Spec.TargetNamespace),
				Spec:       appSpec,
			},
			AccountName:      ctx.AccountName,
			EnvironmentName:  destEnv.Name,
			ResourceMetadata: resourceMetadata(apps[i].DisplayName),
			SyncStatus:       t.GenSyncStatus(t.SyncActionApply, 0),
		}); err != nil {
			return nil, err
		}
	}

	for i := range externalApps {
		externalAppSpec := externalApps[i].Spec
		externalAppSpec.Intercept = nil
		if _, err := d.createAndApplyExternalApp(resCtx, &entities.ExternalApp{
			ExternalApp: crdsv1.ExternalApp{
				TypeMeta:   externalApps[i].TypeMeta,
				ObjectMeta: objectMeta(externalApps[i].ObjectMeta, destEnv.Spec.TargetNamespace),
				Spec:       externalAppSpec,
			},
			AccountName:      ctx.AccountName,
			EnvironmentName:  destEnv.Name,
			ResourceMetadata: resourceMetadata(externalApps[i].DisplayName),
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
			EnvironmentName:  destEnv.Name,
			ResourceMetadata: resourceMetadata(routers[i].DisplayName),
		}); err != nil {
			return nil, err
		}
	}

	for i := range mresources {
		if _, err := d.createAndApplyImportedManagedResource(resCtx, CreateAndApplyImportedManagedResourceArgs{
			ImportedManagedResourceName: mresources[i].Name,
			ManagedResourceRefID:        mresources[i].ManagedResourceRef.ID,
		}); err != nil {
			return nil, err
		}
	}

	if err := d.syncImagePullSecretsToEnvironment(ctx, args.DestinationEnvName); err != nil {
		return nil, err
	}

	return destEnv, nil
}

func (d *domain) getEnvironmentTargetNamespace(envName string) string {
	return fmt.Sprintf("env-%s", envName)
}

func (d *domain) ArchiveEnvironmentsForCluster(ctx ConsoleContext, clusterName string) (bool, error) {
	filters := repos.Filter{
		fields.AccountName: ctx.AccountName,
		fields.ClusterName: clusterName,
	}

	envs, err := d.environmentRepo.Find(ctx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	if err != nil {
		return false, errors.NewE(err)
	}

	for i := range envs {
		patchForUpdate := repos.Document{
			fc.EnvironmentIsArchived: true,
		}
		patchFilter := repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.ClusterName:  clusterName,
			fields.MetadataName: envs[i].Name,
		}

		_, err := d.environmentRepo.Patch(ctx, patchFilter, patchForUpdate)
		if err != nil {
			return false, errors.NewE(err)
		}
	}

	return true, nil
}

func (d *domain) UpdateEnvironment(ctx ConsoleContext, env entities.Environment) (*entities.Environment, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateEnvironment); err != nil {
		return nil, errors.NewE(err)
	}

	env.Namespace = "non-empty-namespace"

	env.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &env.Environment); err != nil {
		return nil, errors.NewE(err)
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&env,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.EnvironmentSpecRouting: env.Spec.Routing,
				fc.EnvironmentSpecSuspend: env.Spec.Suspend,
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

	if err := d.applyK8sResource(ctx, upEnv.Name, &upEnv.Environment, upEnv.RecordVersion); err != nil {
		return nil, errors.NewE(err)
	}

	return upEnv, nil
}

func (d *domain) DeleteEnvironment(ctx ConsoleContext, name string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteEnvironment); err != nil {
		return errors.NewE(err)
	}

	uenv, err := d.environmentRepo.Patch(ctx, entities.EnvironmentDBFilter(ctx.AccountName, name), common.PatchForMarkDeletion())
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeEnvironment, uenv.Name, PublishUpdate)

	if uenv.IsArchived != nil && *uenv.IsArchived {
		if err := d.cleanupEnvironment(ctx, name); err != nil {
			return errors.NewE(err)
		}

		return d.environmentRepo.DeleteById(ctx, uenv.Id)
	}

	if err := d.deleteK8sResource(ctx, uenv.Name, &uenv.Environment); err != nil {
		if errors.Is(err, ErrNoClusterAttached) {
			return d.environmentRepo.DeleteById(ctx, uenv.Id)
		}
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) OnEnvironmentApplyError(ctx ConsoleContext, errMsg, namespace, name string, opts UpdateAndDeleteOpts) error {
	uenv, err := d.environmentRepo.Patch(
		ctx,
		repos.Filter{
			fields.AccountName:       ctx.AccountName,
			fields.MetadataName:      name,
			fields.MetadataNamespace: namespace,
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
	if err := d.cleanupEnvironment(ctx, env.Name); err != nil {
		return errors.NewE(err)
	}

	if err := d.environmentRepo.DeleteOne(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: env.Name,
		},
	); err != nil {
		return errors.NewE(err)
	}

	if _, err := d.iamClient.RemoveResource(ctx, &iam.RemoveResourceIn{
		ResourceRef: iamT.NewResourceRef(ctx.AccountName, iamT.ResourceEnvironment, env.Name),
	}); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishConsoleEvent(ctx, entities.ResourceTypeEnvironment, env.Name, PublishDelete)
	return nil
}

func (d *domain) OnEnvironmentUpdateMessage(ctx ConsoleContext, env entities.Environment, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xenv, err := d.findEnvironment(ctx, env.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xenv == nil {
		return errors.Newf("no environment found")
	}

	recordVersion, err := d.MatchRecordVersion(env.Annotations, xenv.RecordVersion)
	if err != nil {
		return d.resyncK8sResource(ctx, xenv.Name, xenv.SyncStatus.Action, &xenv.Environment, xenv.RecordVersion)
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

func (d *domain) applyEnvironmentTargetNamespace(ctx ConsoleContext, env *entities.Environment) error {
	if err := d.applyK8sResource(ctx, env.Name, &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		ObjectMeta: metav1.ObjectMeta{
			Name: env.Spec.TargetNamespace,
			Labels: map[string]string{
				constants.EnvNameKey: env.Name,
			},
		},
	}, env.RecordVersion); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) ResyncEnvironment(ctx ConsoleContext, name string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateEnvironment); err != nil {
		return errors.NewE(err)
	}

	e, err := d.findEnvironment(ctx, name)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.resyncK8sResource(ctx, e.Name, t.SyncActionApply, &corev1.Namespace{
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

	return d.resyncK8sResource(ctx, e.Name, e.SyncStatus.Action, &e.Environment, e.RecordVersion)
}
