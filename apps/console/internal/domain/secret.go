package domain

import (
	"context"
	"fmt"

	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateSecret(ctx context.Context, secret entities.Secret) (*entities.Secret, error) {
	secret.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &secret.Secret); err != nil {
		return nil, err
	}

	s, err := d.secretRepo.Create(ctx, &secret)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &s.Secret); err != nil {
		return s, err
	}

	return s, nil
}

func (d *domain) DeleteSecret(ctx context.Context, namespace string, name string) error {
	s, err := d.findSecret(ctx, namespace, name)
	if err != nil {
		return err
	}
	return d.k8sYamlClient.DeleteResource(ctx, &s.Secret)
}

func (d *domain) GetSecret(ctx context.Context, namespace string, name string) (*entities.Secret, error) {
	return d.findSecret(ctx, namespace, name)
}

func (d *domain) GetSecrets(ctx context.Context, namespace string) ([]*entities.Secret, error) {
	return d.secretRepo.Find(ctx, repos.Query{Filter: repos.Filter{"metadata.namespace": namespace}})
}

func (d *domain) UpdateSecret(ctx context.Context, secret entities.Secret) (*entities.Secret, error) {
	secret.EnsureGVK()
	if err := d.k8sExtendedClient.ValidateStruct(ctx, &secret.Secret); err != nil {
		return nil, err
	}

	s, err := d.findSecret(ctx, secret.Namespace, secret.Name)
	if err != nil {
		return nil, err
	}

	status := s.Status
	s.Secret = secret.Secret
	s.Status = status

	upSecret, err := d.secretRepo.UpdateById(ctx, s.Id, s)
	if err != nil {
		return nil, err
	}

	if err := d.applyK8sResource(ctx, &upSecret.Secret); err != nil {
		return upSecret, err
	}

	return upSecret, nil
}

func (d *domain) findSecret(ctx context.Context, namespace string, name string) (*entities.Secret, error) {
	scrt, err := d.secretRepo.FindOne(ctx, repos.Filter{"metadata.namespace": namespace, "metadata.name": name})
	if err != nil {
		return nil, err
	}
	if scrt == nil {
		return nil, fmt.Errorf("no secret with name=%s,namespace=%s found", name, namespace)
	}
	return scrt, nil
}
