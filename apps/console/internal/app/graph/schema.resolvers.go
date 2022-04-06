package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/app/graph/model"
	"kloudlite.io/apps/console/internal/domain/entities"
	"kloudlite.io/pkg/errors"
	"kloudlite.io/pkg/repos"
)

func (r *clusterResolver) Devices(ctx context.Context, obj *model.Cluster) ([]*model.Device, error) {
	var e error
	defer errors.HandleErr(&e)
	cluster := obj
	deviceEntities, e := r.Domain.ListClusterDevices(ctx, cluster.ID)
	errors.AssertNoError(e, fmt.Errorf("not able to list devices of cluster %s", cluster.ID))
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

func (r *clusterResolver) Configuration(ctx context.Context, obj *model.Cluster) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *deviceResolver) User(ctx context.Context, obj *model.Device) (*model.User, error) {
	var e error
	defer errors.HandleErr(&e)
	device := obj
	deviceEntity, e := r.Domain.GetDevice(ctx, device.ID)
	errors.AssertNoError(e, fmt.Errorf("not able to get device"))
	return &model.User{
		ID: deviceEntity.UserId,
	}, e
}

func (r *deviceResolver) Cluster(ctx context.Context, obj *model.Device) (*model.Cluster, error) {
	return r.Query().GetCluster(ctx, obj.Cluster.ID)
}

func (r *deviceResolver) Configuration(ctx context.Context, obj *model.Device) (string, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *mutationResolver) CreateCluster(ctx context.Context, name string, provider string, region string) (*model.Cluster, error) {
	var e error
	defer errors.HandleErr(&e)
	cluster, e := r.Domain.CreateCluster(ctx, entities.Cluster{
		Name:     name,
		Provider: provider,
		Region:   region,
	})
	errors.AssertNoError(e, fmt.Errorf("not able to create cluster"))
	return &model.Cluster{
		ID:       cluster.Id,
		Name:     cluster.Name,
		Endpoint: cluster.Address,
		Provider: cluster.Provider,
		Region:   cluster.Region,
	}, e
}

func (r *mutationResolver) AddDevice(ctx context.Context, clusterID repos.ID, userID repos.ID, name string) (*model.Device, error) {
	var e error
	defer errors.HandleErr(&e)
	device, e := r.Domain.AddDevice(ctx, name, clusterID, userID)
	errors.AssertNoError(e, fmt.Errorf("not able to add device"))
	return &model.Device{
		ID:            device.Id,
		Name:          device.Name,
		Configuration: "",
	}, e
}

func (r *mutationResolver) SetupCluster(ctx context.Context, clusterID repos.ID, address string, listenPort int, netInterface string) (*model.Cluster, error) {
	var e error
	defer errors.HandleErr(&e)
	clusterEntity, e := r.Domain.SetupCluster(ctx, clusterID, address, uint16(listenPort), netInterface)
	errors.AssertNoError(e, fmt.Errorf("not able to setup cluster"))
	return &model.Cluster{
		ID:       clusterEntity.Id,
		Name:     clusterEntity.Name,
		Endpoint: clusterEntity.Address,
	}, e
}

func (r *queryResolver) ListClusters(ctx context.Context) ([]*model.Cluster, error) {
	return make([]*model.Cluster, 0), nil
}

func (r *queryResolver) GetCluster(ctx context.Context, clusterID repos.ID) (*model.Cluster, error) {
	var e error
	defer errors.HandleErr(&e)
	clusterEntity, e := r.Domain.GetCluster(ctx, clusterID)
	errors.AssertNoError(e, fmt.Errorf("not able to get cluster"))
	return &model.Cluster{
		ID:       clusterEntity.Id,
		Name:     clusterEntity.Name,
		Endpoint: clusterEntity.Address,
	}, e
}

func (r *queryResolver) GetDevice(ctx context.Context, deviceID repos.ID) (*model.Device, error) {
	var e error
	defer errors.HandleErr(&e)
	device, e := r.Domain.GetDevice(ctx, deviceID)
	errors.AssertNoError(e, fmt.Errorf("not able to get device"))
	return &model.Device{
		ID:            device.Id,
		Name:          device.Name,
		Configuration: "",
	}, e
}

func (r *userResolver) Devices(ctx context.Context, obj *model.User) ([]*model.Device, error) {
	var e error
	defer errors.HandleErr(&e)
	user := obj
	deviceEntities, e := r.Domain.ListUserDevices(ctx, repos.ID(user.ID))
	errors.AssertNoError(e, fmt.Errorf("not able to list devices of user %s", user.ID))
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
