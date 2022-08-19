package domain

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	"kloudlite.io/apps/console/internal/domain/entities"
	opcrds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/pkg/repos"
	"strings"
	"time"
)

func (d *domain) GetSecret(ctx context.Context, secretId repos.ID) (*entities.Secret, error) {
	sec, err := d.secretRepo.FindById(ctx, secretId)
	if err != nil {
		return nil, err
	}
	return sec, nil
}
func (d *domain) GetSecrets(ctx context.Context, projectId repos.ID) ([]*entities.Secret, error) {
	secrets, err := d.secretRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"project_id": projectId,
		},
	})
	if err != nil {
		return nil, err
	}
	return secrets, nil
}
func (d *domain) CreateSecret(ctx context.Context, projectId repos.ID, secretName string, desc *string, secretData []*entities.Entry) (*entities.Secret, error) {
	prj, err := d.projectRepo.FindById(ctx, projectId)
	if err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, fmt.Errorf("project not found")
	}
	create, err := d.secretRepo.Create(ctx, &entities.Secret{
		Name:        strings.ToLower(secretName),
		ProjectId:   projectId,
		Namespace:   prj.Name,
		Data:        secretData,
		Description: desc,
	})
	if err != nil {
		return nil, err
	}
	err = d.workloadMessenger.SendAction("apply", string(create.Id), opcrds.Secret{
		APIVersion: opcrds.SecretAPIVersion,
		Kind:       opcrds.SecretKind,
		Metadata: opcrds.SecretMetadata{
			Name:      secretName,
			Namespace: prj.Name,
		},
		Data: nil,
	})
	if err != nil {
		return nil, err
	}
	return create, nil
}
func (d *domain) UpdateSecret(ctx context.Context, secretId repos.ID, desc *string, secretData []*entities.Entry) (bool, error) {
	cfg, err := d.secretRepo.FindById(ctx, secretId)
	if err != nil {
		return false, err
	}
	if cfg == nil {
		return false, fmt.Errorf("config not found")
	}
	if desc != nil {
		cfg.Description = desc
	}
	cfg.Data = secretData
	_, err = d.secretRepo.UpdateById(ctx, secretId, cfg)
	if err != nil {
		return false, err
	}
	err = d.workloadMessenger.SendAction("apply", string(cfg.Id), opcrds.Secret{
		APIVersion: opcrds.SecretAPIVersion,
		Kind:       opcrds.SecretKind,
		Metadata: opcrds.SecretMetadata{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
		},
		Data: (func() map[string]any {
			data := make(map[string]any, 0)
			for _, d := range cfg.Data {
				encoded := b64.StdEncoding.EncodeToString([]byte(d.Value))
				data[d.Key] = encoded
			}
			return data
		})(),
	})
	if err != nil {
		return false, err
	}
	return true, nil
}
func (d *domain) DeleteSecret(ctx context.Context, secretId repos.ID) (bool, error) {
	secret, err := d.secretRepo.FindById(ctx, secretId)
	if err != nil {
		return false, err
	}
	err = d.secretRepo.DeleteById(ctx, secretId)
	if err != nil {
		return false, err
	}
	err = d.workloadMessenger.SendAction("delete", string(secretId), opcrds.Config{
		APIVersion: opcrds.ConfigAPIVersion,
		Kind:       opcrds.ConfigKind,
		Metadata: opcrds.ConfigMetadata{
			Name:      secret.Name,
			Namespace: secret.Namespace,
		},
	})
	return true, nil
}

func (d *domain) GetConfig(ctx context.Context, configId repos.ID) (*entities.Config, error) {
	cfg, err := d.configRepo.FindById(ctx, configId)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
func (d *domain) GetConfigs(ctx context.Context, projectId repos.ID) ([]*entities.Config, error) {
	configs, err := d.configRepo.Find(ctx, repos.Query{
		Filter: repos.Filter{
			"project_id": projectId,
		},
	})
	if err != nil {
		return nil, err
	}
	return configs, nil
}
func (d *domain) CreateConfig(ctx context.Context, projectId repos.ID, configName string, desc *string, configData []*entities.Entry) (*entities.Config, error) {
	prj, err := d.projectRepo.FindById(ctx, projectId)
	if err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, fmt.Errorf("project not found")
	}
	create, err := d.configRepo.Create(ctx, &entities.Config{
		Name:        strings.ToLower(configName),
		ProjectId:   projectId,
		Namespace:   prj.Name,
		Data:        configData,
		Description: desc,
	})
	if err != nil {
		return nil, err
	}
	err = d.workloadMessenger.SendAction("apply", string(create.Id), opcrds.Config{
		APIVersion: opcrds.ConfigAPIVersion,
		Kind:       opcrds.ConfigKind,
		Metadata: opcrds.ConfigMetadata{
			Name:      string(create.Id),
			Namespace: prj.Name,
		},
		Data: nil,
	})
	time.AfterFunc(3*time.Second, func() {
		fmt.Println("send apply config")
		d.notifier.Notify(create.Id)
	})
	if err != nil {
		return nil, err
	}
	return create, nil
}
func (d *domain) UpdateConfig(ctx context.Context, configId repos.ID, desc *string, configData []*entities.Entry) (bool, error) {
	cfg, err := d.configRepo.FindById(ctx, configId)
	if err != nil {
		return false, err
	}
	if cfg == nil {
		return false, fmt.Errorf("config not found")
	}
	if desc != nil {
		cfg.Description = desc
	}
	cfg.Data = configData
	cfg.Status = entities.ConfigStateSyncing
	_, err = d.configRepo.UpdateById(ctx, configId, cfg)
	if err != nil {
		return false, err
	}
	err = d.workloadMessenger.SendAction("apply", string(cfg.Id), opcrds.Config{
		APIVersion: opcrds.ConfigAPIVersion,
		Kind:       opcrds.ConfigKind,
		Metadata: opcrds.ConfigMetadata{
			Name:      string(cfg.Id),
			Namespace: cfg.Namespace,
		},
		Data: func() map[string]any {
			m := make(map[string]any, 0)
			for _, i := range cfg.Data {
				m[i.Key] = i.Value
			}
			return m
		}(),
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (d *domain) DeleteConfig(ctx context.Context, configId repos.ID) (bool, error) {
	cfg, err := d.configRepo.FindById(ctx, configId)
	if err != nil {
		return false, err
	}
	err = d.configRepo.DeleteById(ctx, configId)
	if err != nil {
		return false, err
	}
	err = d.workloadMessenger.SendAction("delete", string(configId), opcrds.Config{
		APIVersion: opcrds.ConfigAPIVersion,
		Kind:       opcrds.ConfigKind,
		Metadata: opcrds.ConfigMetadata{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
		},
	})
	return true, nil
}
