package domain

import (
	"context"

	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateCloudProvider(ctx context.Context, cloudProvider *entities.CloudProvider) (*entities.CloudProvider, error) {
	return d.providerRepo.Create(ctx, cloudProvider)
}

func (d *domain) DeleteCloudProvider(ctx context.Context, name string) error {
	return d.providerRepo.DeleteOne(ctx, repos.Filter{
		"metadata.name": name,
	})
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
	_, err := d.providerRepo.UpdateOne(ctx, repos.Filter{
		"metadata.name": cloudProvider.Metadata.Name,
	}, cloudProvider)

	return err
}
