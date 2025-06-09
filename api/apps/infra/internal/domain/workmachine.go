package domain

import (
	"github.com/kloudlite/api/apps/infra/internal/entities"
	fc "github.com/kloudlite/api/apps/infra/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/grpc-interfaces/kloudlite.io/rpc/auth"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
	klv1 "github.com/kloudlite/operator/apis/crds/v1"
	"github.com/kloudlite/operator/operators/resource-watcher/types"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (d *domain) applyWorkmachine(ctx InfraContext, wm *entities.Workmachine) error {
	addTrackingId(&wm.WorkMachine, wm.Id)
	err := d.resDispatcher.ApplyToTargetCluster(ctx, wm.DispatchAddr, &v1.Namespace{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Namespace",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: wm.Spec.TargetNamespace,
		},
	}, wm.RecordVersion)
	if err != nil {
		return errors.NewE(err)
	}
	err = d.resDispatcher.ApplyToTargetCluster(ctx, wm.DispatchAddr, &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kl-session",
			Namespace: wm.Spec.TargetNamespace,
		},
		Data: map[string][]byte{
			"session-id": []byte(wm.SessionId),
		},
	}, wm.RecordVersion)
	if err != nil {
		return errors.NewE(err)
	}
	return d.resDispatcher.ApplyToTargetCluster(ctx, wm.DispatchAddr, &wm.WorkMachine, wm.RecordVersion)
}

func (d *domain) findWorkmachine(ctx InfraContext, clusterName string, name string) (*entities.Workmachine, error) {
	wm, err := d.workmachineRepo.FindOne(ctx, repos.Filter{
		fc.AccountName:  ctx.AccountName,
		fc.MetadataName: name,
		fc.ClusterName:  clusterName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	// if wm == nil {
	// 	return nil, errors.Newf("no workmachine for account=%q found", ctx.AccountName)
	// }
	return wm, nil
}

func (d *domain) CreateWorkMachine(ctx InfraContext, clusterName string, workmachine entities.Workmachine) (*entities.Workmachine, error) {
	workmachine.AccountName = ctx.AccountName
	workmachine.ClusterName = clusterName

	workmachine.DispatchAddr = &entities.DispatchAddr{
		AccountName: ctx.AccountName,
		ClusterName: clusterName,
	}

	workmachine.CreatedBy = common.CreatedOrUpdatedBy{
		UserId:    ctx.UserId,
		UserName:  ctx.UserName,
		UserEmail: ctx.UserEmail,
	}

	workmachine.LastUpdatedBy = workmachine.CreatedBy

	workmachine.EnsureGVK()
	if err := d.k8sClient.ValidateObject(ctx, &workmachine.WorkMachine); err != nil {
		return nil, errors.NewE(err)
	}

	out, err := d.authClient.GenerateMachineSession(ctx, &auth.GenerateMachineSessionIn{
		UserId:    string(ctx.UserId),
		MachineId: workmachine.Name,
		Cluster:   workmachine.ClusterName,
	})
	if err != nil {
		return nil, errors.NewE(err)
	}
	workmachine.SessionId = out.SessionId
	workmachine.Spec.JobParams = klv1.WorkMachineJobParams{
		NodeSelector: map[string]string{},
		Tolerations:  []v1.Toleration{},
	}
	workmachine.Spec.AWSMachineConfig.RootVolumeSize = 100
	workmachine.Spec.AWSMachineConfig.RootVolumeType = "gp2"

	wm, err := d.workmachineRepo.Create(ctx, &workmachine)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeWorkmachine, wm.Name, PublishAdd)

	if err := d.applyWorkmachine(ctx, wm); err != nil {
		return nil, errors.NewE(err)
	}

	return wm, nil
}

func (d *domain) UpdateWorkMachine(ctx InfraContext, clusterName string, workmachine entities.Workmachine) (*entities.Workmachine, error) {
	patchForUpdate := repos.Document{
		fc.DisplayName:                          workmachine.DisplayName,
		fc.WorkmachineSpecAwsAmi:                workmachine.Spec.AWSMachineConfig.AMI,
		fc.WorkmachineSpecAwsExternalVolumeSize: workmachine.Spec.AWSMachineConfig.ExternalVolumeSize,
		fc.WorkmachineSpecAwsInstanceType:       workmachine.Spec.AWSMachineConfig.InstanceType,
		fc.WorkmachineSpecSshPublicKeys:         workmachine.Spec.SSHPublicKeys,
		fc.WorkmachineSpecState:                 workmachine.Spec.State,
		fc.LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	}

	upWorkmachine, err := d.workmachineRepo.Patch(
		ctx,
		repos.Filter{
			fc.AccountName:     ctx.AccountName,
			fc.MetadataName:    workmachine.Name,
			fields.ClusterName: clusterName,
		},
		patchForUpdate,
	)
	if err != nil {
		return nil, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, workmachine.ClusterName, ResourceTypeWorkmachine, upWorkmachine.Name, PublishUpdate)

	if err := d.applyWorkmachine(ctx, upWorkmachine); err != nil {
		return nil, errors.NewE(err)
	}

	return upWorkmachine, nil
}

func (d *domain) UpdateWorkmachineStatus(ctx InfraContext, clusterName string, status bool, name string) (bool, error) {
	machineStatus := "OFF"
	if status {
		machineStatus = "ON"
	}

	patchForUpdate := repos.Document{
		fc.WorkmachineSpecState: machineStatus,
		// fc.WorkmachineMachineStatus: status,
		fc.LastUpdatedBy: common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		},
	}

	upWorkmachine, err := d.workmachineRepo.Patch(
		ctx,
		repos.Filter{
			fc.AccountName:     ctx.AccountName,
			fc.MetadataName:    name,
			fields.ClusterName: clusterName,
		},
		patchForUpdate,
	)
	if err != nil {
		return false, errors.NewE(err)
	}

	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeWorkmachine, upWorkmachine.Name, PublishUpdate)

	if err := d.applyWorkmachine(ctx, upWorkmachine); err != nil {
		return false, errors.NewE(err)
	}

	return true, nil
}

func (d *domain) GetWorkmachine(ctx InfraContext, clusterName string, name string) (*entities.Workmachine, error) {
	return d.findWorkmachine(ctx, clusterName, name)
}

func (d *domain) OnWorkmachineDeleteMessage(ctx InfraContext, clusterName string, workmachine entities.Workmachine) error {
	err := d.workmachineRepo.DeleteOne(
		ctx,
		repos.Filter{
			fields.AccountName:  ctx.AccountName,
			fields.ClusterName:  clusterName,
			fields.MetadataName: workmachine.Name,
		},
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeWorkmachine, workmachine.Name, PublishDelete)
	return nil
}

func (d *domain) OnWorkmachineUpdateMessage(ctx InfraContext, clusterName string, workmachine entities.Workmachine, status types.ResourceStatus, opts UpdateAndDeleteOpts) error {
	wm, err := d.findWorkmachine(ctx, clusterName, workmachine.Name)
	if err != nil {
		return errors.NewE(err)
	}

	if wm == nil {
		workmachine.AccountName = ctx.AccountName
		workmachine.ClusterName = clusterName

		workmachine.CreatedBy = common.CreatedOrUpdatedBy{
			UserId:    ctx.UserId,
			UserName:  ctx.UserName,
			UserEmail: ctx.UserEmail,
		}

		workmachine.LastUpdatedBy = workmachine.CreatedBy

		wm, err = d.workmachineRepo.Create(ctx, &workmachine)
		if err != nil {
			return errors.NewE(err)
		}
	}
	patch := common.PatchForSyncFromAgent(
		&workmachine,
		workmachine.RecordVersion,
		status,
		common.PatchOpts{
			MessageTimestamp: opts.MessageTimestamp,
		})
	patch[fc.Status] = workmachine.Status
	upWm, err := d.workmachineRepo.PatchById(
		ctx,
		wm.Id,
		patch,
	)
	if err != nil {
		return errors.NewE(err)
	}
	d.resourceEventPublisher.PublishResourceEvent(ctx, clusterName, ResourceTypeWorkmachine, upWm.Name, PublishUpdate)
	return nil
}
