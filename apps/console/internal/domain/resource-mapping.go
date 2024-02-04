package domain

import (
	"github.com/kloudlite/api/apps/console/internal/entities"
	fc "github.com/kloudlite/api/apps/console/internal/entities/field-constants"
	"github.com/kloudlite/api/common"
	"github.com/kloudlite/api/common/fields"
	"github.com/kloudlite/api/pkg/errors"
	"github.com/kloudlite/api/pkg/repos"
)

type resource interface {
	common.ResourceForSync
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
		fields.ClusterName: clusterName,
		fields.AccountName:     ctx.AccountName,
		fields.ProjectName:     ctx.ProjectName,
		fields.EnvironmentName: ctx.EnvironmentName,

		fc.ResourceMappingResourceHeirarchy: entities.ResourceHeirarchyEnvironment,
		fc.ResourceMappingResourceType:      res.GetResourceType(),
		fc.ResourceMappingResourceName:      res.GetName(),
		fc.ResourceMappingResourceNamespace: res.GetNamespace(),

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
		fields.AccountName: ctx.AccountName,
		fields.ClusterName: *clusterName,
		fields.ProjectName: projectName,
		fc.ResourceMappingResourceHeirarchy: entities.ResourceHeirarchyProject,
		fc.ResourceMappingResourceType:      res.GetResourceType(),
		fc.ResourceMappingResourceName:      res.GetName(),
		fc.ResourceMappingResourceNamespace: res.GetNamespace(),
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
		fields.AccountName:                      ctx.AccountName,
		fields.ClusterName:                      clusterName,
		fc.ResourceMappingResourceHeirarchy: entities.ResourceHeirarchyEnvironment,
		fc.ResourceMappingResourceType:      resType,
		fc.ResourceMappingResourceName:      name,
		fc.ResourceMappingResourceNamespace: namespace,
	})
}

func (d *domain) GetProjectResourceMapping(ctx ConsoleContext, resType entities.ResourceType, clusterName string, name string) (*entities.ResourceMapping, error) {
	return d.resourceMappingRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:                      ctx.AccountName,
		fields.ClusterName:                      clusterName,
		fc.ResourceMappingResourceHeirarchy: entities.ResourceHeirarchyProject,
		fc.ResourceMappingResourceType:      resType,
		fc.ResourceMappingResourceName:      name,
	})
}
