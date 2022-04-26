package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/app/graph/model"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/common"
	"kloudlite.io/pkg/cache"
	wErrors "kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

func (r *accountResolver) Projects(ctx context.Context, obj *model.Account) ([]*model.Project, error) {
	projectEntities, err := r.Domain.GetAccountProjects(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	projectModels := make([]*model.Project, 0)
	for _, pe := range projectEntities {
		projectModels = append(projectModels, projectModelFromEntity(pe))
	}
	return projectModels, err
}

func (r *accountResolver) Clusters(ctx context.Context, obj *model.Account) ([]*model.Cluster, error) {
	clusterEntities, err := r.Domain.ListClusters(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	clusters := make([]*model.Cluster, 0)
	for _, cle := range clusterEntities {
		clusters = append(clusters, &model.Cluster{
			ID:         cle.Id,
			Name:       cle.Name,
			Provider:   cle.Provider,
			Region:     cle.Region,
			IP:         cle.Ip,
			NodesCount: cle.NodesCount,
			Status:     string(cle.Status),
			Account: &model.Account{
				ID: repos.ID(cle.AccountId),
			},
		})
	}
	return clusters, err
}

func (r *clusterResolver) Devices(ctx context.Context, obj *model.Cluster) ([]*model.Device, error) {
	var e error
	defer wErrors.HandleErr(&e)
	cluster := obj
	deviceEntities, e := r.Domain.ListClusterDevices(ctx, cluster.ID)
	wErrors.AssertNoError(e, fmt.Errorf("not able to list devices of cluster %s", cluster.ID))
	devices := make([]*model.Device, len(deviceEntities))
	for i, d := range deviceEntities {
		devices[i] = &model.Device{
			ID:      d.Id,
			User:    &model.User{ID: d.UserId},
			Name:    d.Name,
			Cluster: cluster,
			IP:      d.Ip,
		}
	}
	return devices, e
}

func (r *clusterResolver) UserDevices(ctx context.Context, obj *model.Cluster) ([]*model.Device, error) {
	var e error
	defer wErrors.HandleErr(&e)
	user := obj
	deviceEntities, e := r.Domain.ListUserDevices(ctx, repos.ID(user.ID), &obj.ID)
	wErrors.AssertNoError(e, fmt.Errorf("not able to list devices of user %s", user.ID))
	devices := make([]*model.Device, 0)
	for _, device := range deviceEntities {
		devices = append(devices, &model.Device{
			ID:      device.Id,
			User:    &model.User{ID: user.ID},
			Name:    device.Name,
			Cluster: &model.Cluster{ID: device.ClusterId},
			IP:      device.Ip,
		})
	}
	return devices, e
}

func (r *deviceResolver) User(ctx context.Context, obj *model.Device) (*model.User, error) {
	var e error
	defer wErrors.HandleErr(&e)
	device := obj
	deviceEntity, e := r.Domain.GetDevice(ctx, device.ID)
	wErrors.AssertNoError(e, fmt.Errorf("not able to get device"))
	return &model.User{
		ID: deviceEntity.UserId,
	}, e
}

func (r *deviceResolver) Cluster(ctx context.Context, obj *model.Device) (*model.Cluster, error) {
	deviceEntity, err := r.Domain.GetDevice(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	clusterEntity, err := r.Domain.GetCluster(ctx, deviceEntity.ClusterId)
	if err != nil {
		return nil, err
	}
	return &model.Cluster{
		ID:         clusterEntity.Id,
		Name:       clusterEntity.Name,
		Provider:   clusterEntity.Provider,
		Region:     clusterEntity.Region,
		IP:         clusterEntity.Ip,
		NodesCount: clusterEntity.NodesCount,
		Status:     string(clusterEntity.Status),
	}, nil
}

func (r *deviceResolver) Configuration(ctx context.Context, obj *model.Device) (string, error) {
	return r.Domain.GetDeviceConfig(ctx, obj.ID)
}

func (r *managedResResolver) Outputs(ctx context.Context, obj *model.ManagedRes) (map[string]interface{}, error) {
	return r.Domain.GetResourceOutputs(ctx, obj.ID)
}

func (r *managedSvcResolver) Resources(ctx context.Context, obj *model.ManagedSvc) ([]*model.ManagedRes, error) {
	resources, err := r.Domain.GetManagedResourcesOfService(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	res := make([]*model.ManagedRes, 0)
	for _, r := range resources {
		res = append(res, managedResourceModelFromEntity(r))
	}
	return res, nil
}

func (r *mutationResolver) MangedSvcInstall(ctx context.Context, projectID repos.ID, serviceType repos.ID, name string, values map[string]interface{}) (*model.ManagedSvc, error) {
	svcEntity, err := r.Domain.InstallManagedSvc(ctx, projectID, serviceType, name, values)
	if err != nil {
		return nil, err
	}
	return &model.ManagedSvc{
		ID:      svcEntity.Id,
		Name:    svcEntity.Name,
		Project: &model.Project{ID: projectID},
		Source:  string(svcEntity.ServiceType),
		Values:  values,
	}, nil
}

func (r *mutationResolver) MangedSvcUninstall(ctx context.Context, installationID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) MangedSvcUpdate(ctx context.Context, installationID repos.ID, values map[string]interface{}) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ManagedResCreate(ctx context.Context, installationID repos.ID, name string, resourceType string, values map[string]interface{}) (*model.ManagedRes, error) {
	kvs := make(map[string]string, 0)
	for k, v := range values {
		kvs[k] = v.(string)
	}
	res, err := r.Domain.InstallManagedRes(ctx, installationID, name, resourceType, kvs)
	if err != nil {
		return nil, err
	}
	return &model.ManagedRes{
		ID:           res.Id,
		Name:         res.Name,
		ResourceType: string(res.ResourceType),
		Installation: &model.ManagedSvc{
			ID: installationID,
		},
		Values: values,
	}, nil
}

func (r *mutationResolver) ManagedResUpdate(ctx context.Context, resID repos.ID, values map[string]interface{}) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ManagedResDelete(ctx context.Context, resID repos.ID) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) InfraCreateCluster(ctx context.Context, name string, provider string, region string, nodesCount int) (*model.Cluster, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) InfraUpdateCluster(ctx context.Context, name *string, clusterID repos.ID, nodesCount *int) (*model.Cluster, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) InfraDeleteCluster(ctx context.Context, clusterID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) InfraAddDevice(ctx context.Context, clusterID repos.ID, name string) (*model.Device, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return nil, errors.New("user not logged in")
	}
	device, err := r.Domain.AddDevice(ctx, name, clusterID, session.UserId)
	if err != nil {
		return nil, err
	}
	return &model.Device{
		ID:   device.Id,
		User: &model.User{ID: session.UserId},
		Name: device.Name,
		Cluster: &model.Cluster{
			ID: clusterID,
		},
		IP: device.Ip,
	}, nil
}

func (r *mutationResolver) InfraRemoveDevice(ctx context.Context, deviceID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreCreateProject(ctx context.Context, accountID repos.ID, name string, displayName string, logo *string, description *string) (*model.Project, error) {
	projectEntity, err := r.Domain.CreateProject(ctx, accountID, name, displayName, logo, description)
	if err != nil {
		return nil, err
	}
	return projectModelFromEntity(projectEntity), nil
}

func (r *mutationResolver) CoreUpdateProject(ctx context.Context, projectID repos.ID, displayName *string, cluster *string, logo *string, description *string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreDeleteProject(ctx context.Context, projectID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) IamInviteProjectMember(ctx context.Context, projectID repos.ID, email string, name string, role string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) IamRemoveProjectMember(ctx context.Context, projectID repos.ID, userID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) IamUpdateProjectMember(ctx context.Context, projectID repos.ID, userID repos.ID, role string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreCreateAppFlow(ctx context.Context, projectID repos.ID, app model.AppFlowInput) (bool, error) {
	session := cache.GetSession[*common.AuthSession](ctx)
	if session == nil {
		return false, errors.New("user not logged in")
	}
	ports := make([]entities.ExposedPort, 0)
	for _, port := range app.ExposedServices {
		ports = append(ports, entities.ExposedPort{
			Port:       int64(port.Exposed),
			TargetPort: int64(port.Target),
			Type:       entities.PortType(port.Type),
		})
	}
	containers := make([]entities.ContainerIn, 0)
	for _, container := range app.Containers {
		e := make([]entities.EnvVar, 0)
		for _, env := range container.EnvVars {
			e = append(e, entities.EnvVar{
				Key:    env.Key,
				Type:   env.Value.Type,
				Value:  env.Value.Value,
				Ref:    env.Value.Ref,
				RefKey: env.Value.Key,
			})
		}
		a := make([]entities.AttachedResource, 0)
		for _, attached := range container.AttachedResources {
			a = append(a, entities.AttachedResource{
				ResourceId: attached.ResID,
			})
		}
		i := *container.PipelineData.GithubInstallationID
		containers = append(containers, entities.ContainerIn{
			Pipeline: &entities.PipelineIn{
				Name:                 container.PipelineData.Name,
				ImageName:            container.PipelineData.ImageName,
				GitProvider:          container.PipelineData.GitProvider,
				GitRepoUrl:           container.PipelineData.GitRepoURL,
				DockerFile:           container.PipelineData.DockerFile,
				ContextDir:           container.PipelineData.ContextDir,
				GithubInstallationId: int64(i),
				BuildArgs:            container.PipelineData.BuildArgs,
			},
			Name:            container.Name,
			Image:           container.Image,
			ImagePullSecret: container.PullSecret,
			EnvVars:         e,
			CPULimits: entities.Limit{
				Min: container.CPUMin,
				Max: container.CPUMax,
			},
			MemoryLimits: entities.Limit{
				Min: container.MemMin,
				Max: container.MemMax,
			},
			AttachedResources: a,
		})
	}
	return r.Domain.InstallAppFlow(ctx, session.UserId, projectID, entities.AppIn{
		Name:         app.Name,
		ReadableId:   app.Readable,
		Description:  app.Description,
		Replicas:     1,
		ExposedPorts: ports,
		Containers:   containers,
	})
}

func (r *mutationResolver) CoreDeleteApp(ctx context.Context, appID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreRollbackApp(ctx context.Context, appID repos.ID, version int) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreCreateSecret(ctx context.Context, projectID repos.ID, name string, description *string, data []*model.CSEntryIn) (*model.Secret, error) {
	entries := make([]*entities.Entry, 0)
	for _, i := range data {
		entries = append(entries, &entities.Entry{
			Key:   i.Key,
			Value: i.Value,
		})
	}
	configEntity, err := r.Domain.CreateSecret(ctx, projectID, name, description, entries)
	if err != nil {
		return nil, err
	}
	return secretModelFromEntity(configEntity), nil
}

func (r *mutationResolver) CoreUpdateSecret(ctx context.Context, secretID repos.ID, description *string, data []*model.CSEntryIn) (bool, error) {
	entries := make([]*entities.Entry, 0)
	for _, i := range data {
		entries = append(entries, &entities.Entry{
			Key:   i.Key,
			Value: i.Value,
		})
	}
	return r.Domain.UpdateSecret(ctx, secretID, description, entries)
}

func (r *mutationResolver) CoreDeleteSecret(ctx context.Context, secretID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreCreateConfig(ctx context.Context, projectID repos.ID, name string, description *string, data []*model.CSEntryIn) (*model.Config, error) {
	entries := make([]*entities.Entry, 0)
	for _, i := range data {
		entries = append(entries, &entities.Entry{
			Key:   i.Key,
			Value: i.Value,
		})
	}
	configEntity, err := r.Domain.CreateConfig(ctx, projectID, name, description, entries)
	if err != nil {
		return nil, err
	}
	return configModelFromEntity(configEntity), nil
}

func (r *mutationResolver) CoreUpdateConfig(ctx context.Context, configID repos.ID, description *string, data []*model.CSEntryIn) (bool, error) {
	entries := make([]*entities.Entry, 0)
	for _, i := range data {
		entries = append(entries, &entities.Entry{
			Key:   i.Key,
			Value: i.Value,
		})
	}
	return r.Domain.UpdateConfig(ctx, configID, description, entries)
}

func (r *mutationResolver) CoreDeleteConfig(ctx context.Context, configID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreCreateRouter(ctx context.Context, projectID repos.ID, name string, domains []string, routes []*model.RouteInput) (*model.Router, error) {
	routeEnt := make([]*entities.Route, 0)
	for _, r := range routes {
		routeEnt = append(routeEnt, &entities.Route{
			Path:    r.Path,
			AppName: r.AppName,
			Port:    uint16(r.Port),
		})
	}
	d := domains
	if domains == nil {
		d = []string{}
	}
	router, err := r.Domain.CreateRouter(ctx, projectID, name, d, routeEnt)
	if err != nil {
		return nil, err
	}
	return routerModelFromEntity(router), err
}

func (r *mutationResolver) CoreUpdateRouter(ctx context.Context, routerID repos.ID, domains []string, routes []*model.RouteInput) (bool, error) {
	entries := make([]*entities.Route, 0)
	for _, i := range routes {
		entries = append(entries, &entities.Route{
			Path:    i.Path,
			AppName: i.AppName,
			Port:    uint16(i.Port),
		})
	}
	return r.Domain.UpdateRouter(ctx, routerID, domains, entries)
}

func (r *mutationResolver) CoreDeleteRouter(ctx context.Context, routerID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CoreProjects(ctx context.Context, accountID *repos.ID) ([]*model.Project, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CoreProject(ctx context.Context, projectID repos.ID) (*model.Project, error) {
	p, err := r.Domain.GetProjectWithID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return projectModelFromEntity(p), nil
}

func (r *queryResolver) CoreApps(ctx context.Context, projectID repos.ID, search *string) ([]*model.App, error) {
	appEntities, err := r.Domain.GetApps(ctx, projectID)
	if err != nil {
		return nil, err
	}
	apps := make([]*model.App, 0)

	for _, a := range appEntities {
		services := make([]*model.ExposedService, 0)
		for _, s := range a.ExposedPorts {
			services = append(services, &model.ExposedService{
				Type:    string(s.Type),
				Target:  int(s.TargetPort),
				Exposed: int(s.Port),
			})
		}

		containers := make([]*model.AppContainer, 0)
		for _, c := range a.Containers {
			envVars := make([]*model.EnvVar, 0)
			for _, e := range c.EnvVars {
				envVars = append(envVars, &model.EnvVar{
					Key: e.Key,
					Value: &model.EnvVal{
						Type:  e.Type,
						Value: e.Value,
						Ref:   e.Ref,
						Key:   e.RefKey,
					},
				})
			}
			res := make([]*model.AttachedRes, 0)
			for _, r := range c.AttachedResources {
				res = append(res, &model.AttachedRes{
					ResID: r.ResourceId,
				})
			}
			containers = append(containers, &model.AppContainer{
				Name:              c.Name,
				Image:             c.Image,
				PullSecret:        c.ImagePullSecret,
				EnvVars:           envVars,
				CPUMin:            c.CPULimits.Min,
				CPUMax:            c.CPULimits.Max,
				MemMin:            c.MemoryLimits.Min,
				MemMax:            c.MemoryLimits.Max,
				AttachedResources: res,
			})
		}

		apps = append(apps, &model.App{
			ID:          a.Id,
			Name:        a.Name,
			Namespace:   a.Namespace,
			Description: a.Description,
			ReadableID:  repos.ID(a.ReadableId),
			Replicas:    &a.Replicas,
			Services:    services,
			Containers:  containers,
			Project:     &model.Project{ID: projectID},
		})
	}
	return apps, nil
}

func (r *queryResolver) CoreApp(ctx context.Context, appID repos.ID) (*model.App, error) {
	a, err := r.Domain.GetApp(ctx, appID)
	if err != nil {
		return nil, err
	}
	services := make([]*model.ExposedService, 0)
	for _, s := range a.ExposedPorts {
		services = append(services, &model.ExposedService{
			Type:    string(s.Type),
			Target:  int(s.TargetPort),
			Exposed: int(s.Port),
		})
	}

	containers := make([]*model.AppContainer, 0)
	for _, c := range a.Containers {
		envVars := make([]*model.EnvVar, 0)
		for _, e := range c.EnvVars {
			envVars = append(envVars, &model.EnvVar{
				Key: e.Key,
				Value: &model.EnvVal{
					Type:  e.Type,
					Value: e.Value,
					Ref:   e.Ref,
					Key:   e.RefKey,
				},
			})
		}
		res := make([]*model.AttachedRes, 0)
		for _, r := range c.AttachedResources {
			res = append(res, &model.AttachedRes{
				ResID: r.ResourceId,
			})
		}
		containers = append(containers, &model.AppContainer{
			Name:              c.Name,
			Image:             c.Image,
			PullSecret:        c.ImagePullSecret,
			EnvVars:           envVars,
			CPUMin:            c.CPULimits.Min,
			CPUMax:            c.CPULimits.Max,
			MemMin:            c.MemoryLimits.Min,
			MemMax:            c.MemoryLimits.Max,
			AttachedResources: res,
		})
	}

	return &model.App{
		ID:          a.Id,
		Name:        a.Name,
		Namespace:   a.Namespace,
		Description: a.Description,
		ReadableID:  repos.ID(a.ReadableId),
		Replicas:    &a.Replicas,
		Services:    services,
		Containers:  containers,
		Project:     &model.Project{ID: a.ProjectId},
	}, nil
}

func (r *queryResolver) CoreRouters(ctx context.Context, projectID repos.ID, search *string) ([]*model.Router, error) {
	routerEntities, err := r.Domain.GetRouters(ctx, projectID)
	if err != nil {
		return nil, err
	}
	routers := make([]*model.Router, 0)
	for _, i := range routerEntities {
		routers = append(routers, routerModelFromEntity(i))
	}
	return routers, nil
}

func (r *queryResolver) CoreRouter(ctx context.Context, routerID repos.ID) (*model.Router, error) {
	routerEntity, err := r.Domain.GetRouter(ctx, routerID)
	if err != nil {
		return nil, err
	}
	return routerModelFromEntity(routerEntity), nil
}

func (r *queryResolver) CoreConfigs(ctx context.Context, projectID repos.ID, search *string) ([]*model.Config, error) {
	configEntities, err := r.Domain.GetConfigs(ctx, projectID)
	if err != nil {
		return nil, err
	}
	configs := make([]*model.Config, 0)
	for _, i := range configEntities {
		configs = append(configs, configModelFromEntity(i))
	}
	return configs, nil
}

func (r *queryResolver) CoreConfig(ctx context.Context, configID repos.ID) (*model.Config, error) {
	configEntity, err := r.Domain.GetConfig(ctx, configID)
	if err != nil {
		return nil, err
	}
	return configModelFromEntity(configEntity), nil
}

func (r *queryResolver) CoreSecrets(ctx context.Context, projectID repos.ID, search *string) ([]*model.Secret, error) {
	configEntities, err := r.Domain.GetSecrets(ctx, projectID)
	if err != nil {
		return nil, err
	}
	secrets := make([]*model.Secret, 0)
	for _, i := range configEntities {
		secrets = append(secrets, secretModelFromEntity(i))
	}
	return secrets, nil
}

func (r *queryResolver) CoreSecret(ctx context.Context, secretID repos.ID) (*model.Secret, error) {
	secretEntity, err := r.Domain.GetSecret(ctx, secretID)
	if err != nil {
		return nil, err
	}
	return secretModelFromEntity(secretEntity), nil
}

func (r *queryResolver) ManagedSvcMarketList(ctx context.Context) (map[string]interface{}, error) {
	templates, err := r.Domain.GetManagedServiceTemplates(ctx)
	if err != nil {
		return nil, err
	}
	var res []any
	marshal, err := json.Marshal(templates)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(marshal, &res)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"categories": res}, nil
}

func (r *queryResolver) ManagedSvcListAvailable(ctx context.Context) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ManagedSvcGetInstallation(ctx context.Context, installationID repos.ID, nextVersion *bool) (*model.ManagedSvc, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ManagedSvcListInstallations(ctx context.Context, projectID repos.ID) ([]*model.ManagedSvc, error) {
	svcs, err := r.Domain.GetManagedSvcs(ctx, projectID)
	managedSvcs := make([]*model.ManagedSvc, 0)
	for _, svc := range svcs {
		managedSvcs = append(managedSvcs, managedSvcModelFromEntity(svc))
	}
	return managedSvcs, err
}

func (r *queryResolver) ManagedResGetResource(ctx context.Context, resID repos.ID, nextVersion *bool) (*model.ManagedRes, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ManagedResListResources(ctx context.Context, installationID repos.ID) ([]*model.ManagedRes, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) InfraGetCluster(ctx context.Context, clusterID repos.ID) (*model.Cluster, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *userResolver) Devices(ctx context.Context, obj *model.User) ([]*model.Device, error) {
	var e error
	defer wErrors.HandleErr(&e)
	user := obj
	deviceEntities, e := r.Domain.ListUserDevices(ctx, repos.ID(user.ID), nil)
	wErrors.AssertNoError(e, fmt.Errorf("not able to list devices of user %s", user.ID))
	devices := make([]*model.Device, 0)
	for _, device := range deviceEntities {
		devices = append(devices, &model.Device{
			ID:      device.Id,
			User:    &model.User{ID: user.ID},
			Name:    device.Name,
			Cluster: &model.Cluster{ID: device.ClusterId},
			IP:      device.Ip,
		})
	}
	return devices, e
}

// Account returns generated.AccountResolver implementation.
func (r *Resolver) Account() generated.AccountResolver { return &accountResolver{r} }

// Cluster returns generated.ClusterResolver implementation.
func (r *Resolver) Cluster() generated.ClusterResolver { return &clusterResolver{r} }

// Device returns generated.DeviceResolver implementation.
func (r *Resolver) Device() generated.DeviceResolver { return &deviceResolver{r} }

// ManagedRes returns generated.ManagedResResolver implementation.
func (r *Resolver) ManagedRes() generated.ManagedResResolver { return &managedResResolver{r} }

// ManagedSvc returns generated.ManagedSvcResolver implementation.
func (r *Resolver) ManagedSvc() generated.ManagedSvcResolver { return &managedSvcResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type accountResolver struct{ *Resolver }
type clusterResolver struct{ *Resolver }
type deviceResolver struct{ *Resolver }
type managedResResolver struct{ *Resolver }
type managedSvcResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
