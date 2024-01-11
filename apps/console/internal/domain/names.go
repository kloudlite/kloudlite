package domain

import (
	"context"

	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/pkg/errors"
	fn "github.com/kloudlite/api/pkg/functions"
	"github.com/kloudlite/api/pkg/repos"
)

func checkResourceName[T repos.Entity](ctx context.Context, filters repos.Filter, repo repos.DbRepo[T]) (*CheckNameAvailabilityOutput, error) {
	res, err := repo.FindOne(ctx, filters)
	if err != nil {
		return &CheckNameAvailabilityOutput{Result: false}, errors.NewE(err)
	}

	if fn.IsNil(res) {
		return &CheckNameAvailabilityOutput{Result: true}, nil
	}

	return &CheckNameAvailabilityOutput{
		Result:         false,
		SuggestedNames: fn.GenValidK8sResourceNames(filters["metadata.name"].(string), 3),
	}, nil
}

func (d *domain) CheckNameAvailability(ctx context.Context, accountName string, projectName string, environmentName *string, resType entities.ResourceType, name string) (*CheckNameAvailabilityOutput, error) {
	errEnvironmentRequired := func() error {
		return errors.Newf("param environmentName is required for resource type %q", resType)
	}

	if !fn.IsValidK8sResourceName(name) {
		return &CheckNameAvailabilityOutput{
			Result:         false,
			SuggestedNames: []string{fn.GenReadableName(name), fn.GenReadableName(name), fn.GenReadableName(name)},
		}, nil
	}

	switch resType {
	case entities.ResourceTypeProject:
		{
			return checkResourceName(ctx, repos.Filter{"accountName": accountName, "metadata.name": name}, d.projectRepo)
		}

	case entities.ResourceTypeProjectManagedService:
		{
			return checkResourceName(ctx, repos.Filter{"accountName": accountName, "projectName": projectName, "metadata.name": name}, d.pmsRepo)
		}
	case entities.ResourceTypeEnvironment:
		{
			return checkResourceName(ctx, repos.Filter{"accountName": accountName, "projectName": projectName, "metadata.name": name}, d.environmentRepo)
		}
	default:
		if environmentName == nil {
			return nil, errEnvironmentRequired()
		}

		filter := repos.Filter{
			"accountName":     accountName,
			"projectName":     projectName,
			"environmentName": environmentName,
			"metadata.name":   name,
		}

		switch resType {
		case entities.ResourceTypeApp:
			return checkResourceName(ctx, filter, d.appRepo)
		case entities.ResourceTypeConfig:
			return checkResourceName(ctx, filter, d.configRepo)
		case entities.ResourceTypeSecret:
			return checkResourceName(ctx, filter, d.secretRepo)
		case entities.ResourceTypeRouter:
			return checkResourceName(ctx, filter, d.routerRepo)
		case entities.ResourceTypeManagedResource:
			return checkResourceName(ctx, filter, d.mresRepo)
		case entities.ResourceTypeImagePullSecret:
			return checkResourceName(ctx, filter, d.pullSecretsRepo)
		default:
			{
				return nil, errors.Newf("resource type %q is not acknowledged", resType)
			}
		}
	}
}
