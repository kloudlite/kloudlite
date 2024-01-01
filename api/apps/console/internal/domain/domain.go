package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kloudlite/api/pkg/logging"
	"strconv"

	"github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/common"

	"github.com/kloudlite/api/pkg/messaging"
	msgTypes "github.com/kloudlite/api/pkg/messaging/types"
	"github.com/kloudlite/api/pkg/types"
	"github.com/kloudlite/operator/pkg/constants"

	t "github.com/kloudlite/api/apps/tenant-agent/types"
	"go.uber.org/fx"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/apps/console/internal/env"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/infra"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/repos"
)

type MessageDispatcher messaging.Producer

type domain struct {
	k8sClient k8s.Client
	logger    logging.Logger

	producer MessageDispatcher

	iamClient   iam.IAMClient
	infraClient infra.InfraClient

	projectRepo   repos.DbRepo[*entities.Project]
	workspaceRepo repos.DbRepo[*entities.Workspace]

	appRepo         repos.DbRepo[*entities.App]
	configRepo      repos.DbRepo[*entities.Config]
	secretRepo      repos.DbRepo[*entities.Secret]
	routerRepo      repos.DbRepo[*entities.Router]
	mresRepo        repos.DbRepo[*entities.ManagedResource]
	pullSecretsRepo repos.DbRepo[*entities.ImagePullSecret]

	envVars *env.Env

	resourceEventPublisher ResourceEventPublisher
}

func errAlreadyMarkedForDeletion(label, namespace, name string) error {
	return errors.Newf(
		"%s (namespace=%s, name=%s) already marked for deletion",
		label,
		namespace,
		name,
	)
}

func (d *domain) applyK8sResource(ctx ConsoleContext, obj client.Object, recordVersion int) error {
	ann := obj.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string, 1)
	}
	ann[constants.RecordVersionKey] = fmt.Sprintf("%d", recordVersion)
	obj.SetAnnotations(ann)

	m, err := fn.K8sObjToMap(obj)
	if err != nil {
		return errors.NewE(err)
	}
	b, err := json.Marshal(t.AgentMessage{
		AccountName: ctx.AccountName,
		ClusterName: ctx.ClusterName,
		Action:      t.ActionApply,
		Object:      m,
	})
	if err != nil {
		return errors.NewE(err)
	}

	err = d.producer.Produce(ctx, msgTypes.ProduceMsg{
		Subject: common.GetTenantClusterMessagingTopic(ctx.AccountName, ctx.ClusterName),
		Payload: b,
	})
	return errors.NewE(err)
}

func (d *domain) deleteK8sResource(ctx ConsoleContext, obj client.Object) error {
	m, err := fn.K8sObjToMap(obj)
	if err != nil {
		return errors.NewE(err)
	}
	b, err := json.Marshal(t.AgentMessage{
		AccountName: ctx.AccountName,
		ClusterName: ctx.ClusterName,
		Action:      t.ActionDelete,
		Object:      m,
	})
	if err != nil {
		return errors.NewE(err)
	}

	err = d.producer.Produce(ctx, msgTypes.ProduceMsg{
		Subject: common.GetTenantClusterMessagingTopic(ctx.AccountName, ctx.ClusterName),
		Payload: b,
	})

	return errors.NewE(err)
}

func (d *domain) resyncK8sResource(ctx ConsoleContext, action types.SyncAction, obj client.Object, rv int) error {
	switch action {
	case types.SyncActionApply:
		{
			return d.applyK8sResource(ctx, obj, rv)
		}
	case types.SyncActionDelete:
		{
			return d.deleteK8sResource(ctx, obj)
		}
	default:
		{
			return errors.Newf("unknown sync action %q", action)
		}
	}
}

func (d *domain) parseRecordVersionFromAnnotations(annotations map[string]string) (int, error) {
	annotatedVersion, ok := annotations[constants.RecordVersionKey]
	if !ok {
		return 0, errors.Newf("no annotation with record version key (%s), found on the resource", constants.RecordVersionKey)
	}

	annVersion, err := strconv.ParseInt(annotatedVersion, 10, 32)
	if err != nil {
		return 0, errors.NewE(err)
	}

	return int(annVersion), nil
}

func (d *domain) MatchRecordVersion(annotations map[string]string, rv int) error {
	annVersion, err := d.parseRecordVersionFromAnnotations(annotations)
	if err != nil {
		return errors.NewE(err)
	}

	if annVersion != rv {
		return errors.Newf("record version mismatch, expected %d, got %d", rv, annVersion)
	}

	return nil
}

func (d *domain) canMutateResourcesInProject(ctx ConsoleContext, targetNamespace string) error {
	prj, err := d.findProjectByTargetNs(ctx, targetNamespace)
	if err != nil {
		return errors.NewE(err)
	}

	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, prj.Name),
		},
		Action: string(iamT.MutateResourcesInProject),
	})
	if err != nil {
		return errors.NewE(err)
	}
	if !co.Status {
		return errors.Newf("unauthorized to mutate resources in project %q", prj.Name)
	}
	return nil
}

func (d *domain) canMutateResourcesInWorkspace(ctx ConsoleContext, targetNamespace string) error {
	ws, err := d.findWorkspaceByTargetNs(ctx, targetNamespace)
	if err != nil {
		return errors.NewE(err)
	}

	wsp, err := d.findWorkspace(ctx, ws.Namespace, ws.Name)
	if err != nil {
		return errors.NewE(err)
	}

	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, wsp.Spec.ProjectName),
		},
		Action: string(iamT.MutateResourcesInProject),
	})
	if err != nil {
		return errors.NewE(err)
	}
	if !co.Status {
		return errors.Newf("unauthorized to mutate resources in workspace %q", wsp.Name)
	}
	return nil
}

func (d *domain) canReadResourcesInWorkspace(ctx ConsoleContext, targetNamespace string) error {
	ws, err := d.findWorkspaceByTargetNs(ctx, targetNamespace)
	if err != nil {
		return errors.NewE(err)
	}

	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, ws.Spec.ProjectName),
		},
		Action: string(iamT.GetProject),
	})
	if err != nil {
		return errors.NewE(err)
	}
	if !co.Status {
		return errors.Newf("unauthorized to read resources in project %q", ws.Spec.ProjectName)
	}
	return nil
}

func (d *domain) canReadResourcesInProject(ctx ConsoleContext, targetNamespace string) error {
	prj, err := d.findProjectByTargetNs(ctx, targetNamespace)
	if err != nil {
		return errors.NewE(err)
	}

	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, prj.Name),
		},
		Action: string(iamT.GetProject),
	})
	if err != nil {
		return errors.NewE(err)
	}
	if !co.Status {
		return errors.Newf("unauthorized to read resources in project %q", prj.Name)
	}
	return nil
}

func (d *domain) canMutateSecretsInAccount(ctx context.Context, userId string, accountName string) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: userId,
		ResourceRefs: []string{
			iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
		},
		Action: string(iamT.CreateSecretsInAccount),
	})
	if err != nil {
		return errors.NewE(err)
	}
	if !co.Status {
		return errors.Newf("unauthorized to mutate secrets in account %q", accountName)
	}
	return nil
}

func (d *domain) canReadSecretsFromAccount(ctx context.Context, userId string, accountName string) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: userId,
		ResourceRefs: []string{
			iamT.NewResourceRef(accountName, iamT.ResourceAccount, accountName),
		},
		Action: string(iamT.ReadSecretsFromAccount),
	})
	if err != nil {
		return errors.NewE(err)
	}
	if !co.Status {
		return errors.Newf("unauthorized to read secrets from account  %q", accountName)
	}
	return nil
}

func (d *domain) checkProjectAccess(ctx ConsoleContext, projectName string, action iamT.Action) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, projectName),
		},
		Action: string(action),
	})
	if err != nil {
		return errors.NewE(err)
	}

	if !co.Status {
		return errors.Newf("unauthorized to access project %q", projectName)
	}
	return nil
}

func (d *domain) checkWorkspaceAccess(ctx ConsoleContext, projectName string, workspaceName string, action iamT.Action) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, projectName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceWorkspace, workspaceName),
		},
		Action: string(action),
	})
	if err != nil {
		return errors.NewE(err)
	}

	if !co.Status {
		return errors.Newf("unauthorized to access workspace %q", workspaceName)
	}
	return nil
}

func (d *domain) checkEnvironmentAccess(ctx ConsoleContext, projectName string, environmentName string, action iamT.Action) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, projectName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceEnvironment, environmentName),
		},
		Action: string(action),
	})
	if err != nil {
		return errors.NewE(err)
	}

	if !co.Status {
		return errors.Newf("unauthorized to access environment %q", environmentName)
	}
	return nil
}

func (d *domain) canMutateResourcesInWorkspaceOrEnv(ctx ConsoleContext, projectName string, workspace *entities.Workspace) error {
	if workspace.Spec.IsEnvironment != nil && *workspace.Spec.IsEnvironment {
		return d.checkEnvironmentAccess(ctx, projectName, workspace.Name, iamT.MutateResourcesInEnvironment)
	}
	return d.checkWorkspaceAccess(ctx, projectName, workspace.Name, iamT.MutateResourcesInWorkspace)
}

func (d *domain) canReadResourcesInWorkspaceOrEnv(ctx ConsoleContext, projectName string, workspace *entities.Workspace) error {
	if workspace.Spec.IsEnvironment != nil && *workspace.Spec.IsEnvironment {
		return d.checkEnvironmentAccess(ctx, projectName, workspace.Name, iamT.ReadResourcesInEnvironment)
	}
	return d.checkWorkspaceAccess(ctx, projectName, workspace.Name, iamT.ReadResourcesInWorkspace)
}

var Module = fx.Module("domain",
	fx.Provide(func(
		k8sClient k8s.Client,

		producer MessageDispatcher,

		iamClient iam.IAMClient,
		infraClient infra.InfraClient,

		projectRepo repos.DbRepo[*entities.Project],
		workspaceRepo repos.DbRepo[*entities.Workspace],

		appRepo repos.DbRepo[*entities.App],
		configRepo repos.DbRepo[*entities.Config],
		secretRepo repos.DbRepo[*entities.Secret],
		routerRepo repos.DbRepo[*entities.Router],
		mresRepo repos.DbRepo[*entities.ManagedResource],
		ipsRepo repos.DbRepo[*entities.ImagePullSecret],
		logger logging.Logger,
		resourceEventPublisher ResourceEventPublisher,

		ev *env.Env,
	) Domain {

		return &domain{
			k8sClient: k8sClient,

			producer: producer,

			iamClient:   iamClient,
			infraClient: infraClient,
			logger:      logger,

			projectRepo:     projectRepo,
			workspaceRepo:   workspaceRepo,
			appRepo:         appRepo,
			configRepo:      configRepo,
			routerRepo:      routerRepo,
			secretRepo:      secretRepo,
			mresRepo:        mresRepo,
			pullSecretsRepo: ipsRepo,

			envVars: ev,

			resourceEventPublisher: resourceEventPublisher,
		}
	}),
)
