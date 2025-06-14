package domain

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"

	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/console"

	"sigs.k8s.io/yaml"

	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/k8s"

	"github.com/kloudlite/api/apps/infra/internal/domain/ports"
	"github.com/kloudlite/api/apps/infra/internal/entities"

	"github.com/kloudlite/api/apps/infra/internal/env"
	constant "github.com/kloudlite/api/constants"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/iam"

	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/helm"
	"github.com/kloudlite/api/pkg/repos"
	"github.com/kloudlite/operator/pkg/constants"
	"go.uber.org/fx"
	"sigs.k8s.io/controller-runtime/pkg/client"

	types "github.com/kloudlite/api/pkg/types"
)

type domain struct {
	logger *slog.Logger
	env    *env.Env

	clusterRepo     repos.DbRepo[*entities.Cluster]
	byokClusterRepo repos.DbRepo[*entities.BYOKCluster]
	nodeRepo        repos.DbRepo[*entities.Node]
	nodePoolRepo    repos.DbRepo[*entities.NodePool]

	gvpnConnRepo            repos.DbRepo[*entities.GlobalVPNConnection]
	freeClusterSvcCIDRRepo  repos.DbRepo[*entities.FreeClusterSvcCIDR]
	claimClusterSvcCIDRRepo repos.DbRepo[*entities.ClaimClusterSvcCIDR]

	gvpnRepo        repos.DbRepo[*entities.GlobalVPN]
	gvpnDevicesRepo repos.DbRepo[*entities.GlobalVPNDevice]

	freeDeviceIpRepo  repos.DbRepo[*entities.FreeDeviceIP]
	claimDeviceIPRepo repos.DbRepo[*entities.ClaimDeviceIP]

	domainEntryRepo      repos.DbRepo[*entities.DomainEntry]
	secretRepo           repos.DbRepo[*entities.CloudProviderSecret]
	pvcRepo              repos.DbRepo[*entities.PersistentVolumeClaim]
	namespaceRepo        repos.DbRepo[*entities.Namespace]
	pvRepo               repos.DbRepo[*entities.PersistentVolume]
	volumeAttachmentRepo repos.DbRepo[*entities.VolumeAttachment]
	workspaceRepo        repos.DbRepo[*entities.Workspace]
	workmachineRepo      repos.DbRepo[*entities.Workmachine]

	iamClient              iam.IAMClient
	consoleClient          console.ConsoleClient
	authClient             auth.AuthClient
	accountsSvc            AccountsSvc
	moSvc                  ports.MessageOfficeService
	resDispatcher          ResourceDispatcher
	k8sClient              k8s.Client
	resourceEventPublisher ResourceEventPublisher

	msvcTemplates    []*entities.MsvcTemplate
	msvcTemplatesMap map[string]map[string]*entities.MsvcTemplateEntry

	helmClient helm.Client
}

func (d *domain) resyncToTargetCluster(ctx InfraContext, action types.SyncAction, dispatchAddr *entities.DispatchAddr, obj client.Object, recordVersion int) error {
	switch action {
	case types.SyncActionApply:
		return d.resDispatcher.ApplyToTargetCluster(ctx, dispatchAddr, obj, recordVersion)
	case types.SyncActionDelete:
		return d.resDispatcher.DeleteFromTargetCluster(ctx, dispatchAddr, obj)
	}
	return errors.Newf("unknonw action: %q", action)
}

func addTrackingId(obj client.Object, id repos.ID) {
	ann := obj.GetAnnotations()
	if ann == nil {
		ann = make(map[string]string, 1)
	}
	ann[constant.ObservabilityTrackingKey] = string(id)

	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string, 2)
	}
	labels[constant.ObservabilityTrackingKey] = string(id)
	obj.SetLabels(labels)
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

func (d *domain) matchRecordVersion(annotations map[string]string, rv int) (int, error) {
	annVersion, err := d.parseRecordVersionFromAnnotations(annotations)
	if err != nil {
		return -1, errors.NewE(err)
	}

	if annVersion != rv {
		return -1, errors.Newf("record version mismatch, expected %d, got %d", rv, annVersion)
	}

	return annVersion, nil
}

func (d *domain) getAccNamespace(ctx InfraContext) (string, error) {
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
			byokClusterRepo repos.DbRepo[*entities.BYOKCluster],
			nodeRepo repos.DbRepo[*entities.Node],
			nodePoolRepo repos.DbRepo[*entities.NodePool],
			secretRepo repos.DbRepo[*entities.CloudProviderSecret],
			domainNameRepo repos.DbRepo[*entities.DomainEntry],
			resourceDispatcher ResourceDispatcher,
			// helmReleaseRepo repos.DbRepo[*entities.HelmRelease],

			gvpnConnRepo repos.DbRepo[*entities.GlobalVPNConnection],
			gvpnRepo repos.DbRepo[*entities.GlobalVPN],
			gvpnDevicesRepo repos.DbRepo[*entities.GlobalVPNDevice],

			freeDeviceIpRepo repos.DbRepo[*entities.FreeDeviceIP],
			claimDeviceIPRepo repos.DbRepo[*entities.ClaimDeviceIP],

			freeClusterSvcCIDRRepo repos.DbRepo[*entities.FreeClusterSvcCIDR],
			claimClusterSvcCIDRRepo repos.DbRepo[*entities.ClaimClusterSvcCIDR],

			pvcRepo repos.DbRepo[*entities.PersistentVolumeClaim],
			pvRepo repos.DbRepo[*entities.PersistentVolume],
			namespaceRepo repos.DbRepo[*entities.Namespace],
			volumeAttachmentRepo repos.DbRepo[*entities.VolumeAttachment],
			workspaceRepo repos.DbRepo[*entities.Workspace],
			workmachineRepo repos.DbRepo[*entities.Workmachine],

			k8sClient k8s.Client,

			iamClient iam.IAMClient,
			authClient auth.AuthClient,
			consoleClient console.ConsoleClient,
			accountsSvc AccountsSvc,
			moSvc ports.MessageOfficeService,
			logger *slog.Logger,
			resourceEventPublisher ResourceEventPublisher,

			helmClient helm.Client,
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
				logger:           logger,
				env:              env,
				clusterRepo:      clusterRepo,
				gvpnConnRepo:     gvpnConnRepo,
				// deviceAddressPoolRepo:   deviceAddressPoolRepo,

				claimDeviceIPRepo:       claimDeviceIPRepo,
				freeDeviceIpRepo:        freeDeviceIpRepo,
				freeClusterSvcCIDRRepo:  freeClusterSvcCIDRRepo,
				claimClusterSvcCIDRRepo: claimClusterSvcCIDRRepo,

				gvpnRepo:        gvpnRepo,
				gvpnDevicesRepo: gvpnDevicesRepo,

				byokClusterRepo:        byokClusterRepo,
				nodeRepo:               nodeRepo,
				nodePoolRepo:           nodePoolRepo,
				secretRepo:             secretRepo,
				domainEntryRepo:        domainNameRepo,
				resDispatcher:          resourceDispatcher,
				k8sClient:              k8sClient,
				iamClient:              iamClient,
				consoleClient:          consoleClient,
				authClient:             authClient,
				accountsSvc:            accountsSvc,
				moSvc:                  moSvc,
				resourceEventPublisher: resourceEventPublisher,

				pvcRepo:              pvcRepo,
				volumeAttachmentRepo: volumeAttachmentRepo,
				pvRepo:               pvRepo,
				namespaceRepo:        namespaceRepo,
				workspaceRepo:        workspaceRepo,
				workmachineRepo:      workmachineRepo,

				helmClient: helmClient,
			}, nil
		}),
)
