package domain

import (
	"context"

	infrav1 "github.com/kloudlite/internal_operator_v2/apis/infra/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateCloudProvider(ctx context.Context, cloudProvider *entities.CloudProvider) (*entities.CloudProvider, error) {
	var cp *entities.CloudProvider
	var err error
	cloudProvider.TypeMeta = v1.TypeMeta{
		Kind:       "CloudProvider",
		APIVersion: infrav1.GroupVersion.String(),
	}

	if cp, err = d.providerRepo.Create(ctx, cloudProvider); err != nil {
		return nil, err
	}

	clusterId, err := d.getClusterForAccount(ctx, repos.ID(cp.Spec.AccountId))
	if err != nil {
		return nil, err
	}

	if err = d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId), string(cp.Id), &infrav1.CloudProvider{
			TypeMeta:   cp.TypeMeta,
			ObjectMeta: cp.ObjectMeta,
			Spec:       cp.Spec,
		},
	); err != nil {
		return nil, err
	}

	return cp, err
}

func (d *domain) DeleteCloudProvider(ctx context.Context, name string) error {
	var cp *entities.CloudProvider
	var err error

	if cp, err = d.providerRepo.FindOne(ctx, repos.Filter{
		"metadata.name": name,
	}); err != nil {
		return err
	}

	if err = d.providerRepo.DeleteOne(ctx, repos.Filter{
		"metadata.name": name,
	}); err != nil {
		return err
	}

	clusterId, err := d.getClusterForAccount(ctx, repos.ID(cp.Spec.AccountId))
	if err != nil {
		return err
	}

	if err = d.workloadMessenger.SendAction("delete", d.getDispatchKafkaTopic(clusterId), string(cp.Id), cp.CloudProvider); err != nil {
		return err
	}

	return nil
}

func (d *domain) GetCloudProvider(ctx context.Context, name string) (*entities.CloudProvider, error) {
	return d.providerRepo.FindOne(ctx, repos.Filter{
		"metadata.name": name,
	})
}

func (d *domain) ListCloudProviders(ctx context.Context, accountId repos.ID) ([]*entities.CloudProvider, error) {
	return d.providerRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"spec.accountId": accountId,
		},
	})
}

func (d *domain) UpdateCloudProvider(ctx context.Context, cloudProvider *entities.CloudProvider) error {
	var cp *entities.CloudProvider
	var err error

	if cp, err = d.providerRepo.FindOne(ctx, repos.Filter{
		"metadata.name": cloudProvider.Name,
	}); err != nil {
		return err
	}

	clusterId, err := d.getClusterForAccount(ctx, repos.ID(cloudProvider.Spec.AccountId))
	if err != nil {
		return err
	}

	if _, err := d.providerRepo.UpdateOne(ctx, repos.Filter{
		"metadata.name": cloudProvider.Name,
	}, cloudProvider); err != nil {
		return err
	}

	if err := d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId),
		string(cp.Id),
		cp.CloudProvider,
	); err != nil {
		return err
	}

	return nil
}
