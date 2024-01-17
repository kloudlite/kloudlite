package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

type resource interface {
	GetName() string
	GetNamespace() string
	GetResourceType() entities.ResourceType
}

var (
	_ resource = (*entities.App)(nil)
	_ resource = (*entities.Config)(nil)
	_ resource = (*entities.Secret)(nil)
	_ resource = (*entities.Router)(nil)
	_ resource = (*entities.ManagedResource)(nil)
	_ resource = (*entities.ImagePullSecret)(nil)
	_ resource = (*entities.Project)(nil)
	_ resource = (*entities.Environment)(nil)
	_ resource = (*entities.ProjectManagedService)(nil)
)

func (d *domain) upsertEnvironmentResourceMapping(ctx ResourceContext, res resource) (*entities.ResourceMapping, error) {
	clusterName, err := d.getClusterAttachedToProject(ctx, ctx.ProjectName)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if clusterName == nil {
		// silent exit
		return nil, nil
	}

	return d.resourceMappingRepo.Upsert(ctx, repos.Filter{
		fc.ResourceMappingResourceHeirarchy: entities.ResourceHeirarchyEnvironment,

		fc.ClusterName: clusterName,

		fc.ResourceMappingResourceType:      res.GetResourceType(),
		fc.ResourceMappingResourceName:      res.GetName(),
		fc.ResourceMappingResourceNamespace: res.GetNamespace(),

		fc.AccountName:     ctx.AccountName,
		fc.ProjectName:     ctx.ProjectName,
		fc.EnvironmentName: ctx.EnvironmentName,
	}, &entities.ResourceMapping{
		ResourceHeirarchy: entities.ResourceHeirarchyEnvironment,

		ResourceType:      res.GetResourceType(),
		ResourceName:      res.GetName(),
		ResourceNamespace: res.GetNamespace(),

		AccountName: ctx.AccountName,
		ClusterName: *clusterName,

		ProjectName:     ctx.ProjectName,
		EnvironmentName: ctx.EnvironmentName,
	})
}

func (d *domain) upsertProjectResourceMapping(ctx ConsoleContext, projectName string, res resource) (*entities.ResourceMapping, error) {
	clusterName, err := d.getClusterAttachedToProject(ctx, projectName)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if clusterName == nil {
		// silent exit
		return nil, nil
	}

	return d.resourceMappingRepo.Upsert(ctx, repos.Filter{
		fc.ResourceMappingResourceHeirarchy: entities.ResourceHeirarchyProject,

		fc.ResourceMappingResourceType:      res.GetResourceType(),
		fc.ResourceMappingResourceName:      res.GetName(),
		fc.ResourceMappingResourceNamespace: res.GetNamespace(),

		fc.AccountName: ctx.AccountName,
		fc.ClusterName: *clusterName,

		fc.ProjectName: projectName,
	}, &entities.ResourceMapping{
		ResourceHeirarchy: entities.ResourceHeirarchyProject,

		ResourceType:      res.GetResourceType(),
		ResourceName:      res.GetName(),
		ResourceNamespace: res.GetNamespace(),

		AccountName: ctx.AccountName,
		ClusterName: *clusterName,

		ProjectName: projectName,
	})
}

func (d *domain) GetEnvironmentResourceMapping(ctx ConsoleContext, resType entities.ResourceType, clusterName string, namespace string, name string) (*entities.ResourceMapping, error) {
	return d.resourceMappingRepo.FindOne(ctx, repos.Filter{
		fc.ResourceMappingResourceHeirarchy: entities.ResourceHeirarchyEnvironment,
		fc.AccountName:                      ctx.AccountName,
		fc.ResourceMappingResourceType:      resType,
		fc.ResourceMappingResourceName:      name,
		fc.ClusterName:                      clusterName,
		fc.ResourceMappingResourceNamespace: namespace,
	})
}

func (d *domain) GetProjectResourceMapping(ctx ConsoleContext, resType entities.ResourceType, clusterName string, name string) (*entities.ResourceMapping, error) {
	return d.resourceMappingRepo.FindOne(ctx, repos.Filter{
		fc.ResourceMappingResourceHeirarchy: entities.ResourceHeirarchyProject,
		fc.AccountName:                      ctx.AccountName,
		fc.ClusterName:                      clusterName,
		fc.ResourceMappingResourceType:      resType,
		fc.ResourceMappingResourceName:      name,
	})
}
