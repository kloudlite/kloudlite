package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kloudlite/api/pkg/kv"

	"github.com/kloudlite/api/pkg/logging"

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

	projectRepo     repos.DbRepo[*entities.Project]
	environmentRepo repos.DbRepo[*entities.Environment]
	vpnDeviceRepo   repos.DbRepo[*entities.VPNDevice]

	appRepo         repos.DbRepo[*entities.App]
	configRepo      repos.DbRepo[*entities.Config]
	secretRepo      repos.DbRepo[*entities.Secret]
	routerRepo      repos.DbRepo[*entities.Router]
	mresRepo        repos.DbRepo[*entities.ManagedResource]
	pullSecretsRepo repos.DbRepo[*entities.ImagePullSecret]

	envVars *env.Env

	resourceEventPublisher ResourceEventPublisher
	consoleCacheStore      kv.BinaryDataRepo
	resourceMappingRepo    repos.DbRepo[*entities.ResourceMapping]
	pmsRepo                repos.DbRepo[*entities.ProjectManagedService]
}

// GetSecretEntries implements Domain.
func (*domain) GetSecretEntries(ctx ResourceContext, keyrefs []SecretKeyRef) ([]*SecretKeyValueRef, error) {
	panic("unimplemented")
}

func errAlreadyMarkedForDeletion(label, namespace, name string) error {
	return errors.Newf(
		"%s (namespace=%s, name=%s) already marked for deletion",
		label,
		namespace,
		name,
	)
}

var ErrNoClusterAttached = errors.New("cluster not attached")

type K8sContext interface {
	context.Context
	GetAccountName() string
}

func (d *domain) applyK8sResource(ctx K8sContext, projectName string, obj client.Object, recordVersion int) error {
	clusterName, err := d.getClusterAttachedToProject(ctx, projectName)
	if err != nil {
		return errors.NewE(err)
	}

	if clusterName == nil {
		d.logger.Infof("skipping apply of k8s resource %s/%s, cluster name not provided", obj.GetNamespace(), obj.GetName())
		return nil
	}

	if obj.GetObjectKind().GroupVersionKind().Empty() {
		return errors.Newf("object GVK is not set, can not apply")
	}

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
		AccountName: ctx.GetAccountName(),
		ClusterName: *clusterName,
		Action:      t.ActionApply,
		Object:      m,
	})
	if err != nil {
		return errors.NewE(err)
	}

	err = d.producer.Produce(ctx, msgTypes.ProduceMsg{
		Subject: common.GetTenantClusterMessagingTopic(ctx.GetAccountName(), *clusterName),
		Payload: b,
	})
	return errors.NewE(err)
}

func (d *domain) deleteK8sResource(ctx K8sContext, projectName string, obj client.Object) error {
	clusterName, err := d.getClusterAttachedToProject(ctx, projectName)
	if err != nil {
		return errors.NewE(err)
	}

	if clusterName == nil {
		d.logger.Infof("skipping delete of k8s resource %s/%s, cluster name not provided", obj.GetNamespace(), obj.GetName())
		return ErrNoClusterAttached
	}

	if obj.GetObjectKind().GroupVersionKind().Empty() {
		return errors.Newf("object GVK is not set, can not apply")
	}

	m, err := fn.K8sObjToMap(obj)
	if err != nil {
		return errors.NewE(err)
	}
	b, err := json.Marshal(t.AgentMessage{
		AccountName: ctx.GetAccountName(),
		ClusterName: *clusterName,
		Action:      t.ActionDelete,
		Object:      m,
	})
	if err != nil {
		return errors.NewE(err)
	}

	err = d.producer.Produce(ctx, msgTypes.ProduceMsg{
		Subject: common.GetTenantClusterMessagingTopic(ctx.GetAccountName(), *clusterName),
		Payload: b,
	})

	return errors.NewE(err)
}

func (d *domain) resyncK8sResource(ctx K8sContext, projectName string, action types.SyncAction, obj client.Object, rv int) error {
	switch action {
	case types.SyncActionApply:
		{
			return d.applyK8sResource(ctx, projectName, obj, rv)
		}
	case types.SyncActionDelete:
		{
			return d.deleteK8sResource(ctx, projectName, obj)
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

func (d *domain) canMutateResourcesInProject(ctx ConsoleContext, projectName string) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, projectName),
		},
		Action: string(iamT.MutateResourcesInProject),
	})
	if err != nil {
		return errors.NewE(err)
	}
	if !co.Status {
		return errors.Newf("unauthorized to mutate resources in project (%s)", projectName)
	}
	return nil
}

func (d *domain) canReadResourcesInProject(ctx ConsoleContext, projectName string) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, projectName),
		},
		Action: string(iamT.GetProject),
	})
	if err != nil {
		return errors.NewE(err)
	}
	if !co.Status {
		return errors.Newf("unauthorized to read resources in project (%s)", projectName)
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

func (d *domain) checkEnvironmentAccess(ctx ResourceContext, action iamT.Action) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceProject, ctx.ProjectName),
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceEnvironment, ctx.EnvironmentName),
		},
		Action: string(action),
	})
	if err != nil {
		return errors.NewE(err)
	}

	if !co.Status {
		return errors.Newf("unauthorized to access environment %q", ctx.EnvironmentName)
	}
	return nil
}

func (d *domain) canMutateResourcesInEnvironment(ctx ResourceContext) error {
	return d.checkEnvironmentAccess(ctx, iamT.MutateResourcesInEnvironment)
}

func (d *domain) canReadResourcesInEnvironment(ctx ResourceContext) error {
	return d.checkEnvironmentAccess(ctx, iamT.ReadResourcesInEnvironment)
}

func (d *domain) canPerformActionInAccount(ctx ConsoleContext, action iamT.Action) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
		},
		Action: string(action),
	})
	if err != nil {
		return errors.NewE(err)
	}
	if !co.Status {
		return errors.Newf("unauthorized to perform action %q in account %q", action, ctx.AccountName)
	}
	return nil
}

func (d *domain) canPerformActionInDevice(ctx ConsoleContext, action iamT.Action, devName string) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceVPNDevice, devName),
		},
		Action: string(action),
	})
	if err != nil {
		return errors.NewE(err)
	}
	if !co.Status {
		return errors.Newf("unauthorized to perform action %q in device %q", action, devName)
	}
	return nil
}

func cloneResource[T repos.Entity](ctx ResourceContext, d *domain, repoName repos.DbRepo[T], resource T, obj client.Object) error {
	_, err := repoName.Create(ctx, resource)
	if err != nil {
		if !repoName.ErrAlreadyExists(err) {
			return errors.NewE(err)
		}
	}
	if err := d.applyK8sResource(ctx, ctx.ProjectName, obj, 0); err != nil {
		return errors.NewE(err)
	}
	return nil
}

type ConsoleCacheStore kv.BinaryDataRepo

var Module = fx.Module("domain",
	fx.Provide(func(
		k8sClient k8s.Client,

		producer MessageDispatcher,

		iamClient iam.IAMClient,
		infraClient infra.InfraClient,

		projectRepo repos.DbRepo[*entities.Project],
		environmentRepo repos.DbRepo[*entities.Environment],
		appRepo repos.DbRepo[*entities.App],
		configRepo repos.DbRepo[*entities.Config],
		secretRepo repos.DbRepo[*entities.Secret],
		routerRepo repos.DbRepo[*entities.Router],
		mresRepo repos.DbRepo[*entities.ManagedResource],
		ipsRepo repos.DbRepo[*entities.ImagePullSecret],
		pmsRepo repos.DbRepo[*entities.ProjectManagedService],
		resourceMappingRepo repos.DbRepo[*entities.ResourceMapping],
		vpnDeviceRepo repos.DbRepo[*entities.VPNDevice],

		logger logging.Logger,
		resourceEventPublisher ResourceEventPublisher,

		ev *env.Env,

		consoleCacheStore ConsoleCacheStore,
	) Domain {
		return &domain{
			k8sClient: k8sClient,

			producer: producer,

			iamClient:   iamClient,
			infraClient: infraClient,
			logger:      logger,

			projectRepo:         projectRepo,
			environmentRepo:     environmentRepo,
			appRepo:             appRepo,
			configRepo:          configRepo,
			routerRepo:          routerRepo,
			secretRepo:          secretRepo,
			mresRepo:            mresRepo,
			pullSecretsRepo:     ipsRepo,
			resourceMappingRepo: resourceMappingRepo,
			vpnDeviceRepo:       vpnDeviceRepo,

			envVars: ev,

			resourceEventPublisher: resourceEventPublisher,
			consoleCacheStore:      consoleCacheStore,
			pmsRepo:                pmsRepo,
		}
	}),
)
