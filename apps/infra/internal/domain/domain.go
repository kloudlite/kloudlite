package domain

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/k8s"
	"gopkg.in/yaml.v2"

	"github.com/kloudlite/api/apps/infra/internal/entities"

	"github.com/kloudlite/api/apps/infra/internal/env"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"
	message_office_internal "github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/message-office-internal"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/logging"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/operator/pkg/constants"
	"go.uber.org/fx"
	"sigs.k8s.io/controller-runtime/pkg/client"

	types "github.com/kloudlite/api/pkg/types"
)

type domain struct {
	logger logging.Logger
	env    *env.Env

	clusterRepo               repos.DbRepo[*entities.Cluster]
	clusterManagedServiceRepo repos.DbRepo[*entities.ClusterManagedService]
	helmReleaseRepo           repos.DbRepo[*entities.HelmRelease]
	nodeRepo                  repos.DbRepo[*entities.Node]
	nodePoolRepo              repos.DbRepo[*entities.NodePool]
	domainEntryRepo           repos.DbRepo[*entities.DomainEntry]
	secretRepo                repos.DbRepo[*entities.CloudProviderSecret]
	vpnDeviceRepo             repos.DbRepo[*entities.VPNDevice]
	pvcRepo                   repos.DbRepo[*entities.PersistentVolumeClaim]

	iamClient                   iam.IAMClient
	accountsSvc                 AccountsSvc
	messageOfficeInternalClient message_office_internal.MessageOfficeInternalClient
	resDispatcher               ResourceDispatcher
	k8sClient                   k8s.Client
	resourceEventPublisher      ResourceEventPublisher

	msvcTemplates    []*entities.MsvcTemplate
	msvcTemplatesMap map[string]map[string]*entities.MsvcTemplateEntry
}

func (d *domain) resyncToTargetCluster(ctx InfraContext, action types.SyncAction, clusterName string, obj client.Object, recordVersion int) error {
	switch action {
	case types.SyncActionApply:
		return d.resDispatcher.ApplyToTargetCluster(ctx, clusterName, obj, recordVersion)
	case types.SyncActionDelete:
		return d.resDispatcher.DeleteFromTargetCluster(ctx, clusterName, obj)
	}
	return errors.Newf("unknonw action: %q", action)
}

func (d *domain) applyK8sResource(ctx InfraContext, obj client.Object, recordVersion int) error {
	if recordVersion > 0 {
		ann := obj.GetAnnotations()
		if ann == nil {
			ann = make(map[string]string, 1)
		}
		ann[constants.RecordVersionKey] = fmt.Sprintf("%d", recordVersion)
		obj.SetAnnotations(ann)
	}

	b, err := fn.K8sObjToYAML(obj)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.k8sClient.ApplyYAML(ctx, b); err != nil {
		return errors.NewE(err)
	}
	return nil
}

func (d *domain) deleteK8sResource(ctx InfraContext, obj client.Object) error {
	b, err := fn.K8sObjToYAML(obj)
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.k8sClient.DeleteYAML(ctx, b); err != nil {
		return errors.NewE(err)
	}
	return nil
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

func (d *domain) matchRecordVersion(annotations map[string]string, rv int) error {
	annVersion, err := d.parseRecordVersionFromAnnotations(annotations)
	if err != nil {
		return errors.NewE(err)
	}

	if annVersion != rv {
		return errors.Newf("record version mismatch, expected %d, got %d", rv, annVersion)
	}

	return nil
}

func (d *domain) getAccNamespace(ctx InfraContext, name string) (string, error) {
	acc, err := d.accountsSvc.GetAccount(ctx, string(ctx.UserId), ctx.AccountName)
	if err != nil {
		return "", errors.NewE(err)
	}
	if !acc.IsActive {
		return "", errors.Newf("account %q is not active", ctx.AccountName)
	}

	return acc.TargetNamespace, nil
}

var Module = fx.Module("domain",
	fx.Provide(
		func(
			env *env.Env,
			clusterRepo repos.DbRepo[*entities.Cluster],
			clustermanagedserviceRepo repos.DbRepo[*entities.ClusterManagedService],
			nodeRepo repos.DbRepo[*entities.Node],
			nodePoolRepo repos.DbRepo[*entities.NodePool],
			secretRepo repos.DbRepo[*entities.CloudProviderSecret],
			domainNameRepo repos.DbRepo[*entities.DomainEntry],
			vpnDeviceRepo repos.DbRepo[*entities.VPNDevice],
			pvcRepo repos.DbRepo[*entities.PersistentVolumeClaim],
			resourceDispatcher ResourceDispatcher,
			helmReleaseRepo repos.DbRepo[*entities.HelmRelease],

			k8sClient k8s.Client,

			iamClient iam.IAMClient,
			accountsSvc AccountsSvc,
			msgOfficeInternalClient message_office_internal.MessageOfficeInternalClient,
			logger logging.Logger,
			resourceEventPublisher ResourceEventPublisher,

		) (Domain, error) {

			open, err := os.Open(env.MsvcTemplateFilePath)
			if err != nil {
				return nil, errors.NewE(err)
			}

			b, err := io.ReadAll(open)
			if err != nil {
				return nil, errors.NewE(err)
			}

			var templates []*entities.MsvcTemplate

			if err := yaml.Unmarshal(b, &templates); err != nil {
				return nil, errors.NewE(err)
			}

			msvcTemplatesMap := map[string]map[string]*entities.MsvcTemplateEntry{}

			for _, t := range templates {
				if _, ok := msvcTemplatesMap[t.Category]; !ok {
					msvcTemplatesMap[t.Category] = make(map[string]*entities.MsvcTemplateEntry, len(t.Items))
				}
				for i := range t.Items {
					msvcTemplatesMap[t.Category][t.Items[i].Name] = &t.Items[i]
				}
			}



			return &domain{
				msvcTemplatesMap: msvcTemplatesMap,
				msvcTemplates:    templates,
				logger:                      logger,
				env:                         env,
				clusterRepo:                 clusterRepo,
				clusterManagedServiceRepo:   clustermanagedserviceRepo,
				nodeRepo:                    nodeRepo,
				nodePoolRepo:                nodePoolRepo,
				secretRepo:                  secretRepo,
				domainEntryRepo:             domainNameRepo,
				vpnDeviceRepo:               vpnDeviceRepo,
				pvcRepo:                     pvcRepo,
				resDispatcher:               resourceDispatcher,
				k8sClient:                   k8sClient,
				iamClient:                   iamClient,
				accountsSvc:                 accountsSvc,
				messageOfficeInternalClient: msgOfficeInternalClient,
				resourceEventPublisher:      resourceEventPublisher,
				helmReleaseRepo:             helmReleaseRepo,
			}, nil
		}),
)
