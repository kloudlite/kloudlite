package domain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"kloudlite.io/apps/console/internal/domain/entities"
	opcrds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/constants"
	"kloudlite.io/pkg/beacon"
	"kloudlite.io/pkg/repos"
)

func (d *domain) GetSecret(ctx context.Context, secretId repos.ID) (*entities.Secret, error) {
	sec, err := d.secretRepo.FindById(ctx, secretId)
	if err = mongoError(err, "secret not found"); err != nil {
		return nil, err
	}

	err = d.checkProjectAccess(ctx, sec.ProjectId, ReadProject)
	if err != nil {
		return nil, err
	}
	return sec, nil
}

func (d *domain) GetSecrets(ctx context.Context, projectId repos.ID) ([]*entities.Secret, error) {
	err := d.checkProjectAccess(ctx, projectId, ReadProject)
	if err != nil {
		return nil, err
	}
	secrets, err := d.secretRepo.Find(
		ctx, repos.Query{
			Filter: repos.Filter{
				"project_id": projectId,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}

func (d *domain) CreateSecret(ctx context.Context, projectId repos.ID, secretName string, desc *string, secretData []*entities.Entry) (*entities.Secret, error) {
	err := d.checkProjectAccess(ctx, projectId, UpdateProject)
	if err != nil {
		return nil, err
	}

	prj, err := d.projectRepo.FindById(ctx, projectId)
	if err = mongoError(err, "project not found"); err != nil {
		return nil, err
	}

	scrt, err := d.secretRepo.Create(
		ctx, &entities.Secret{
			Name:        strings.ToLower(secretName),
			ProjectId:   projectId,
			Namespace:   prj.Name + "-blueprint",
			Data:        secretData,
			Description: desc,
		},
	)
	if err != nil {
		return nil, err
	}

	clusterId, err := d.getClusterForAccount(ctx, prj.AccountId)
	if err != nil {
		return nil, err
	}

	if err = d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId), string(scrt.Id), opcrds.Secret{
			APIVersion: opcrds.SecretAPIVersion,
			Kind:       opcrds.SecretKind,
			Metadata: opcrds.SecretMetadata{
				Name:      string(scrt.Id),
				Namespace: prj.Name + "-blueprint",
				Annotations: map[string]string{
					"kloudlite.io/account-ref":  string(prj.AccountId),
					"kloudlite.io/project-ref":  string(prj.Id),
					"kloudlite.io/resource-ref": string(scrt.Id),
				},
			},
			Data: nil,
		},
	); err != nil {
		return nil, err
	}

	accountId, err := d.getAccountIdForProject(ctx, scrt.ProjectId)
	if err != nil {
		return nil, err
	}

	go d.beacon.TriggerWithUserCtx(ctx, accountId, beacon.EventAction{
		Action:       constants.CreateSecret,
		Status:       beacon.StatusOK(),
		ResourceType: constants.ResourceSecret,
		ResourceId:   scrt.Id,
		Tags:         map[string]string{"projectId": string(scrt.ProjectId)},
	})

	return scrt, nil
}

func (d *domain) UpdateSecret(ctx context.Context, secretId repos.ID, desc *string, secretData []*entities.Entry) (bool, error) {
	secret, err := d.secretRepo.FindById(ctx, secretId)
	if err = mongoError(err, "secret not found"); err != nil {
		return false, err
	}

	err = d.checkProjectAccess(ctx, secret.ProjectId, UpdateProject)
	if err != nil {
		return false, err
	}

	if desc != nil {
		secret.Description = desc
	}

	secret.Data = secretData
	_, err = d.secretRepo.UpdateById(ctx, secretId, secret)
	if err != nil {
		return false, err
	}

	clusterId, err := d.getClusterIdForProject(ctx, secret.ProjectId)
	if err != nil {
		return false, err
	}

	if err = d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId), string(secret.Id), opcrds.Secret{
			APIVersion: opcrds.SecretAPIVersion,
			Kind:       opcrds.SecretKind,
			Metadata: opcrds.SecretMetadata{
				Name:      string(secret.Id),
				Namespace: secret.Namespace + "-blueprint",
				Annotations: map[string]string{
					"kloudlite.io/project-ref":  string(secret.ProjectId),
					"kloudlite.io/resource-ref": string(secret.Id),
				},
			},
			Data: (func() map[string][]byte {
				data := make(map[string][]byte, 0)
				for _, d := range secret.Data {
					// encoded := b64.StdEncoding.EncodeToString([]byte(d.Value))
					data[d.Key] = []byte(d.Value)
				}
				return data
			})(),
		},
	); err != nil {
		return false, err
	}

	accountId, err := d.getAccountIdForProject(ctx, secret.ProjectId)
	if err != nil {
		return false, err
	}

	go d.beacon.TriggerWithUserCtx(ctx, accountId, beacon.EventAction{
		Action:       constants.UpdateSecret,
		Status:       beacon.StatusOK(),
		ResourceType: constants.ResourceSecret,
		ResourceId:   secretId,
		Tags:         map[string]string{"projectId": string(secret.ProjectId)},
	})

	return true, nil
}

func (d *domain) DeleteSecret(ctx context.Context, secretId repos.ID) (bool, error) {
	secret, err := d.secretRepo.FindById(ctx, secretId)
	if err = mongoError(err, "secret not found"); err != nil {
		return false, err
	}

	err = d.checkProjectAccess(ctx, secret.ProjectId, UpdateProject)
	if err != nil {
		return false, err
	}

	err = d.secretRepo.DeleteById(ctx, secretId)
	if err != nil {
		return false, err
	}

	clusterId, err := d.getClusterIdForProject(ctx, secret.ProjectId)
	if err != nil {
		return false, err
	}

	if err = d.workloadMessenger.SendAction(
		"delete", d.getDispatchKafkaTopic(clusterId), string(secretId), opcrds.Config{
			APIVersion: opcrds.ConfigAPIVersion,
			Kind:       opcrds.ConfigKind,
			Metadata: opcrds.ConfigMetadata{
				Name:      string(secret.Id),
				Namespace: secret.Namespace + "-blueprint",
				Annotations: map[string]string{
					"kloudlite.io/project-ref":  string(secret.ProjectId),
					"kloudlite.io/resource-ref": string(secretId),
				},
			},
		},
	); err != nil {
		return false, err
	}

	accountId, err := d.getAccountIdForProject(ctx, secret.ProjectId)
	if err != nil {
		return false, err
	}

	go d.beacon.TriggerWithUserCtx(ctx, accountId, beacon.EventAction{
		Action:       constants.DeleteSecret,
		Status:       beacon.StatusOK(),
		ResourceType: constants.ResourceSecret,
		ResourceId:   secretId,
		Tags:         map[string]string{"projectId": string(secret.ProjectId)},
	})

	return true, nil
}

func (d *domain) GetConfig(ctx context.Context, configId repos.ID) (*entities.Config, error) {
	cfg, err := d.configRepo.FindById(ctx, configId)
	if err = mongoError(err, "config not found"); err != nil {
		return nil, err
	}

	err = d.checkProjectAccess(ctx, cfg.ProjectId, ReadProject)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func (d *domain) GetConfigs(ctx context.Context, projectId repos.ID) ([]*entities.Config, error) {
	err := d.checkProjectAccess(ctx, projectId, ReadProject)
	if err != nil {
		return nil, err
	}
	configs, err := d.configRepo.Find(
		ctx, repos.Query{
			Filter: repos.Filter{
				"project_id": projectId,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

func (d *domain) CreateConfig(ctx context.Context, projectId repos.ID, configName string, desc *string, configData []*entities.Entry) (*entities.Config, error) {
	err := d.checkProjectAccess(ctx, projectId, UpdateProject)
	if err != nil {
		return nil, err
	}
	prj, err := d.projectRepo.FindById(ctx, projectId)
	if err != nil {
		return nil, err
	}
	if prj == nil {
		return nil, fmt.Errorf("project not found")
	}

	cfg, err := d.configRepo.Create(
		ctx, &entities.Config{
			Name:        strings.ToLower(configName),
			ProjectId:   projectId,
			Namespace:   prj.Name + "-blueprint",
			Data:        configData,
			Description: desc,
		},
	)
	if err != nil {
		return nil, err
	}

	clusterId, err := d.getClusterIdForProject(ctx, projectId)
	if err != nil {
		return nil, err
	}

	if err = d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId), string(cfg.Id), opcrds.Config{
			APIVersion: opcrds.ConfigAPIVersion,
			Kind:       opcrds.ConfigKind,
			Metadata: opcrds.ConfigMetadata{
				Name:      string(cfg.Id),
				Namespace: prj.Name + "-blueprint",
				Annotations: map[string]string{
					"kloudlite.io/account-ref":  string(prj.AccountId),
					"kloudlite.io/project-ref":  string(prj.Id),
					"kloudlite.io/resource-ref": string(cfg.Id),
				},
			},
			Data: nil,
		},
	); err != nil {
		return nil, err
	}

	time.AfterFunc(
		3*time.Second, func() {
			fmt.Println("send apply config")
			d.notifier.Notify(cfg.Id)
		},
	)

	go d.beacon.TriggerWithUserCtx(ctx, prj.AccountId, beacon.EventAction{
		Action:       constants.CreateConfig,
		Status:       beacon.StatusOK(),
		ResourceType: constants.ResourceConfig,
		ResourceId:   cfg.Id,
		Tags:         map[string]string{"projectId": string(prj.Id)},
	})

	return cfg, nil
}

func (d *domain) UpdateConfig(ctx context.Context, configId repos.ID, desc *string, configData []*entities.Entry) (bool, error) {
	cfg, err := d.configRepo.FindById(ctx, configId)
	if err = mongoError(err, "config not found"); err != nil {
		return false, err
	}

	err = d.checkProjectAccess(ctx, cfg.ProjectId, UpdateProject)
	if err != nil {
		return false, err
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

	clusterId, err := d.getClusterIdForProject(ctx, cfg.ProjectId)
	if err != nil {
		return false, err
	}

	if err = d.workloadMessenger.SendAction(
		"apply", d.getDispatchKafkaTopic(clusterId), string(cfg.Id), opcrds.Config{
			APIVersion: opcrds.ConfigAPIVersion,
			Kind:       opcrds.ConfigKind,
			Metadata: opcrds.ConfigMetadata{
				Name:      string(cfg.Id),
				Namespace: cfg.Namespace + "-blueprint",
				Annotations: map[string]string{
					"kloudlite.io/project-ref":  string(cfg.ProjectId),
					"kloudlite.io/resource-ref": string(cfg.Id),
				},
			},
			Data: func() map[string]string {
				m := make(map[string]string, 0)
				for _, i := range cfg.Data {
					m[i.Key] = i.Value
				}
				return m
			}(),
		},
	); err != nil {
		return false, err
	}

	accountId, err := d.getAccountIdForProject(ctx, cfg.ProjectId)
	if err != nil {
		return false, err
	}

	go d.beacon.TriggerWithUserCtx(ctx, accountId, beacon.EventAction{
		Action:       constants.UpdateConfig,
		Status:       beacon.StatusOK(),
		ResourceType: constants.ResourceConfig,
		ResourceId:   cfg.Id,
		Tags:         map[string]string{"projectId": string(cfg.ProjectId)},
	})

	return true, nil
}

func (d *domain) DeleteConfig(ctx context.Context, configId repos.ID) (bool, error) {
	cfg, err := d.configRepo.FindById(ctx, configId)
	if err = mongoError(err, "config not found"); err != nil {
		return false, err
	}

	if err = d.checkProjectAccess(ctx, cfg.ProjectId, UpdateProject); err != nil {
		return false, err
	}

	err = d.configRepo.DeleteById(ctx, configId)
	if err != nil {
		return false, err
	}

	clusterId, err := d.getClusterIdForProject(ctx, cfg.ProjectId)
	if err != nil {
		return false, err
	}

	if err = d.workloadMessenger.SendAction(
		"delete", d.getDispatchKafkaTopic(clusterId), string(configId), opcrds.Config{
			APIVersion: opcrds.ConfigAPIVersion,
			Kind:       opcrds.ConfigKind,
			Metadata: opcrds.ConfigMetadata{
				Name:      string(cfg.Id),
				Namespace: cfg.Namespace + "-blueprint",
				Annotations: map[string]string{
					"kloudlite.io/project-ref":  string(cfg.ProjectId),
					"kloudlite.io/resource-ref": string(cfg.Id),
				},
			},
		},
	); err != nil {
		return false, err
	}

	accountId, err := d.getAccountIdForProject(ctx, cfg.ProjectId)
	if err != nil {
		return false, err
	}

	go d.beacon.TriggerWithUserCtx(ctx, accountId, beacon.EventAction{
		Action:       constants.DeleteConfig,
		Status:       beacon.StatusOK(),
		ResourceType: constants.ResourceConfig,
		ResourceId:   cfg.Id,
		Tags:         map[string]string{"projectId": string(cfg.ProjectId)},
	})

	return true, nil
}
