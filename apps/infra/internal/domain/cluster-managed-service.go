package domain

import (
	"encoding/json"
	"fmt"

	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/apps/infra/internal/entities"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/console"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *domain) ListClusterManagedServices(ctx InfraContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ClusterManagedService], error) {
	if err := d.canPerformActionInAccount(ctx, iamT.ListClusterManagedServices); err != nil {
		return nil, errors.NewE(err)
	}

	f := repos.Filter{
		fields.AccountName: ctx.AccountName,
	}

	pr, err := d.clusterManagedServiceRepo.FindPaginated(ctx, d.secretRepo.MergeMatchFilters(f, search), pagination)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return pr, nil
}

func (d *domain) findClusterManagedService(ctx InfraContext, name string) (*entities.ClusterManagedService, error) {
	cmsvc, err := d.clusterManagedServiceRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if cmsvc == nil {
		return nil, errors.Newf("cmsvc with name %q not found", name)
	}
	return cmsvc, nil
}

func (d *domain) GetClusterManagedService(ctx InfraContext, serviceName string) (*entities.ClusterManagedService, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	c, err := d.findClusterManagedService(ctx, serviceName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return c, nil
}

func (d *domain) applyClusterManagedService(ctx InfraContext, cmsvc *entities.ClusterManagedService) error {
	addTrackingId(&cmsvc.ClusterManagedService, cmsvc.Id)
	return d.resDispatcher.ApplyToTargetCluster(ctx, cmsvc.ClusterName, &cmsvc.ClusterManagedService, cmsvc.RecordVersion)
}

func (d *domain) CreateClusterManagedService(ctx InfraContext, cmsvc entities.ClusterManagedService) (*entities.ClusterManagedService, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CreateClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	cmsvc.IncrementRecordVersion()

	cmsvc.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	cmsvc.LastUpdatedBy = cmsvc.CreatedBy

	existing, err := d.clusterManagedServiceRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:  ctx.AccountName,
		fields.MetadataName: cmsvc.Name,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}

	if existing != nil {
		return nil, errors.Newf("cluster managed service with name %q already exists", cmsvc.ClusterName)
	}

	cmsvc.AccountName = ctx.AccountName
	cmsvc.SyncStatus = t.GenSyncStatus(t.SyncActionApply, cmsvc.RecordVersion)

	// cmsvc.Spec.SharedSecret = fn.New(fn.CleanerNanoid(40))

	cmsvc.EnsureGVK()

	if err := d.k8sClient.ValidateObject(ctx, &cmsvc.ClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	ncms, err := d.clusterManagedServiceRepo.Create(ctx, &cmsvc)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyClusterManagedService(ctx, &cmsvc); err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, cmsvc.ClusterName, ResourceTypeClusterManagedService, ncms.Name, PublishAdd)

	return ncms, nil
}

type CloneManagedServiceArgs struct {
	SourceMsvcName      string
	DestinationMsvcName string
	DisplayName         string
	ClusterName         string
}

func (d *domain) getClusterManagedServiceTargetNamespace(msvcName string) string {
	return fmt.Sprintf("cmsvc-%s", msvcName)
}

func (d *domain) CloneClusterManagedService(ctx InfraContext, args CloneManagedServiceArgs) (*entities.ClusterManagedService, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.CloneClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	sourceMsvc, err := d.findClusterManagedService(ctx, args.SourceMsvcName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	destMsvc := &entities.ClusterManagedService{
		ClusterManagedService: crdsv1.ClusterManagedService{
			TypeMeta: sourceMsvc.TypeMeta,
			ObjectMeta: metav1.ObjectMeta{
				Name:      args.DestinationMsvcName,
				Namespace: sourceMsvc.Namespace,
			},
			Spec: crdsv1.ClusterManagedServiceSpec{
				TargetNamespace: d.getClusterManagedServiceTargetNamespace(args.DestinationMsvcName),
				MSVCSpec:        sourceMsvc.Spec.MSVCSpec,
			},
		},
		AccountName:           ctx.AccountName,
		ClusterName:           args.ClusterName,
		SyncedOutputSecretRef: sourceMsvc.SyncedOutputSecretRef,
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

	if err := d.k8sClient.ValidateObject(ctx, &destMsvc.ClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	destMsvc, err = d.clusterManagedServiceRepo.Create(ctx, destMsvc)
	if err != nil {
		return nil, errors.NewE(err)
	}

	if err := d.applyClusterManagedService(ctx, destMsvc); err != nil {
		return nil, errors.NewE(err)
	}

	return destMsvc, nil
}

func (d *domain) ArchiveClusterManagedService(ctx InfraContext, clusterName string) error {
	filter := repos.Filter{
		fields.AccountName: ctx.AccountName,
		fields.ClusterName: clusterName,
	}

	msvc, err := d.clusterManagedServiceRepo.Find(ctx, repos.Query{
		Filter: filter,
		Sort:   nil,
	})
	if err != nil {
		return errors.NewE(err)
	}

	for i := range msvc {
		patchForUpdate := repos.Document{
			fc.ClusterManagedServiceIsArchived: true,
		}
		patchFilter := repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.ClusterName:  clusterName,
			fields.MetadataName: msvc[i].Name,
		}

		_, err := d.clusterManagedServiceRepo.Patch(ctx, patchFilter, patchForUpdate)
		if err != nil {
			return errors.NewE(err)
		}
	}
	return nil
}

func (d *domain) UpdateClusterManagedService(ctx InfraContext, cmsvc entities.ClusterManagedService) (*entities.ClusterManagedService, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	cmsvc.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &cmsvc.ClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	patchForUpdate := common.PatchForUpdate(
		ctx,
		&cmsvc,
		common.PatchOpts{
			XPatch: repos.Document{
				fc.ClusterManagedServiceSpecMsvcSpec: cmsvc.Spec.MSVCSpec,
			},
		})

	ucmsvc, err := d.clusterManagedServiceRepo.Patch(ctx, repos.Filter{fields.AccountName: ctx.AccountName, fields.MetadataName: cmsvc.Name}, patchForUpdate)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, ucmsvc.ClusterName, ResourceTypeClusterManagedService, ucmsvc.Name, PublishUpdate)

	if err := d.applyClusterManagedService(ctx, ucmsvc); err != nil {
		return nil, errors.NewE(err)
	}

	return ucmsvc, nil
}

func (d *domain) DeleteClusterManagedService(ctx InfraContext, name string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteClusterManagedService); err != nil {
		return errors.NewE(err)
	}

	ucmsvc, err := d.clusterManagedServiceRepo.Patch(ctx, repos.Filter{fields.AccountName: ctx.AccountName, fields.MetadataName: name}, common.PatchForMarkDeletion())
	if err != nil {
		return errors.NewE(err)
	}

	if ucmsvc.IsArchived != nil && *ucmsvc.IsArchived {
		return d.clusterManagedServiceRepo.DeleteById(ctx, ucmsvc.Id)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, ucmsvc.ClusterName, ResourceTypeClusterManagedService, ucmsvc.Name, PublishUpdate)

	return d.resDispatcher.DeleteFromTargetCluster(ctx, ucmsvc.ClusterName, &ucmsvc.ClusterManagedService)
}

func (d *domain) OnClusterManagedServiceApplyError(ctx InfraContext, clusterName, name, errMsg string, opts UpdateAndDeleteOpts) error {
	ucmsvc, err := d.clusterManagedServiceRepo.Patch(
		ctx,
		repos.Filter{
			fields.ClusterName:  clusterName,
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

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeClusterManagedService, ucmsvc.Name, PublishDelete)
	return errors.NewE(err)
}

func (d *domain) OnClusterManagedServiceDeleteMessage(ctx InfraContext, clusterName string, service entities.ClusterManagedService) error {
	err := d.clusterManagedServiceRepo.DeleteOne(
		ctx,
		repos.Filter{
			fields.ClusterName:  clusterName,
			fields.AccountName:  ctx.AccountName,
			fields.MetadataName: service.Name,
		},
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeClusterManagedService, service.Name, PublishDelete)
	return err
}

func (d *domain) OnClusterManagedServiceUpdateMessage(ctx InfraContext, clusterName string, service entities.ClusterManagedService, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xService, err := d.findClusterManagedService(ctx, service.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xService == nil {
		return errors.Newf("no cluster manage service found")
	}

	if _, err := d.matchRecordVersion(service.Annotations, xService.RecordVersion); err != nil {
		return d.resyncToTargetCluster(ctx, xService.SyncStatus.Action, clusterName, xService, xService.RecordVersion)
	}

	recordVersion, err := d.matchRecordVersion(service.Annotations, xService.RecordVersion)
	if err != nil {
		return errors.NewE(err)
	}

	patch := repos.Document{
		fc.ClusterManagedServiceSpecTargetNamespace:   service.Spec.TargetNamespace,
		fc.ClusterManagedServiceSyncedOutputSecretRef: service.SyncedOutputSecretRef,
	}

	if service.SyncedOutputSecretRef != nil {
		b, err := json.Marshal(service.SyncedOutputSecretRef)
		if err != nil {
			return errors.NewE(err)
		}
		accNs, err := d.getAccNamespace(ctx)
		if err != nil {
			return errors.NewE(err)
		}

		d.consoleClient.CreateManagedResource(ctx, &console.CreateManagedResourceIn{
			UserId:              string(ctx.UserId),
			UserName:            string(ctx.UserName),
			UserEmail:           string(ctx.UserEmail),
			AccountName:         ctx.AccountName,
			ClusterName:         xService.ClusterName,
			MsvcName:            xService.Name,
			AccountNamespace:    accNs,
			MsvcTargetNamespace: xService.Spec.TargetNamespace,
			MresName:            "root-credentials",
			MresType:            "root-credentials",
			OutputSecret:        b,
      MsvcApiVersion:      xService.Spec.MSVCSpec.ServiceTemplate.APIVersion,
		})
	}

	ucmsvc, err := d.clusterManagedServiceRepo.PatchById(ctx, xService.Id, common.PatchForSyncFromAgent(&service, recordVersion, status, common.PatchOpts{
		MessageTimestamp: opts.MessageTimestamp,
		XPatch:           patch,
	}))
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeClusterManagedService, ucmsvc.GetName(), PublishUpdate)
	return nil
}
