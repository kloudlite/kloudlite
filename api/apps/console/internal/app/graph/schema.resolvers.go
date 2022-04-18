package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/app/graph/model"
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

func (r *mutationResolver) MangedSvcUninstall(ctx context.Context, installationID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) MangedSvcUpdate(ctx context.Context, installationID repos.ID, values string) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ManagedResCreate(ctx context.Context, installationID repos.ID, name string, resourceName string, values string) (*model.ManagedRes, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) ManagedResUpdate(ctx context.Context, resID repos.ID, values *string) (bool, error) {
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

func (r *mutationResolver) InfraAddDevice(ctx context.Context, clusterID repos.ID, userID repos.ID, name string) (*model.Device, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) InfraRemoveDevice(ctx context.Context, deviceID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreCreateProject(ctx context.Context, accountID repos.ID, name string, displayName string, cluster string, logo *string, description *string) (*model.Project, error) {
	panic(fmt.Errorf("not implemented"))
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

func (r *mutationResolver) GithubEvent(ctx context.Context, installationID repos.ID, sourceRepo string) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) GitlabEvent(ctx context.Context, email repos.ID, sourceRepo string) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreCreateAppFlow(ctx context.Context, projectID repos.ID, app string, pipelines *model.GitPipelineInput, configs *string, secrets *string, mServices *string, mResources *string) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreUpdateApp(ctx context.Context, appID repos.ID, name *string, description *string, service *model.AppServiceInput, replicas *int, containers *model.AppContainerIn) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreDeleteApp(ctx context.Context, appID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreRollbackApp(ctx context.Context, appID repos.ID, version int) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreCreateSecret(ctx context.Context, projectID repos.ID, name string, data *string) (*model.Secret, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreUpdateSecret(ctx context.Context, projectID repos.ID, name string, data *string) (*model.Secret, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreDeleteSecret(ctx context.Context, projectID repos.ID, name string) (*model.Secret, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreCreateConfig(ctx context.Context, projectID repos.ID, name string, data *string) (*model.Config, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreUpdateConfig(ctx context.Context, projectID repos.ID, name string, data *string) (*model.Config, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreDeleteConfig(ctx context.Context, configID repos.ID) (*model.Config, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CiDeleteGitPipeline(ctx context.Context, pipelineID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreCreateRouter(ctx context.Context, projectID repos.ID, name string, domains []*string, routes []*model.RouteInput) (*model.Router, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreUpdateRouter(ctx context.Context, routerID repos.ID, name string, domains []*string, routes []*model.RouteInput) (*bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CoreDeleteRouter(ctx context.Context, routerID repos.ID) (bool, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CoreProjects(ctx context.Context, accountID *repos.ID) ([]*model.Project, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CoreProject(ctx context.Context, accountID *repos.ID, projectID repos.ID) (*model.Project, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CoreApps(ctx context.Context, projectID repos.ID, search *string) ([]*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CoreApp(ctx context.Context, appID repos.ID, version *string) (*model.App, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CoreRouters(ctx context.Context, projectID repos.ID, search *string) ([]*model.Router, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CoreRouter(ctx context.Context, routerID *repos.ID, projectID *repos.ID) (*model.Router, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CoreConfigs(ctx context.Context, projectID repos.ID, search *string) ([]*model.Config, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CoreConfig(ctx context.Context, configID *repos.ID, projectID *repos.ID) (*model.Config, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CoreSecrets(ctx context.Context, projectID repos.ID, search *string) ([]*model.Secret, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CoreSecret(ctx context.Context, secretID *repos.ID, projectID *repos.ID, secretName *string) (*model.Secret, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGitPullRepoToken(ctx context.Context, imageID repos.ID) (*string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGitlabRepos(ctx context.Context, groupID repos.ID, search *string, limit *int, page *int) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGitlabGroups(ctx context.Context, search *string, limit *int, page *int) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGitlabRepoBranches(ctx context.Context, repoURL string, search *string) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGithubInstallations(ctx context.Context) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGithubRepos(ctx context.Context, installationID string, limit *int, page *int) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGithubRepoBranches(ctx context.Context, repoURL string, limit *int, page *int) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiSearchGithubRepos(ctx context.Context, search *string, org string, limit *int, page *int) ([]string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGitPipelines(ctx context.Context, projectID repos.ID, query *string) ([]*model.GitPipeline, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CiGitPipeline(ctx context.Context, pipelineID repos.ID) (*model.GitPipeline, error) {
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

func (r *queryResolver) InfraListClusters(ctx context.Context) ([]*model.Cluster, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) InfraGetCluster(ctx context.Context, clusterID repos.ID) (*model.Cluster, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) InfraGetDevices(ctx context.Context, deviceID repos.ID) (*model.Device, error) {
	panic(fmt.Errorf("not implemented"))
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
