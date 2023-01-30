package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"kloudlite.io/apps/consolev2/internal/app/graph/generated"
	"kloudlite.io/apps/consolev2/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (r *mutationResolver) CoreCreateCloudProvider(ctx context.Context, in entities.CloudProvider, creds entities.SecretData) (*entities.CloudProvider, error) {
	return r.Domain.CreateCloudProvider(ctx, &in, creds)
}

func (r *mutationResolver) CoreUpdateCloudProvider(ctx context.Context, in entities.CloudProvider, creds entities.SecretData) (*entities.CloudProvider, error) {
	cp, err := r.Domain.UpdateCloudProvider(ctx, in, creds)
	if err != nil {
		return nil, err
	}
	return cp, nil
}

func (r *mutationResolver) CoreDeleteCloudProvider(ctx context.Context, name string) (bool, error) {
	if err := r.Domain.DeleteCloudProvider(ctx, name); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreSample(ctx context.Context, j map[string]interface{}) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreCreateEdgeRegion(ctx context.Context, edgeRegion entities.EdgeRegion, providerID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreUpdateEdgeRegion(ctx context.Context, edgeID repos.ID, edgeRegion entities.EdgeRegion) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreDeleteEdgeRegion(ctx context.Context, edgeID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreCreateEnvironment(ctx context.Context, environment entities.Environment) (*entities.Environment, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreDeleteEnvironment(ctx context.Context, name string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreCreateProject(ctx context.Context, project entities.Project) (*entities.Project, error) {
	return r.Domain.CreateProject(ctx, project)
}

func (r *mutationResolver) CoreUpdateProject(ctx context.Context, project entities.Project) (*entities.Project, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreDeleteProject(ctx context.Context, name string) (bool, error) {
	return r.Domain.DeleteProject(ctx, name)
}

func (r *mutationResolver) CoreCreateApp(ctx context.Context, app entities.App) (*entities.App, error) {
	return r.Domain.CreateApp(ctx, app)
}

func (r *mutationResolver) CoreUpdateApp(ctx context.Context, app entities.App) (*entities.App, error) {
	return r.Domain.UpdateApp(ctx, app)
}

func (r *mutationResolver) CoreDeleteApp(ctx context.Context, namespace string, name string) (bool, error) {
	return r.Domain.DeleteApp(ctx, namespace, name)
}

func (r *mutationResolver) CoreCreateSecret(ctx context.Context, secret entities.Secret) (*entities.Secret, error) {
	return r.Domain.CreateSecret(ctx, secret)
}

func (r *mutationResolver) CoreUpdateSecret(ctx context.Context, secret entities.Secret) (*entities.Secret, error) {
	return r.Domain.UpdateSecret(ctx, secret)
}

func (r *mutationResolver) CoreDeleteSecret(ctx context.Context, namespace string, name string) (bool, error) {
	return r.Domain.DeleteSecret(ctx, namespace, name)
}

func (r *mutationResolver) CoreCreateConfig(ctx context.Context, config entities.Config) (*entities.Config, error) {
	return r.Domain.CreateConfig(ctx, config)
}

func (r *mutationResolver) CoreUpdateConfig(ctx context.Context, config entities.Config) (*entities.Config, error) {
	return r.Domain.UpdateConfig(ctx, config)
}

func (r *mutationResolver) CoreDeleteConfig(ctx context.Context, namespace string, name string) (bool, error) {
	return r.Domain.DeleteConfig(ctx, namespace, name)
}

func (r *mutationResolver) CoreCreateRouter(ctx context.Context, router entities.Router) (*entities.Router, error) {
	return r.Domain.CreateRouter(ctx, router)
}

func (r *mutationResolver) CoreUpdateRouter(ctx context.Context, router entities.Router) (*entities.Router, error) {
	return r.Domain.UpdateRouter(ctx, router)
}

func (r *mutationResolver) CoreDeleteRouter(ctx context.Context, namespace string, name string) (bool, error) {
	return r.Domain.DeleteRouter(ctx, namespace, name)
}

func (r *mutationResolver) CoreInstallManagedSvc(ctx context.Context, msvc entities.ManagedService) (*entities.ManagedService, error) {
	return r.Domain.InstallManagedSvc(ctx, msvc)
}

func (r *mutationResolver) CoreUpdateManagedSvc(ctx context.Context, msvc entities.ManagedService) (*entities.ManagedService, error) {
	return r.Domain.UpdateManagedSvc(ctx, msvc)
}

func (r *mutationResolver) CoreUninstallManagedSvc(ctx context.Context, namespace string, name string) (bool, error) {
	return r.Domain.UnInstallManagedSvc(ctx, namespace, name)
}

func (r *mutationResolver) CoreCreateManagedRes(ctx context.Context, mres entities.ManagedResource) (*entities.ManagedResource, error) {
	return r.Domain.CreateManagedRes(ctx, mres)
}

func (r *mutationResolver) CoreUpdateManagedRes(ctx context.Context, mres entities.ManagedResource) (*entities.ManagedResource, error) {
	return r.Domain.UpdateManagedRes(ctx, mres)
}

func (r *mutationResolver) CoreDeleteManagedRes(ctx context.Context, namespace string, name string) (bool, error) {
	return r.Domain.DeleteManagedRes(ctx, namespace, name)
}

func (r *queryResolver) CoreListCloudProviders(ctx context.Context, accountID string) ([]*entities.CloudProvider, error) {
	return r.Domain.ListCloudProviders(ctx, repos.ID(accountID))
}

func (r *queryResolver) CoreGetCloudProvider(ctx context.Context, name string) (*entities.CloudProvider, error) {
	return r.Domain.GetCloudProvider(ctx, name)
}

func (r *queryResolver) CoreSample(ctx context.Context) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CoreGetEnvironments(ctx context.Context, projectName string) ([]*entities.Environment, error) {
	return r.Domain.GetEnvironments(ctx, projectName)
}

func (r *queryResolver) CoreGetEnvironment(ctx context.Context, namespace string) (*entities.Environment, error) {
	return r.Domain.GetEnvironment(ctx, namespace)
}

func (r *queryResolver) CoreProjects(ctx context.Context, accountID repos.ID) ([]*entities.Project, error) {
	return r.Domain.GetAccountProjects(ctx, accountID)
}

func (r *queryResolver) CoreProject(ctx context.Context, name string) (*entities.Project, error) {
	return r.Domain.GetProjectWithName(ctx, name)
}

func (r *queryResolver) CoreApps(ctx context.Context, namespace string, search *string) ([]*entities.App, error) {
	return r.Domain.GetApps(ctx, namespace, search)
}

func (r *queryResolver) CoreApp(ctx context.Context, namespace string, name string) (*entities.App, error) {
	return r.Domain.GetApp(ctx, namespace, name)
}

func (r *queryResolver) CoreRouters(ctx context.Context, namespace string, search *string) ([]*entities.Router, error) {
	return r.Domain.GetRouters(ctx, namespace, search)
}

func (r *queryResolver) CoreRouter(ctx context.Context, namespace string, name string) (*entities.Router, error) {
	return r.Domain.GetRouter(ctx, namespace, name)
}

func (r *queryResolver) CoreConfigs(ctx context.Context, namespace string, search *string) ([]*entities.Config, error) {
	return r.Domain.GetConfigs(ctx, namespace, search)
}

func (r *queryResolver) CoreConfig(ctx context.Context, namespace string, name string) (*entities.Config, error) {
	return r.Domain.GetConfig(ctx, namespace, name)
}

func (r *queryResolver) CoreSecrets(ctx context.Context, namespace string, search *string) ([]*entities.Secret, error) {
	return r.Domain.GetSecrets(ctx, namespace, search)
}

func (r *queryResolver) CoreSecret(ctx context.Context, namespace string, name string) (*entities.Secret, error) {
	return r.Domain.GetSecret(ctx, namespace, name)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
