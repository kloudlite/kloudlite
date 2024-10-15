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

// just a compile-time validation, that these types satisfies resource
var (
	_ resource = (*entities.App)(nil)
	_ resource = (*entities.ExternalApp)(nil)
	_ resource = (*entities.Config)(nil)
	_ resource = (*entities.Secret)(nil)
	_ resource = (*entities.Router)(nil)
	_ resource = (*entities.ManagedResource)(nil)
	_ resource = (*entities.ImagePullSecret)(nil)
	_ resource = (*entities.Environment)(nil)
)

func (d *domain) upsertEnvironmentResourceMapping(ctx ResourceContext, res resource) (*entities.ResourceMapping, error) {
	clusterName, err := d.getClusterAttachedToEnvironment(ctx, ctx.EnvironmentName)
	if err != nil {
		return nil, errors.NewE(err)
	}
	if clusterName == nil || *clusterName == "" {
		// silent exit
		return nil, nil
	}

	return d.resourceMappingRepo.Upsert(ctx, repos.Filter{
		fields.ClusterName:     clusterName,
		fields.AccountName:     ctx.AccountName,
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

		EnvironmentName: ctx.EnvironmentName,
	})
}

func (d *domain) GetEnvironmentResourceMapping(ctx ConsoleContext, resType entities.ResourceType, clusterName string, namespace string, name string) (*entities.ResourceMapping, error) {
	return d.resourceMappingRepo.FindOne(ctx, repos.Filter{
		fields.AccountName:                  ctx.AccountName,
		fields.ClusterName:                  clusterName,
		fc.ResourceMappingResourceHeirarchy: entities.ResourceHeirarchyEnvironment,
		fc.ResourceMappingResourceType:      resType,
		fc.ResourceMappingResourceName:      name,
		fc.ResourceMappingResourceNamespace: namespace,
	})
}
