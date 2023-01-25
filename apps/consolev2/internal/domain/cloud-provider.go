package domain

import (
	"context"
	"fmt"
	crdsv1 "github.com/kloudlite/operator/apis/crds/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/pkg/constants"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateCloudProvider(ctx context.Context, cp *entities.CloudProvider, creds entities.SecretData) (*entities.CloudProvider, error) {
	var err error

	cp.FillTypeMeta()
	if cp, err = d.providerRepo.Create(ctx, cp); err != nil {
		return nil, err
	}

	clusterId, err := d.getClusterForAccount(ctx, repos.ID(cp.Spec.AccountId))
	if err != nil {
		return nil, err
	}

	scrt := entities.Secret{
		Secret: crdsv1.Secret{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("provider-%s", cp.Name),
				Namespace: constants.NamespaceCore,
			},
			Spec: crdsv1.SecretSpec{
				Data: creds,
			},
		},
	}

	if _, err := d.secretRepo.Create(ctx, &scrt); err != nil {
		return nil, err
	}

	if err = d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId), string(cp.Id), scrt,
	); err != nil {
		return nil, err
	}

	cp.Spec.ProviderSecretRef = corev1.SecretReference{
		Name:      scrt.Name,
		Namespace: constants.NamespaceCore,
	}

	if err = d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId), string(cp.Id), cp.CloudProvider,
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

func (d *domain) UpdateCloudProvider(ctx context.Context, cloudProvider entities.CloudProvider, creds entities.SecretData) (*entities.CloudProvider, error) {
	var cp *entities.CloudProvider
	var err error

	if cp, err = d.providerRepo.FindOne(ctx, repos.Filter{
		"metadata.name": cloudProvider.Name,
	}); err != nil {
		return nil, err
	}

	if creds != nil {
		one, _ := d.secretRepo.FindOne(ctx, repos.Filter{
			"metadata.name": fmt.Sprintf("provider-%s", cp.Name),
		})

		if one == nil {
			_, err := d.secretRepo.Create(ctx, &entities.Secret{
				Secret: crdsv1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fmt.Sprintf("provider-%s", cp.Name),
						Namespace: constants.NamespaceCore,
					},
					Spec: crdsv1.SecretSpec{
						Data: creds,
					},
					Enabled: false,
				},
			})
			if err != nil {
				return nil, err
			}
		} else {
			one.Spec.Data = creds.ToMap()
			_, err := d.secretRepo.UpdateById(ctx, one.Id, one)
			if err != nil {
				return nil, err
			}
		}
	}

	clusterId, err := d.getClusterForAccount(ctx, repos.ID(cp.Spec.AccountId))
	if err != nil {
		return nil, err
	}

	if _, err := d.providerRepo.UpdateOne(ctx, repos.Filter{
		"metadata.name": cp.Name,
	}, cp); err != nil {
		return nil, err
	}

	if err := d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId),
		string(cp.Id),
		cp.CloudProvider,
	); err != nil {
		return nil, err
	}

	return cp, nil
}
