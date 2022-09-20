package domain

import (
	"context"
	"kloudlite.io/apps/console/internal/domain/entities"
	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/pkg/repos"
)

type CloudProviderUpdate struct {
	Name        *string
	Credentials map[string]string
}

type EdgeRegionUpdate struct {
	Name      *string
	NodePools []entities.NodePool
}

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
					"account_id": "kl-core",
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
	} else {
		k := repos.ID("kl-core")
		accountId = &k
	}

	provider.AccountId = accountId
	_, err := d.providerRepo.Create(ctx, provider)
	if err != nil {
		return err
	}

	err = d.workloadMessenger.SendAction("apply", string(provider.Id), &op_crds.Secret{
		APIVersion: op_crds.SecretAPIVersion,
		Kind:       op_crds.SecretKind,
		Metadata: op_crds.SecretMetadata{
			Name:      "provider-" + string(provider.Id),
			Namespace: "wg-" + string(*provider.AccountId),
		},
		StringData: func() map[string]any {
			data := make(map[string]any)
			for k, v := range provider.Credentials {
				data[k] = v
			}
			return data
		}(),
	})

	if err != nil {
		return err
	}

	return nil
}
func (d *domain) DeleteCloudProvider(ctx context.Context, providerId repos.ID) error {
	provider, err := d.providerRepo.FindById(ctx, providerId)
	if err != nil {
		return err
	}
	if err := d.checkAccountAccess(ctx, *provider.AccountId, "update_account"); err != nil {
		return err
	}
	err = d.workloadMessenger.SendAction("delete", string(provider.Id), &op_crds.Secret{
		APIVersion: op_crds.SecretAPIVersion,
		Kind:       op_crds.SecretKind,
		Metadata: op_crds.SecretMetadata{
			Name: "provider-" + string(provider.Id),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (d *domain) UpdateCloudProvider(ctx context.Context, providerId repos.ID, update *CloudProviderUpdate) error {
	provider, err := d.providerRepo.FindById(ctx, providerId)
	if err != nil {
		return err
	}
	if err := d.checkAccountAccess(ctx, *provider.AccountId, "update_account"); err != nil {
		return err
	}

	if update.Name != nil {
		provider.Name = *update.Name
	}

	if update.Credentials != nil {
		provider.Credentials = update.Credentials
	}

	_, err = d.providerRepo.UpdateById(ctx, providerId, provider)
	if err != nil {
		return err
	}
	err = d.workloadMessenger.SendAction("apply", string(provider.Id), &op_crds.Secret{
		APIVersion: op_crds.SecretAPIVersion,
		Kind:       op_crds.SecretKind,
		Metadata: op_crds.SecretMetadata{
			Name:      "provider-" + string(provider.Id),
			Namespace: "wg-" + string(*provider.AccountId),
		},
		StringData: func() map[string]any {
			data := make(map[string]any)
			for k, v := range provider.Credentials {
				data[k] = v
			}
			return data
		}(),
	})

	if err != nil {
		return err
	}
	return nil
}
func (d *domain) CreateEdgeRegion(ctx context.Context, providerId repos.ID, region *entities.EdgeRegion) error {
	provider, err := d.providerRepo.FindById(ctx, region.ProviderId)
	if err = mongoError(err, "provider not found"); err != nil {
		return err
	}

	if err = d.checkAccountAccess(ctx, *provider.AccountId, "update_account"); err != nil {
		return err
	}

	createdRegion, err := d.regionRepo.Create(ctx, region)
	if err != nil {
		return err
	}

	err = d.workloadMessenger.SendAction("apply", string(region.Id), &op_crds.Region{
		APIVersion: op_crds.EdgeAPIVersion,
		Kind:       op_crds.EdgeKind,
		Metadata: op_crds.EdgeMetadata{
			Name: "region-" + string(createdRegion.Id),
		},
		Spec: op_crds.EdgeSpec{
			CredentialsRef: op_crds.CredentialsRef{
				SecretName: "provider-" + string(provider.Id),
				Namespace:  "wg-" + string(*provider.AccountId),
			},
			AccountId: func() *string {
				if provider.AccountId != nil {
					s := string(*provider.AccountId)
					return &s
				}
				return nil
			}(),
			Provider: provider.Provider,
			Region:   region.Region,
			Pools: func() []op_crds.NodePool {
				var pools []op_crds.NodePool
				for _, pool := range region.Pools {
					pools = append(pools, op_crds.NodePool{
						Name:   pool.Name,
						Min:    pool.Min,
						Max:    pool.Max,
						Config: pool.Config,
					})
				}
				return pools
			}(),
		},
	})
	if err != nil {
		return err
	}
	return nil
}
func (d *domain) GetEdgeRegions(ctx context.Context, providerId repos.ID) ([]*entities.EdgeRegion, error) {
	edgeRegions, err := d.regionRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"provider_id": providerId,
		},
	})
	if err != nil {
		return nil, err
	}
	return edgeRegions, nil
}
func (d *domain) DeleteEdgeRegion(ctx context.Context, edgeId repos.ID) error {
	edge, err := d.regionRepo.FindById(ctx, edgeId)
	if err != nil {
		return err
	}

	provider, err := d.providerRepo.FindById(ctx, edge.ProviderId)
	if err = mongoError(err, "provider not found"); err != nil {
		return err
	}
	if provider.AccountId != nil {
		if err = d.checkAccountAccess(ctx, *provider.AccountId, READ_ACCOUNT); err != nil {
			return err
		}
	}

	err = d.workloadMessenger.SendAction("delete", string(edge.Id), &op_crds.Region{
		APIVersion: op_crds.EdgeAPIVersion,
		Kind:       op_crds.EdgeKind,
		Metadata: op_crds.EdgeMetadata{
			Name: "region-" + string(edge.Id),
		},
	})
	if err != nil {
		return err
	}
	return nil
}
func (d *domain) UpdateEdgeRegion(ctx context.Context, edgeId repos.ID, update *EdgeRegionUpdate) error {
	region, err := d.regionRepo.FindById(ctx, edgeId)
	if err != nil {
		return err
	}
	provider, err := d.providerRepo.FindById(ctx, region.ProviderId)
	if err = mongoError(err, "provider not found"); err != nil {
		return err
	}

	if err = d.checkAccountAccess(ctx, *provider.AccountId, "update_account"); err != nil {
		return err
	}

	if update.Name != nil {
		region.Name = *update.Name
	}

	if update.NodePools != nil {
		region.Pools = update.NodePools
	}

	createdRegion, err := d.regionRepo.UpdateById(ctx, edgeId, region)
	if err != nil {
		return err
	}

	d.workloadMessenger.SendAction("apply", string(region.Id), &op_crds.Region{
		APIVersion: op_crds.EdgeAPIVersion,
		Kind:       op_crds.EdgeKind,
		Metadata: op_crds.EdgeMetadata{
			Name: "region-" + string(createdRegion.Id),
		},
		Spec: op_crds.EdgeSpec{
			CredentialsRef: op_crds.CredentialsRef{
				SecretName: "provider-" + string(provider.Id),
				Namespace:  "wg-" + string(*provider.AccountId),
			},
			AccountId: func() *string {
				if provider.AccountId != nil {
					s := string(*provider.AccountId)
					return &s
				}
				return nil
			}(),
			Provider: provider.Provider,
			Region:   region.Region,
			Pools: func() []op_crds.NodePool {
				var pools []op_crds.NodePool
				for _, pool := range region.Pools {
					pools = append(pools, op_crds.NodePool{
						Name:   pool.Name,
						Min:    pool.Min,
						Max:    pool.Max,
						Config: pool.Config,
					})
				}
				return pools
			}(),
		},
	})
	return nil

}
