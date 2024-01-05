package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	"github.com/kloudlite/api/pkg/repos"
)

type resource interface {
	GetName() string
	GetNamespace() string
	GetResourceType() entities.ResourceType
}

var _ resource = (*entities.App)(nil)
var _ resource = (*entities.Config)(nil)
var _ resource = (*entities.Secret)(nil)
var _ resource = (*entities.Router)(nil)
var _ resource = (*entities.ManagedResource)(nil)
var _ resource = (*entities.ImagePullSecret)(nil)
var _ resource = (*entities.Project)(nil)
var _ resource = (*entities.Environment)(nil)

func (d *domain) upsertResourceMapping(ctx ResourceContext, res resource) (*entities.ResourceMapping, error) {
	return d.resourceMappingRepo.Upsert(ctx, repos.Filter{
		"resourceType":      res.GetResourceType(),
		"resourceName":      res.GetName(),
		"resourceNamespace": res.GetNamespace(),

		"accountName":     ctx.AccountName,
		"projectName":     ctx.ProjectName,
		"environmentName": ctx.EnvironmentName,
	}, &entities.ResourceMapping{
		ResourceType:      res.GetResourceType(),
		ResourceName:      res.GetName(),
		ResourceNamespace: res.GetNamespace(),

		AccountName:     ctx.AccountName,
		ProjectName:     ctx.ProjectName,
		EnvironmentName: ctx.EnvironmentName,
	})
}

func (d *domain) GetResourceMapping(ctx ConsoleContext, resType entities.ResourceType, namespace string, name string) (*entities.ResourceMapping, error) {
	return d.resourceMappingRepo.FindOne(ctx, repos.Filter{
		"accountName":       ctx.AccountName,
		"resourceType":      resType,
		"resourceName":      name,
		"resourceNamespace": namespace,
	})
}
