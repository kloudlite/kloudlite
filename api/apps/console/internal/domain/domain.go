package domain

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kloudlite/api/pkg/grpc"
	"github.com/kloudlite/api/pkg/kv"

	"github.com/kloudlite/api/pkg/logging"

	mo_errors "github.com/kloudlite/api/apps/message-office/errors"
	"github.com/kloudlite/api/pkg/errors"

	"github.com/kloudlite/api/common"

	"github.com/kloudlite/api/pkg/messaging"
	msgTypes "github.com/kloudlite/api/pkg/messaging/types"
	"github.com/kloudlite/api/pkg/types"

	"github.com/kloudlite/api/constants"

	platform_edge "github.com/kloudlite/api/apps/message-office/protobufs/platform-edge"
	t "github.com/kloudlite/api/apps/tenant-agent/types"
	"go.uber.org/fx"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kloudlite/api/apps/console/internal/domain/ports"
	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/apps/console/internal/env"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/k8s"
	"github.com/kloudlite/api/pkg/repos"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type MessageDispatcher messaging.Producer

type domain struct {
	k8sClient k8s.Client
	logger    logging.Logger

	producer MessageDispatcher

	iamClient          iam.IAMClient
	infraSvc           ports.InfraService
	platformEdgeClient platform_edge.PlatformEdgeClient
	AccountsSvc

	environmentRepo repos.DbRepo[*entities.Environment]

	appRepo          repos.DbRepo[*entities.App]
	externalAppRepo  repos.DbRepo[*entities.ExternalApp]
	configRepo       repos.DbRepo[*entities.Config]
	secretRepo       repos.DbRepo[*entities.Secret]
	routerRepo       repos.DbRepo[*entities.Router]
	mresRepo         repos.DbRepo[*entities.ManagedResource]
	importedMresRepo repos.DbRepo[*entities.ImportedManagedResource]
	pullSecretsRepo  repos.DbRepo[*entities.ImagePullSecret]

	registryImageRepo repos.DbRepo[*entities.RegistryImage]

	serviceBindingRepo        repos.DbRepo[*entities.ServiceBinding]
	clusterManagedServiceRepo repos.DbRepo[*entities.ClusterManagedService]

	envVars *env.Env

	resourceEventPublisher ResourceEventPublisher
	consoleCacheStore      kv.BinaryDataRepo
	resourceMappingRepo    repos.DbRepo[*entities.ResourceMapping]
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

func addTrackingId(obj client.Object, id repos.ID) {
	ann := obj.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string, 1)
	}
	ann[constants.ObservabilityTrackingKey] = string(id)
	obj.SetAnnotations(ann)

	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string, 1)
	}

	labels[constants.ObservabilityTrackingKey] = string(id)
	obj.SetLabels(labels)
}

type K8sContext interface {
	context.Context
	GetUserId() repos.ID
	GetUserEmail() string
	GetUserName() string
	GetAccountName() string
}

func (d *domain) applyK8sResourceOnCluster(ctx K8sContext, clusterName string, obj client.Object, recordVersion int) error {
	if clusterName == "" {
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
		// ClusterName: clusterName,
		Action: t.ActionApply,
		Object: m,
	})
	if err != nil {
		return errors.NewE(err)
	}

	subject := common.SendToAgentSubjectName(ctx.GetAccountName(), clusterName, obj.GetObjectKind().GroupVersionKind().String(), obj.GetNamespace(), obj.GetName())

	err = d.producer.Produce(ctx, msgTypes.ProduceMsg{
		Subject: subject,
		Payload: b,
	})
	return errors.NewE(err)
}

func (d *domain) applyK8sResource(ctx K8sContext, envName string, obj client.Object, recordVersion int) error {
	var dispatchAddr struct {
		AccountName string
		ClusterName string
	}

	clusterName, err := d.getClusterAttachedToEnvironment(ctx, envName)
	if err != nil {
		return errors.NewE(err)
	}

	if clusterName == nil || *clusterName == "" {
		d.logger.Infof("skipping apply of k8s resource %s/%s, cluster name not provided", obj.GetNamespace(), obj.GetName())
		return nil
	}

	switch *clusterName {
	case "__kloudlite_enabled_cluster":
		{
			allocatedEdge, err := d.platformEdgeClient.GetAllocatedPlatformEdgeCluster(ctx, &platform_edge.GetAllocatedPlatformEdgeClusterIn{AccountName: ctx.GetAccountName()})
			if err != nil {
				gErr := grpc.ParseErr(err)
				if gErr == nil {
					return errors.NewEf(err, "failed to get allocated edge cluster")
				}

				if gErr.GetMessage() != mo_errors.ErrEdgeClusterNotAllocated.Error() {
					return errors.NewEf(err, "failed to get allocated edge cluster")
				}

				// INFO: not allocated, allocating a new one
				accountRegion, err := d.GetAccountRegion(ctx, string(ctx.GetUserId()), ctx.GetAccountName())
				if err != nil {
					return errors.NewEf(err, "failed to get account region")
				}

				allocatedEdge, err = d.platformEdgeClient.AllocatePlatformEdgeCluster(ctx, &platform_edge.AllocatePlatformEdgeClusterIn{
					Region:      accountRegion,
					AccountName: ctx.GetAccountName(),
				})
				if err != nil {
					return errors.NewEf(err, "failed to allocate platform edge cluster")
				}

				if err := d.infraSvc.EnsureGlobalVPNConnection(ctx, ports.EnsureGlobalVPNConnectionIn{
					UserId:        string(ctx.GetUserId()),
					UserEmail:     ctx.GetUserEmail(),
					UserName:      ctx.GetUserName(),
					AccountName:   ctx.GetAccountName(),
					ClusterName:   allocatedEdge.ClusterName,
					GlobalVPNName: "default",

					DispatchAddrAccountName: allocatedEdge.OwnedByAccount,
					DispatchAddrClusterName: allocatedEdge.ClusterName,
				}); err != nil {
					return errors.NewEf(err, "failed to ensure global vpn connection")
				}
			}

			dispatchAddr.AccountName = allocatedEdge.OwnedByAccount
			dispatchAddr.ClusterName = allocatedEdge.ClusterName
		}
	default:
		{
			dispatchAddr.AccountName = ctx.GetAccountName()
			dispatchAddr.ClusterName = *clusterName
		}
	}

	if obj.GetObjectKind().GroupVersionKind().Empty() {
		return errors.Newf("object GVK is not set, can not apply")
	}

	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string, 2)
	}
	labels[constants.AccountNameKey] = ctx.GetAccountName()
	labels[constants.EnvNameKey] = envName
	obj.SetLabels(labels)

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
		Action:      t.ActionApply,
		Object:      m,
	})
	if err != nil {
		return errors.NewE(err)
	}

	subject := common.SendToAgentSubjectName(dispatchAddr.AccountName, dispatchAddr.ClusterName, obj.GetObjectKind().GroupVersionKind().String(), obj.GetNamespace(), obj.GetName())

	err = d.producer.Produce(ctx, msgTypes.ProduceMsg{
		Subject: subject,
		Payload: b,
	})
	return errors.NewE(err)
}

type ApplyK8sResourceArgs struct {
	ClusterName   string
	Object        client.Object
	RecordVersion int

	Dispatcher MessageDispatcher
}

func applyK8sResource(ctx K8sContext, args ApplyK8sResourceArgs) error {
	if args.ClusterName == "" {
		// d.logger.Infof("skipping apply of k8s resource %s/%s, cluster name not provided", args.Object.GetNamespace(), args.Object.GetName())
		return nil
	}

	if args.Object.GetObjectKind().GroupVersionKind().Empty() {
		return errors.Newf("args.Objectect GVK is not set, can not apply")
	}

	ann := args.Object.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string, 1)
	}
	ann[constants.RecordVersionKey] = fmt.Sprintf("%d", args.RecordVersion)
	args.Object.SetAnnotations(ann)

	m, err := fn.K8sObjToMap(args.Object)
	if err != nil {
		return errors.NewE(err)
	}
	b, err := json.Marshal(t.AgentMessage{
		AccountName: ctx.GetAccountName(),
		Action:      t.ActionApply,
		Object:      m,
	})
	if err != nil {
		return errors.NewE(err)
	}

	subject := common.SendToAgentSubjectName(ctx.GetAccountName(), args.ClusterName, args.Object.GetObjectKind().GroupVersionKind().String(), args.Object.GetNamespace(), args.Object.GetName())

	err = args.Dispatcher.Produce(ctx, msgTypes.ProduceMsg{Subject: subject, Payload: b})
	return errors.NewE(err)
}

func (d *domain) restartK8sResource(ctx K8sContext, projectName string, namespace string, labels map[string]string) error {
	var dispatchAddr struct {
		AccountName string
		ClusterName string
	}

	clusterName, err := d.getClusterAttachedToEnvironment(ctx, projectName)
	if err != nil {
		return errors.NewE(err)
	}

	if clusterName == nil || *clusterName == "" {
		return nil
	}

	switch *clusterName {
	case "__kloudlite_enabled_cluster":
		{
			allocatedEdge, err := d.platformEdgeClient.GetAllocatedPlatformEdgeCluster(ctx, &platform_edge.GetAllocatedPlatformEdgeClusterIn{AccountName: ctx.GetAccountName()})
			if err != nil {
				return errors.NewEf(err, "failed to get allocated edge cluster")
			}

			dispatchAddr.AccountName = allocatedEdge.OwnedByAccount
			dispatchAddr.ClusterName = allocatedEdge.ClusterName
		}
	default:
		{

			dispatchAddr.AccountName = ctx.GetAccountName()
			dispatchAddr.ClusterName = *clusterName
		}
	}

	obj := unstructured.Unstructured{
		Object: map[string]any{
			"metadata": map[string]any{
				"namespace": namespace,
				"labels":    labels,
			},
		},
	}

	b, err := json.Marshal(t.AgentMessage{
		AccountName: ctx.GetAccountName(),
		Action:      t.ActionRestart,
		Object:      obj.Object,
	})
	if err != nil {
		return errors.NewE(err)
	}

	subject := common.SendToAgentSubjectName(dispatchAddr.AccountName, dispatchAddr.ClusterName, obj.GetObjectKind().GroupVersionKind().String(), obj.GetNamespace(), obj.GetName())

	err = d.producer.Produce(ctx, msgTypes.ProduceMsg{
		Subject: subject,
		Payload: b,
	})
	return errors.NewE(err)
}

func (d *domain) deleteK8sResourceOfCluster(ctx K8sContext, clusterName string, obj client.Object) error {
	if obj.GetObjectKind().GroupVersionKind().Empty() {
		return errors.Newf("object GVK is not set, can not apply")
	}

	m, err := fn.K8sObjToMap(obj)
	if err != nil {
		return errors.NewE(err)
	}
	b, err := json.Marshal(t.AgentMessage{
		AccountName: ctx.GetAccountName(),
		Action:      t.ActionDelete,
		Object:      m,
	})
	if err != nil {
		return errors.NewE(err)
	}

	subject := common.SendToAgentSubjectName(ctx.GetAccountName(), clusterName, obj.GetObjectKind().GroupVersionKind().String(), obj.GetNamespace(), obj.GetName())
	err = d.producer.Produce(ctx, msgTypes.ProduceMsg{
		Subject: subject,
		Payload: b,
	})

	return errors.NewE(err)
}

func (d *domain) deleteK8sResource(ctx K8sContext, environmentName string, obj client.Object) error {
	var dispatchAddr struct {
		AccountName string
		ClusterName string
	}

	clusterName, err := d.getClusterAttachedToEnvironment(ctx, environmentName)
	if err != nil {
		return ErrNoClusterAttached
	}

	if clusterName == nil || *clusterName == "" {
		d.logger.Infof("skipping delete of k8s resource %s/%s, cluster name not provided", obj.GetNamespace(), obj.GetName())
		return ErrNoClusterAttached
	}

	switch *clusterName {
	case "__kloudlite_enabled_cluster":
		{
			allocatedEdge, err := d.platformEdgeClient.GetAllocatedPlatformEdgeCluster(ctx, &platform_edge.GetAllocatedPlatformEdgeClusterIn{AccountName: ctx.GetAccountName()})
			if err != nil {
				return errors.NewEf(err, "failed to get allocated edge cluster")
			}

			dispatchAddr.AccountName = allocatedEdge.OwnedByAccount
			dispatchAddr.ClusterName = allocatedEdge.ClusterName
		}
	default:
		{

			dispatchAddr.AccountName = ctx.GetAccountName()
			dispatchAddr.ClusterName = *clusterName
		}
	}

	if obj.GetObjectKind().GroupVersionKind().Empty() {
		return errors.Newf("object GVK is not set, can not apply")
	}

	m, err := fn.K8sObjToMap(obj)
	if err != nil {
		return errors.NewE(err)
	}
	b, err := json.Marshal(t.AgentMessage{
		AccountName: dispatchAddr.AccountName,
		Action:      t.ActionDelete,
		Object:      m,
	})
	if err != nil {
		return errors.NewE(err)
	}

	subject := common.SendToAgentSubjectName(dispatchAddr.AccountName, dispatchAddr.ClusterName, obj.GetObjectKind().GroupVersionKind().String(), obj.GetNamespace(), obj.GetName())

	err = d.producer.Produce(ctx, msgTypes.ProduceMsg{
		Subject: subject,
		Payload: b,
	})

	return errors.NewE(err)
}

func (d *domain) resyncK8sResource(ctx K8sContext, environmentName string, action types.SyncAction, obj client.Object, rv int) error {
	switch action {
	case types.SyncActionApply:
		{
			return d.applyK8sResource(ctx, environmentName, obj, rv)
		}
	case types.SyncActionDelete:
		{
			return d.deleteK8sResource(ctx, environmentName, obj)
		}
	default:
		{
			return errors.Newf("unknown sync action %q", action)
		}
	}
}

func (d *domain) resyncK8sResourceToCluster(ctx K8sContext, clusterName string, action types.SyncAction, obj client.Object, rv int) error {
	switch action {
	case types.SyncActionApply:
		{
			return d.applyK8sResourceOnCluster(ctx, clusterName, obj, rv)
		}
	case types.SyncActionDelete:
		{
			return d.deleteK8sResourceOfCluster(ctx, clusterName, obj)
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

func (d *domain) MatchRecordVersion(annotations map[string]string, rv int) (int, error) {
	annVersion, err := d.parseRecordVersionFromAnnotations(annotations)
	if err != nil {
		return -1, errors.NewE(err)
	}

	if annVersion != rv {
		return -1, errors.Newf("record version mismatch, expected %d, got %d", rv, annVersion)
	}

	return annVersion, nil
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

func (d *domain) checkEnvironmentAccess(ctx ResourceContext, action iamT.Action) error {
	co, err := d.iamClient.Can(ctx, &iam.CanIn{
		UserId: string(ctx.UserId),
		ResourceRefs: []string{
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceAccount, ctx.AccountName),
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
			iamT.NewResourceRef(ctx.AccountName, iamT.ResourceConsoleVPNDevice, devName),
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

	if err := d.applyK8sResource(ctx, ctx.EnvironmentName, obj, 0); err != nil {
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
		infraSvc ports.InfraService,
		platformEdgeClient platform_edge.PlatformEdgeClient,
		accountsSvc AccountsSvc,

		environmentRepo repos.DbRepo[*entities.Environment],
		registryImageRepo repos.DbRepo[*entities.RegistryImage],

		appRepo repos.DbRepo[*entities.App],
		externalAppRepo repos.DbRepo[*entities.ExternalApp],
		configRepo repos.DbRepo[*entities.Config],
		secretRepo repos.DbRepo[*entities.Secret],
		routerRepo repos.DbRepo[*entities.Router],
		mresRepo repos.DbRepo[*entities.ManagedResource],
		importedMresRepo repos.DbRepo[*entities.ImportedManagedResource],
		ipsRepo repos.DbRepo[*entities.ImagePullSecret],
		resourceMappingRepo repos.DbRepo[*entities.ResourceMapping],
		serviceBindingRepo repos.DbRepo[*entities.ServiceBinding],
		clusterManagedServiceRepo repos.DbRepo[*entities.ClusterManagedService],

		logger logging.Logger,
		resourceEventPublisher ResourceEventPublisher,

		ev *env.Env,

		consoleCacheStore ConsoleCacheStore,
	) Domain {
		return &domain{
			k8sClient: k8sClient,

			producer: producer,

			iamClient:          iamClient,
			infraSvc:           infraSvc,
			platformEdgeClient: platformEdgeClient,
			AccountsSvc:        accountsSvc,

			logger: logger,

			environmentRepo:           environmentRepo,
			appRepo:                   appRepo,
			externalAppRepo:           externalAppRepo,
			configRepo:                configRepo,
			routerRepo:                routerRepo,
			secretRepo:                secretRepo,
			mresRepo:                  mresRepo,
			importedMresRepo:          importedMresRepo,
			pullSecretsRepo:           ipsRepo,
			resourceMappingRepo:       resourceMappingRepo,
			serviceBindingRepo:        serviceBindingRepo,
			clusterManagedServiceRepo: clusterManagedServiceRepo,
			registryImageRepo:         registryImageRepo,

			envVars: ev,

			resourceEventPublisher: resourceEventPublisher,
			consoleCacheStore:      consoleCacheStore,
		}
	}))
