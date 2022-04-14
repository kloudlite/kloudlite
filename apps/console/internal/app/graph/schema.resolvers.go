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

func (r *queryResolver) Sample(ctx context.Context) (*string, error) {
	s, ok := ctx.Value("session").(*entities.AuthSession)
	if !ok {
		return nil, wErrors.Newf("could not retrieve session from ctx")
	}
	return &s.UserEmail, nil
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
