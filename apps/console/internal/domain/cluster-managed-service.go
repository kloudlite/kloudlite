package domain

import (
	"fmt"

	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	iamT "github.com/kloudlite/api/apps/iam/types"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	t "github.com/kloudlite/api/pkg/types"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	"github.com/kloudlite/operator/toolkit/plugin"
	"github.com/kloudlite/operator/toolkit/reconciler"
	toolkit_types "github.com/kloudlite/operator/toolkit/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *domain) ListClusterManagedServices(ctx ConsoleContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.ClusterManagedService], error) {
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

func (d *domain) findClusterManagedService(ctx ConsoleContext, name string) (*entities.ClusterManagedService, error) {
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

func (d *domain) GetClusterManagedService(ctx ConsoleContext, serviceName string) (*entities.ClusterManagedService, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.GetClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	c, err := d.findClusterManagedService(ctx, serviceName)
	if err != nil {
		return nil, errors.NewE(err)
	}

	return c, nil
}

func (d *domain) applyClusterManagedService(ctx ConsoleContext, cmsvc *entities.ClusterManagedService) error {
	addTrackingId(&cmsvc.ClusterManagedService, cmsvc.Id)

	// return d.applyK8sResource(ctx, envName string, obj client.Object, recordVersion int)
	return d.applyK8sResourceOnCluster(ctx, cmsvc.ClusterName, &cmsvc.ClusterManagedService, cmsvc.RecordVersion)
}

func (d *domain) CreateClusterManagedService(ctx ConsoleContext, cmsvc entities.ClusterManagedService) (*entities.ClusterManagedService, error) {
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

	d.resourceEventPublisher.PublishClusterManagedServiceEvent(ctx, ncms.Name, entities.ResourceTypeClusterManagedService, ncms.Name, PublishAdd)

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

func (d *domain) CloneClusterManagedService(ctx ConsoleContext, args CloneManagedServiceArgs) (*entities.ClusterManagedService, error) {
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

func (d *domain) ArchiveClusterManagedServicesForCluster(ctx ConsoleContext, clusterName string) (bool, error) {
	filter := repos.Filter{
		fields.AccountName: ctx.AccountName,
		fields.ClusterName: clusterName,
	}

	msvc, err := d.clusterManagedServiceRepo.Find(ctx, repos.Query{
		Filter: filter,
		Sort:   nil,
	})
	if err != nil {
		return false, errors.NewE(err)
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
			return false, errors.NewE(err)
		}
	}
	return true, nil
}

func (d *domain) UpdateClusterManagedService(ctx ConsoleContext, cmsvc entities.ClusterManagedService) (*entities.ClusterManagedService, error) {
	if err := d.canPerformActionInAccount(ctx, iamT.UpdateClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	cmsvc.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &cmsvc.ClusterManagedService); err != nil {
		return nil, errors.NewE(err)
	}

	patchForUpdate := common.PatchForUpdate(ctx, &cmsvc, common.PatchOpts{
		XPatch: repos.Document{
			fc.ClusterManagedServiceSpecMsvcSpec: cmsvc.Spec.MSVCSpec,
		},
	})

	ucmsvc, err := d.clusterManagedServiceRepo.Patch(ctx, repos.Filter{fields.AccountName: ctx.AccountName, fields.MetadataName: cmsvc.Name}, patchForUpdate)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishClusterManagedServiceEvent(ctx, ucmsvc.Name, entities.ResourceTypeClusterManagedService, ucmsvc.Name, PublishUpdate)

	if err := d.applyClusterManagedService(ctx, ucmsvc); err != nil {
		return nil, errors.NewE(err)
	}

	return ucmsvc, nil
}

func (d *domain) DeleteClusterManagedService(ctx ConsoleContext, name string) error {
	if err := d.canPerformActionInAccount(ctx, iamT.DeleteClusterManagedService); err != nil {
		return errors.NewE(err)
	}

	ucmsvc, err := d.clusterManagedServiceRepo.Patch(ctx, repos.Filter{fields.AccountName: ctx.AccountName, fields.MetadataName: name}, common.PatchForMarkDeletion())
	if err != nil {
		return errors.NewE(err)
	}

	if err := d.cleanupClusterManagedServiceResources(ctx, ucmsvc); err != nil {
		return err
	}

	isArchived := ucmsvc.IsArchived != nil && *ucmsvc.IsArchived

	if isArchived {
		return d.clusterManagedServiceRepo.DeleteById(ctx, ucmsvc.Id)
	}

	d.resourceEventPublisher.PublishClusterManagedServiceEvent(ctx, ucmsvc.Name, entities.ResourceTypeClusterManagedService, ucmsvc.Name, PublishUpdate)

	return d.deleteK8sResourceOfCluster(ctx, ucmsvc.ClusterName, &ucmsvc.ClusterManagedService)
}

func (d *domain) OnClusterManagedServiceApplyError(ctx ConsoleContext, clusterName, name, errMsg string, opts UpdateAndDeleteOpts) error {
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

	d.resourceEventPublisher.PublishClusterManagedServiceEvent(ctx, ucmsvc.Name, entities.ResourceTypeClusterManagedService, ucmsvc.Name, PublishUpdate)
	return errors.NewE(err)
}

func (d *domain) cleanupClusterManagedServiceResources(ctx ConsoleContext, msvc *entities.ClusterManagedService) error {
	if err := d.deleteAllManagedResources(ctx, msvc.Name); err != nil {
		return errors.NewE(err)
	}

	if err := d.deleteImportedManagedResources(ctx, msvc.Spec.TargetNamespace); err != nil {
		return errors.NewE(err)
	}

	return nil
}

func (d *domain) OnClusterManagedServiceDeleteMessage(ctx ConsoleContext, clusterName string, service entities.ClusterManagedService) error {
	xService, err := d.findClusterManagedService(ctx, service.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xService == nil {
		return errors.Newf("no cluster manage service found")
	}

	if _, err := d.MatchRecordVersion(service.Annotations, xService.RecordVersion); err != nil {
		return nil
	}

	if err := d.cleanupClusterManagedServiceResources(ctx, xService); err != nil {
		return err
	}

	if err := d.clusterManagedServiceRepo.DeleteById(ctx, xService.Id); err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishClusterManagedServiceEvent(ctx, service.Name, entities.ResourceTypeClusterManagedService, service.Name, PublishUpdate)
	return err
}

func (d *domain) OnClusterManagedServiceUpdateMessage(ctx ConsoleContext, clusterName string, service entities.ClusterManagedService, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	xService, err := d.findClusterManagedService(ctx, service.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if xService == nil {
		return errors.Newf("no cluster manage service found")
	}

	if _, err := d.MatchRecordVersion(service.Annotations, xService.RecordVersion); err != nil {
		return nil
	}

	patch := repos.Document{
		fc.ClusterManagedServiceSpecTargetNamespace:   service.Spec.TargetNamespace,
		fc.ClusterManagedServiceSyncedOutputSecretRef: service.SyncedOutputSecretRef,
	}

	if service.SyncedOutputSecretRef != nil && xService.Spec.MSVCSpec.Plugin != nil {
		service.SyncedOutputSecretRef.Namespace = xService.Spec.TargetNamespace

		if _, err := d.createRootManagedResource(ctx, &entities.ManagedResource{
			ManagedResource: crdsv1.ManagedResource{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "root-credentials",
					Namespace: xService.Spec.TargetNamespace,
				},
				Spec: crdsv1.ManagedResourceSpec{
					ManagedServiceRef: toolkit_types.ObjectReference{
						APIVersion: xService.Spec.MSVCSpec.Plugin.APIVersion,
						Kind:       "RootCredentials",
						Namespace:  xService.Spec.TargetNamespace,
						Name:       xService.Name,
					},
					Plugin: crdsv1.PluginTemplate{
						APIVersion: xService.Spec.MSVCSpec.Plugin.APIVersion,
						Kind:       "RootCredentials",
						Spec:       nil,
						Export: plugin.Export{
							ViaSecret: xService.SyncedOutputSecretRef.Name,
						},
					},
					// ResourceTemplate: crdsv1.MresResourceTemplate{
					// 	TypeMeta: metav1.TypeMeta{
					// 		Kind:       "RootCredentials",
					// 		APIVersion: xService.Spec.MSVCSpec.ServiceTemplate.APIVersion,
					// 	},
					// 	MsvcRef: common_types.MsvcRef{
					// 		Name:      xService.Name,
					// 		Namespace: xService.Spec.TargetNamespace,
					// 	},
				},
				Status: reconciler.Status{
					IsReady: true,
				},
			},
			// Output: common_types.ManagedResourceOutput{
			// 	CredentialsRef: common_types.LocalObjectReference{
			// 		Name: service.SyncedOutputSecretRef.Name,
			// 	},
			// },
			// },
			ResourceMetadata: common.ResourceMetadata{
				DisplayName:   fmt.Sprintf("%s/%s", xService.Name, "root-credentials"),
				CreatedBy:     xService.CreatedBy,
				LastUpdatedBy: xService.LastUpdatedBy,
			},
			AccountName:           ctx.AccountName,
			ManagedServiceName:    xService.Name,
			ClusterName:           xService.ClusterName,
			SyncedOutputSecretRef: service.SyncedOutputSecretRef,
		}); err != nil {
			return errors.NewE(err)
		}
	}

	ucmsvc, err := d.clusterManagedServiceRepo.PatchById(ctx, xService.Id, common.PatchForSyncFromAgent(&service, xService.RecordVersion, status, common.PatchOpts{
		MessageTimestamp: opts.MessageTimestamp,
		XPatch:           patch,
	}))
	if err != nil {
		return errors.NewE(err)
	}

	d.resourceEventPublisher.PublishClusterManagedServiceEvent(ctx, ucmsvc.Name, entities.ResourceTypeClusterManagedService, ucmsvc.Name, PublishAdd)

	return nil
}
