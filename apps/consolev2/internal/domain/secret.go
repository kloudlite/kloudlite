package domain

import (
	"context"
	"fmt"
	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateSecret(ctx context.Context, secret entities.Secret) (*entities.Secret, error) {
	exists, err := d.secretRepo.Exists(ctx, repos.Filter{"metadata.name": secret.Name, "metadata.namespace": secret.Namespace})
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.Newf("secret  %s already exists", secret.Name)
	}

	clusterId, err := d.getClusterForProject(ctx, secret.ProjectName)

	scrt, err := d.secretRepo.Create(ctx, &secret)
	if err != nil {
		return nil, err
	}

	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(scrt.Id), scrt); err != nil {
		return nil, err
	}

	return scrt, nil
}

func (d *domain) UpdateSecret(ctx context.Context, secret entities.Secret) (*entities.Secret, error) {
	existingSecret, err := d.secretRepo.FindOne(ctx, repos.Filter{"metadata.name": secret.Name, "metadata.namespace": secret.Namespace})
	if err != nil {
		return nil, err
	}
	if existingSecret == nil {
		return nil, errors.Newf("secret %s does not exist", secret.Name)
	}

	existingSecret.Secret = secret.Secret
	uScrt, err := d.secretRepo.UpdateById(ctx, existingSecret.Id, &secret)
	if err != nil {
		return nil, err
	}

	clusterId, err := d.getClusterForProject(ctx, secret.ProjectName)

	if err := d.workloadMessenger.SendAction(ActionApply, d.getDispatchKafkaTopic(clusterId), string(existingSecret.Id), secret.Secret); err != nil {
		return nil, err
	}

	return uScrt, nil
}

func (d *domain) DeleteSecret(ctx context.Context, namespace string, name string) (bool, error) {
	if err := d.secretRepo.DeleteOne(ctx, repos.Filter{"metadata.namespace": namespace, "metadata.name": name}); err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) GetSecrets(ctx context.Context, namespace string, search *string) ([]*entities.Secret, error) {
	if search == nil {
		return d.secretRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": namespace}})
	}
	return d.secretRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": namespace, "metadata.name": fmt.Sprintf("/%s/", *search)}})
}

func (d *domain) GetSecret(ctx context.Context, namespace string, name string) (*entities.Secret, error) {
	return d.secretRepo.FindOne(ctx, repos.Filter{"metadata.name": name, "metadata.namespace": namespace})
}
