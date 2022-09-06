package domain

import (
	"context"
	"kloudlite.io/apps/console/internal/domain/entities"
	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/pkg/repos"
)

func (d *domain) GetCloudProviders(ctx context.Context, accountId repos.ID) ([]*entities.CloudProvider, error) {

	if err := d.checkAccountAccess(ctx, accountId, READ_ACCOUNT); err != nil {
		return nil, err
	}

	providers, err := d.providerRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"$or": []repos.Filter{
				{
					"account_id": accountId,
				},
				{
					"account_id": nil,
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return providers, nil
}

func (d *domain) CreateCloudProvider(ctx context.Context, accountId *repos.ID, provider *entities.CloudProvider) error {

	if accountId != nil {
		if err := d.checkAccountAccess(ctx, *accountId, "update_account"); err != nil {
			return err
		}
	}

	if accountId != nil {
		provider.AccountId = accountId
	}

	_, err := d.providerRepo.Create(ctx, provider)
	if err != nil {
		return err
	}
	return nil
}

func (d *domain) CreateRegion(ctx context.Context, region *entities.EdgeRegion) error {
	provider, err := d.providerRepo.FindById(ctx, region.ProviderId)
	if err = mongoError(err, "provider not found"); err != nil {
		return err
	}

	if err = d.checkAccountAccess(ctx, *provider.AccountId, "update_account"); err != nil {
		return err
	}

	_, err = d.regionRepo.Create(ctx, region)
	if err != nil {
		return err
	}

	d.workloadMessenger.SendAction("apply", string(region.Id), &op_crds.Region{
		APIVersion: op_crds.RegionAPIVersion,
		Kind:       op_crds.RegionKind,
		Metadata: op_crds.RegionMetadata{
			Name: region.Region,
		},
		Spec: op_crds.RegionSpec{
			Name: region.Region,
			Account: func() *string {
				if provider.AccountId != nil {
					s := string(*provider.AccountId)
					return &s
				}
				return nil
			}(),
		},
	})
	return nil
}

func (d *domain) GetRegions(ctx context.Context, providerId repos.ID) ([]*entities.EdgeRegion, error) {

	provider, err := d.providerRepo.FindById(ctx, providerId)
	if err = mongoError(err, "provider not found"); err != nil {
		return nil, err
	}
	if provider.AccountId != nil {
		if err = d.checkAccountAccess(ctx, *provider.AccountId, READ_ACCOUNT); err != nil {
			return nil, err
		}
	}

	regions, err := d.regionRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"provider_id": providerId,
		},
	})
	if err != nil {
		return nil, err
	}

	return regions, nil
}
