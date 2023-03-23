package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/domain/entities"
)

func (r *mutationResolver) CoreCreateProject(ctx context.Context, project entities.Project) (*entities.Project, error) {
	return r.Domain.CreateProject(toConsoleContext(ctx), project)
}

func (r *mutationResolver) CoreUpdateProject(ctx context.Context, project entities.Project) (*entities.Project, error) {
	return r.Domain.UpdateProject(toConsoleContext(ctx), project)
}

func (r *mutationResolver) CoreDeleteProject(ctx context.Context, name string) (bool, error) {
	if err := r.Domain.DeleteProject(toConsoleContext(ctx), name); err != nil {
		return false, nil
	}
	return true, nil
}

func (r *mutationResolver) CoreCreateApp(ctx context.Context, app entities.App) (*entities.App, error) {
	return r.Domain.CreateApp(toConsoleContext(ctx), app)
}

func (r *mutationResolver) CoreUpdateApp(ctx context.Context, app entities.App) (*entities.App, error) {
	return r.Domain.UpdateApp(toConsoleContext(ctx), app)
}

func (r *mutationResolver) CoreDeleteApp(ctx context.Context, namespace string, name string) (bool, error) {
	if err := r.Domain.DeleteApp(toConsoleContext(ctx), namespace, name); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreCreateConfig(ctx context.Context, config entities.Config) (*entities.Config, error) {
	return r.Domain.CreateConfig(toConsoleContext(ctx), config)
}

func (r *mutationResolver) CoreUpdateConfig(ctx context.Context, config entities.Config) (*entities.Config, error) {
	return r.Domain.UpdateConfig(toConsoleContext(ctx), config)
}

func (r *mutationResolver) CoreDeleteConfig(ctx context.Context, namespace string, name string) (bool, error) {
	if err := r.Domain.DeleteConfig(toConsoleContext(ctx), namespace, name); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreCreateSecret(ctx context.Context, secret entities.Secret) (*entities.Secret, error) {
	return r.Domain.CreateSecret(toConsoleContext(ctx), secret)
}

func (r *mutationResolver) CoreUpdateSecret(ctx context.Context, secret entities.Secret) (*entities.Secret, error) {
	return r.Domain.UpdateSecret(toConsoleContext(ctx), secret)
}

func (r *mutationResolver) CoreDeleteSecret(ctx context.Context, namespace string, name string) (bool, error) {
	if err := r.Domain.DeleteSecret(toConsoleContext(ctx), namespace, name); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreCreateRouter(ctx context.Context, router entities.Router) (*entities.Router, error) {
	return r.Domain.CreateRouter(toConsoleContext(ctx), router)
}

func (r *mutationResolver) CoreUpdateRouter(ctx context.Context, router entities.Router) (*entities.Router, error) {
	return r.Domain.UpdateRouter(toConsoleContext(ctx), router)
}

func (r *mutationResolver) CoreDeleteRouter(ctx context.Context, namespace string, name string) (bool, error) {
	if err := r.Domain.DeleteRouter(toConsoleContext(ctx), namespace, name); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreCreateManagedService(ctx context.Context, msvc entities.MSvc) (*entities.MSvc, error) {
	return r.Domain.CreateManagedService(toConsoleContext(ctx), msvc)
}

func (r *mutationResolver) CoreUpdateManagedService(ctx context.Context, msvc entities.MSvc) (*entities.MSvc, error) {
	return r.Domain.UpdateManagedService(toConsoleContext(ctx), msvc)
}

func (r *mutationResolver) CoreDeleteManagedService(ctx context.Context, namespace string, name string) (bool, error) {
	if err := r.Domain.DeleteManagedService(toConsoleContext(ctx), namespace, name); err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CoreCreateManagedResource(ctx context.Context, mres entities.MRes) (*entities.MRes, error) {
	return r.Domain.CreateManagedResource(toConsoleContext(ctx), mres)
}

func (r *mutationResolver) CoreUpdateManagedResource(ctx context.Context, mres entities.MRes) (*entities.MRes, error) {
	return r.Domain.UpdateManagedResource(toConsoleContext(ctx), mres)
}

func (r *mutationResolver) CoreDeleteManagedResource(ctx context.Context, namespace string, name string) (bool, error) {
	if err := r.Domain.DeleteManagedResource(toConsoleContext(ctx), namespace, name); err != nil {
		return false, err
	}
	return true, nil
}

func (r *queryResolver) CoreListProjects(ctx context.Context) ([]*entities.Project, error) {
	p, err := r.Domain.ListProjects(toConsoleContext(ctx))
	if err != nil {
		return nil, err
	}
	if p == nil {
		p = make([]*entities.Project, 0)
	}
	return p, nil
}

func (r *queryResolver) CoreGetProject(ctx context.Context, name string) (*entities.Project, error) {
	return r.Domain.GetProject(toConsoleContext(ctx), name)
}

func (r *queryResolver) CoreListApps(ctx context.Context, namespace string) ([]*entities.App, error) {
	a, err := r.Domain.ListApps(toConsoleContext(ctx), namespace)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return make([]*entities.App, 0), nil
	}
	return a, nil
}

func (r *queryResolver) CoreGetApp(ctx context.Context, namespace string, name string) (*entities.App, error) {
	return r.Domain.GetApp(toConsoleContext(ctx), namespace, name)
}

func (r *queryResolver) CoreListConfigs(ctx context.Context, namespace string) ([]*entities.Config, error) {
	c, err := r.Domain.ListConfigs(toConsoleContext(ctx), namespace)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return make([]*entities.Config, 0), nil
	}
	return c, nil
}

func (r *queryResolver) CoreGetConfig(ctx context.Context, namespace string, name string) (*entities.Config, error) {
	return r.Domain.GetConfig(toConsoleContext(ctx), namespace, name)
}

func (r *queryResolver) CoreListSecrets(ctx context.Context, namespace string) ([]*entities.Secret, error) {
	s, err := r.Domain.ListSecrets(toConsoleContext(ctx), namespace)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return make([]*entities.Secret, 0), nil
	}
	return s, nil
}

func (r *queryResolver) CoreGetSecret(ctx context.Context, namespace string, name string) (*entities.Secret, error) {
	return r.Domain.GetSecret(toConsoleContext(ctx), namespace, name)
}

func (r *queryResolver) CoreListRouters(ctx context.Context, namespace string) ([]*entities.Router, error) {
	routers, err := r.Domain.ListRouters(toConsoleContext(ctx), namespace)
	if err != nil {
		return nil, err
	}
	if routers == nil {
		return make([]*entities.Router, 0), nil
	}
	return routers, nil
}

func (r *queryResolver) CoreGetRouter(ctx context.Context, namespace string, name string) (*entities.Router, error) {
	return r.Domain.GetRouter(toConsoleContext(ctx), namespace, name)
}

func (r *queryResolver) CoreListManagedServices(ctx context.Context, namespace string) ([]*entities.MSvc, error) {
	m, err := r.Domain.ListManagedServices(toConsoleContext(ctx), namespace)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return make([]*entities.MSvc, 0), nil
	}
	return m, nil
}

func (r *queryResolver) CoreGetManagedService(ctx context.Context, namespace string, name string) (*entities.MSvc, error) {
	return r.Domain.GetManagedService(toConsoleContext(ctx), namespace, name)
}

func (r *queryResolver) CoreListManagedResources(ctx context.Context, namespace string) ([]*entities.MRes, error) {
	m, err := r.Domain.ListManagedResources(toConsoleContext(ctx), namespace)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return make([]*entities.MRes, 0), nil
	}
	return m, nil
}

func (r *queryResolver) CoreGetManagedResource(ctx context.Context, namespace string, name string) (*entities.MRes, error) {
	return r.Domain.GetManagedResource(toConsoleContext(ctx), namespace, name)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
