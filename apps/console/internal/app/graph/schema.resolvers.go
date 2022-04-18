package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/app/graph/model"
	"kloudlite.io/apps/console/internal/domain/entities"
	wErrors "kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

func (r *clusterResolver) Devices(ctx context.Context, obj *model.Cluster) ([]*model.Device, error) {
	var e error
	defer wErrors.HandleErr(&e)
	cluster := obj
	deviceEntities, e := r.Domain.ListClusterDevices(ctx, cluster.ID)
	wErrors.AssertNoError(e, fmt.Errorf("not able to list devices of cluster %s", cluster.ID))
	devices := make([]*model.Device, len(deviceEntities))
	for i, d := range deviceEntities {
		devices[i] = &model.Device{
			ID:            d.Id,
			Name:          d.Name,
			Cluster:       cluster,
			Configuration: "",
		}
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
	//deviceEntity, err := r.Domain.GetDevice(ctx, obj.ID)
	//	return fmt.Sprintf(`
	//[Interface]
	//PrivateKey = %v
	//Address = %v/32
	//DNS = 10.43.0.10
	//
	//[Peer]
	//PublicKey = %v
	//AllowedIPs = 10.42.0.0/16, 10.43.0.0/16, 10.13.13.0/16
	//Endpoint = %v:31820
	//`, deviceEntity.PrivateKey, deviceEntity.), err
	return "nil", nil
}

func (r *mutationResolver) MangedSvcInstall(ctx context.Context, projectID repos.ID, templateID repos.ID, name string, values string) (*model.ManagedSvc, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) MangedSvcUninstall(ctx context.Context, installationID repos.ID) (*model.ManagedSvc, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) MangedSvcUpdate(ctx context.Context, installationID repos.ID, values string) (*model.ManagedSvc, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ManagedResCreate(ctx context.Context, installationID repos.ID, name string, resourceName string, values string) (*model.ManagedRes, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ManagedResUpdate(ctx context.Context, resID repos.ID, values *string) (*model.ManagedRes, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ManagedResDelete(ctx context.Context, resID repos.ID) (*model.ManagedRes, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) JobCommit(ctx context.Context, jobID repos.ID) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) JobUndo(ctx context.Context, jobID repos.ID) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateCluster(ctx context.Context, name string, provider string, region string, nodesCount int) (*model.Cluster, error) {
	cluster, e := r.Domain.CreateCluster(ctx, entities.Cluster{
		Name:       name,
		Provider:   provider,
		Region:     region,
		NodesCount: nodesCount,
	})
	if e != nil {
		return nil, e
	}
	return &model.Cluster{
		ID:         cluster.Id,
		Name:       cluster.Name,
		Provider:   cluster.Provider,
		Region:     cluster.Region,
		NodesCount: cluster.NodesCount,
	}, e
}

func (r *mutationResolver) UpdateCluster(ctx context.Context, name *string, clusterID repos.ID, nodesCount *int) (*model.Cluster, error) {
	clusterEntity, err := r.Domain.UpdateCluster(ctx, clusterID, name, nodesCount)
	return &model.Cluster{
		ID:         clusterEntity.Id,
		Name:       clusterEntity.Name,
		Provider:   clusterEntity.Provider,
		Region:     clusterEntity.Region,
		IP:         clusterEntity.Ip,
		NodesCount: clusterEntity.NodesCount,
	}, err
}

func (r *mutationResolver) DeleteCluster(ctx context.Context, clusterID repos.ID) (bool, error) {
	err := r.Domain.DeleteCluster(ctx, clusterID)
	if err != nil {
		return false, err
	}
	return true, err
}

func (r *mutationResolver) AddDevice(ctx context.Context, clusterID repos.ID, userID repos.ID, name string) (*model.Device, error) {
	device, e := r.Domain.AddDevice(ctx, name, clusterID, userID)
	if e != nil {
		return nil, e
	}
	return &model.Device{
		ID:            device.Id,
		Name:          device.Name,
		Configuration: "",
	}, e
}

func (r *mutationResolver) RemoveDevice(ctx context.Context, deviceID repos.ID) (bool, error) {
	err := r.Domain.RemoveDevice(ctx, deviceID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CreateProject(ctx context.Context, accountID repos.ID, name string, displayName string, cluster string, logo *string, description *string) (*model.Project, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CiAppJobHook(ctx context.Context, appID repos.ID, hasStarted *bool, hasCompleted *bool, error *string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateProject(ctx context.Context, projectID repos.ID, displayName *string, cluster *string, logo *string, description *string) (*model.Project, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteProject(ctx context.Context, projectID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) InviteProjectMember(ctx context.Context, projectID repos.ID, email string, name string, role string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) RemoveProjectMember(ctx context.Context, projectID repos.ID, userID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateProjectMember(ctx context.Context, projectID repos.ID, userID repos.ID, role string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) GithubEvent(ctx context.Context, installationID repos.ID, sourceRepo string) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) GitlabEvent(ctx context.Context, email repos.ID, sourceRepo string) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateAppFlow(ctx context.Context, projectID repos.ID, app *string, configs *string, secrets *string, mServices *string, mResources *string) (*model.Job, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateApp(ctx context.Context, projectID repos.ID, name string, description *string, services []*model.AppServiceInput, replicas *int, containers []*model.AppContainerIn) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) PatchApp(ctx context.Context, appID repos.ID, patch []*model.JSONPatchInput) (*model.PatchApp, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateApp(ctx context.Context, appID repos.ID, name *string, description *string, service *model.AppServiceInput, replicas *int, containers *model.AppContainerIn) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) RollbackApp(ctx context.Context, appID repos.ID, version int) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteApp(ctx context.Context, appID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) AddAppContainer(ctx context.Context, appID repos.ID, container model.AppContainerIn) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateAppContainer(ctx context.Context, appID repos.ID, containerName string, container model.AppContainerUpdateInput) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) RemoveAppContainer(ctx context.Context, appID repos.ID, containerName string) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) AddVolume(ctx context.Context, appID repos.ID, containerName string, volume model.IAppVolume) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateVolume(ctx context.Context, appID repos.ID, containerName string, volumeName string, volume model.IAppVolumeUpdate) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteVolume(ctx context.Context, appID repos.ID, containerName string, volumeName string) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateSecret(ctx context.Context, projectID repos.ID, name string, data *string) (*model.Secret, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpsertSecretEntry(ctx context.Context, secretID repos.ID, key string, value string) (*model.Secret, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) RemoveSecretEntry(ctx context.Context, secretID repos.ID, key string) (*model.Secret, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteSecret(ctx context.Context, secretID repos.ID) (*model.Secret, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateConfig(ctx context.Context, projectID repos.ID, name string, data *string) (*model.Config, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpsertConfigEntry(ctx context.Context, configID repos.ID, key string, value string) (*model.Config, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) RemoveConfigEntry(ctx context.Context, configID repos.ID, key string) (*model.Config, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteConfig(ctx context.Context, configID repos.ID) (*model.Config, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateGitPipeline(ctx context.Context, projectID repos.ID, name string, gitRepoURL string, gitProvider string, dockerFile *string, contextDir *string, buildArgs []*model.KVInput, pipelineEnv string, pullSecret *string) (*model.GitPipeline, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteGitPipeline(ctx context.Context, pipelineID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateRouter(ctx context.Context, projectID repos.ID, name string, routes []*model.RouteInput) (*model.Router, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) AddRoute(ctx context.Context, routerID repos.ID, route model.RouteInput) (*model.Router, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) UpdateRoute(ctx context.Context, routerID repos.ID, routeID repos.ID, route model.RouteInput) (*model.Router, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) RemoveRoute(ctx context.Context, routerID repos.ID, routeID repos.ID) (*model.Router, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) DeleteRouter(ctx context.Context, routerID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Projects(ctx context.Context, accountID *repos.ID) ([]*model.Project, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Project(ctx context.Context, projectID repos.ID) (*model.Project, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ProjectsMemberships(ctx context.Context, accountID *repos.ID) ([]*model.ProjectMembership, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ProjectMemberships(ctx context.Context, projectID repos.ID) (*model.ProjectMembership, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Apps(ctx context.Context, projectID repos.ID, search *string) ([]*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) App(ctx context.Context, appID repos.ID, version *string) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Routers(ctx context.Context, projectID repos.ID, search *string) ([]*model.Router, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Router(ctx context.Context, routerID *repos.ID, projectID *repos.ID, routerName *string) (*model.Router, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Configs(ctx context.Context, projectID repos.ID) ([]*model.Config, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) SearchConfigs(ctx context.Context, projectID repos.ID, search *string) ([]*model.CSEntry, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Config(ctx context.Context, configID *repos.ID, projectID *repos.ID, configName *string) (*model.Config, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Secrets(ctx context.Context, projectID repos.ID, search *string) ([]*model.Secret, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) SearchSecrets(ctx context.Context, projectID repos.ID, search *string) ([]*model.CSEntry, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Secret(ctx context.Context, secretID *repos.ID, projectID *repos.ID, secretName *string) (*model.Secret, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GitPullRepoToken(ctx context.Context, imageID repos.ID) (*string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GitlabRepos(ctx context.Context, groupID repos.ID, search *string, limit *int, page *int) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GitlabGroups(ctx context.Context, search *string, limit *int, page *int) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GitlabRepoBranches(ctx context.Context, repoURL string, search *string) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GithubInstallations(ctx context.Context) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GithubRepos(ctx context.Context, installationID string, limit *int, page *int) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GithubRepoBranches(ctx context.Context, repoURL string, limit *int, page *int) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) SearchGithubRepos(ctx context.Context, search *string, org string, limit *int, page *int) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GitPipelines(ctx context.Context, projectID repos.ID, query *string) ([]*model.GitPipeline, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GitPipeline(ctx context.Context, pipelineID repos.ID) (*model.GitPipeline, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ManagedSvcListAvailable(ctx context.Context) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ManagedSvcGetInstallation(ctx context.Context, installationID repos.ID, nextVersion *bool) (*model.ManagedSvc, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ManagedSvcListInstallations(ctx context.Context, projectID repos.ID) ([]*model.ManagedSvc, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ManagedResGetResource(ctx context.Context, resID repos.ID, nextVersion *bool) (*model.ManagedRes, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ManagedResListResources(ctx context.Context, installationID repos.ID) ([]*model.ManagedRes, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Jobs(ctx context.Context, projectID repos.ID, search *string) ([]*model.Job, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) Job(ctx context.Context, jobID repos.ID) (*model.Job, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ListClusters(ctx context.Context) ([]*model.Cluster, error) {
	clusterEntities, e := r.Domain.ListClusters(ctx)
	clusters := make([]*model.Cluster, 0)

	for _, cE := range clusterEntities {
		clusters = append(clusters, &model.Cluster{
			ID:         cE.Id,
			Name:       cE.Name,
			Provider:   cE.Provider,
			Region:     cE.Region,
			IP:         cE.Ip,
			NodesCount: cE.NodesCount,
			Status:     string(cE.Status),
		})
	}
	fmt.Println(clusters)
	return clusters, e
}

func (r *queryResolver) GetCluster(ctx context.Context, clusterID repos.ID) (*model.Cluster, error) {
	var e error
	defer wErrors.HandleErr(&e)
	clusterEntity, e := r.Domain.GetCluster(ctx, clusterID)
	wErrors.AssertNoError(e, fmt.Errorf("not able to get cluster"))
	return &model.Cluster{
		ID:         clusterEntity.Id,
		Name:       clusterEntity.Name,
		NodesCount: clusterEntity.NodesCount,
		Status:     string(clusterEntity.Status),
	}, e
}

func (r *queryResolver) GetDevice(ctx context.Context, deviceID repos.ID) (*model.Device, error) {
	var e error
	defer wErrors.HandleErr(&e)
	deviceEntity, e := r.Domain.GetDevice(ctx, deviceID)
	wErrors.AssertNoError(e, fmt.Errorf("not able to get device"))
	return &model.Device{
		ID:   deviceEntity.Id,
		Name: deviceEntity.Name,
		//Configuration: "",
	}, e
}

func (r *userResolver) Devices(ctx context.Context, obj *model.User) ([]*model.Device, error) {
	var e error
	defer wErrors.HandleErr(&e)
	user := obj
	deviceEntities, e := r.Domain.ListUserDevices(ctx, repos.ID(user.ID))
	wErrors.AssertNoError(e, fmt.Errorf("not able to list devices of user %s", user.ID))
	devices := make([]*model.Device, 0)
	for _, device := range deviceEntities {
		devices = append(devices, &model.Device{
			ID:            device.Id,
			Name:          device.Name,
			Configuration: "",
		})
	}
	return devices, e
}

// Cluster returns generated.ClusterResolver implementation.
func (r *Resolver) Cluster() generated.ClusterResolver { return &clusterResolver{r} }

// Device returns generated.DeviceResolver implementation.
func (r *Resolver) Device() generated.DeviceResolver { return &deviceResolver{r} }

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// User returns generated.UserResolver implementation.
func (r *Resolver) User() generated.UserResolver { return &userResolver{r} }

type clusterResolver struct{ *Resolver }
type deviceResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type userResolver struct{ *Resolver }
