package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CloneEnvironmentTemplateArgs struct {
	SourceAccountName  string
	SourceEnvName      string
	DestinationEnvName string
	DisplayName        string
	EnvRoutingMode     crdsv1.EnvironmentRoutingMode
}

func (d *domain) CloneEnvTemplate(ctx ConsoleContext, args CloneEnvironmentTemplateArgs) (*entities.Environment, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CloneEnvironment); err != nil {
		return nil, errors.NewE(err)
	}

	srcCtx := NewConsoleContext(ctx, "sys-user:console-resource-updater", args.SourceAccountName)

	sourceEnv, err := d.findEnvironment(srcCtx, args.SourceEnvName)
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
		ClusterName: "",
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

	resCtx := ResourceContext{
		ConsoleContext:  ctx,
		EnvironmentName: destEnv.Name,
	}

	filters := repos.Filter{
		fields.AccountName:     args.SourceAccountName,
		fields.EnvironmentName: args.SourceEnvName,
	}

	apps, err := d.appRepo.Find(srcCtx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	externalApps, err := d.externalAppRepo.Find(srcCtx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	secrets, err := d.secretRepo.Find(srcCtx, repos.Query{
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

	configs, err := d.configRepo.Find(srcCtx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	routers, err := d.routerRepo.Find(srcCtx, repos.Query{
		Filter: filters,
		Sort:   nil,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	mresources, err := d.importedMresRepo.Find(srcCtx, repos.Query{
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

	// if err := d.syncImagePullSecretsToEnvironment(ctx, args.DestinationEnvName); err != nil {
	// 	return nil, err
	// }

	return destEnv, nil
}
