package domain

import (
	"context"
	"fmt"

	"kloudlite.io/apps/console/internal/domain/entities"
	op_crds "kloudlite.io/apps/console/internal/domain/op-crds"
	"kloudlite.io/pkg/repos"
)

func (d *domain) CreateEnvironment(ctx context.Context, blueprintID *repos.ID, name string, readableId string) (*entities.Environment, error) {

	p, err := d.projectRepo.FindById(ctx, *blueprintID)
	if err != nil {
		return nil, err
	}

	env, err := d.environmentRepo.Create(ctx, &entities.Environment{
		BlueprintId: blueprintID,
		Name:        name,
		ReadableId:  readableId,
	})
	if err != nil {
		return nil, err
	}

	clusterId, err := d.getClusterForAccount(ctx, p.AccountId)
	if err != nil {
		return nil, err
	}

	d.workloadMessenger.SendAction("apply", d.getDispatchKafkaTopic(clusterId), string(env.Id), &op_crds.Environment{
		APIVersion: op_crds.APIVersion,
		Kind:       op_crds.EnvKind,
		Metadata: op_crds.EnvMetadata{
			Name: fmt.Sprintf("%s-%s", p.Name, env.ReadableId),
			Annotations: map[string]string{
				"kloudlite.io/account-ref":  string(p.AccountId),
				"kloudlite.io/resource-ref": string(env.Id),
			},
			Labels: map[string]string{
				"kloudlite.io/account-ref":  string(p.AccountId),
				"kloudlite.io/resource-ref": string(env.Id),
			},
		},
		Spec: op_crds.EnvSpec{
			ProjectName:   string(p.Name),
			BlueprintName: string(p.Name) + "-blueprint",
			AccountRef:    string(p.AccountId),
			// RouterBaseDomain: "example.com",
		},
	})

	return env, nil

}

func (d *domain) GetEnvironment(ctx context.Context, envId repos.ID) (*entities.Environment, error) {
	return d.environmentRepo.FindById(ctx, envId)
}
