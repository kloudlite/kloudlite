package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"kloudlite.io/apps/wireguard/internal/app/graph/generated"
	"kloudlite.io/apps/wireguard/internal/app/graph/model"
	"kloudlite.io/apps/wireguard/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func (r *clusterResolver) Devices(ctx context.Context, cluster *model.Cluster) ([]*model.Device, error) {
	devices, err := r.Domain.ListClusterDevices(ctx, cluster.ID)
	if err != nil {
		return nil, err
	}
	res := make([]*model.Device, len(devices))
	for i, d := range devices {
		res[i] = &model.Device{
			ID:            d.Id,
			Name:          d.Name,
			Cluster:       cluster,
			Configuration: "",
		}
	}
	return res, err
}

func (r *deviceResolver) User(ctx context.Context, device *model.Device) (*model.User, error) {
	deviceEntity, err := r.Domain.GetDevice(ctx, device.ID)
	if err != nil {
		return nil, err
	}
	return &model.User{
		ID: deviceEntity.UserId,
	}, nil
}

func (r *deviceResolver) Cluster(ctx context.Context, obj *model.Device) (*model.Cluster, error) {
	return r.Query().GetCluster(ctx, obj.Cluster.ID)
}

func (r *mutationResolver) CreateCluster(ctx context.Context, name string) (*model.Cluster, error) {
	cluster, e := r.Domain.CreateCluster(ctx, entities.Cluster{
		Name: name,
	})

	if e != nil {
		return nil, e
	}

	return &model.Cluster{
		ID:       cluster.Id,
		Name:     cluster.Name,
		Endpoint: cluster.Address,
	}, nil
}

func (r *mutationResolver) AddDevice(ctx context.Context, clusterID repos.ID, userID repos.ID, name string) (*model.Device, error) {
	device, err := r.Domain.AddDevice(ctx, name, clusterID, userID)
	if err != nil {
		return nil, err
	}
	return &model.Device{
		ID:            device.Id,
		Name:          device.Name,
		Configuration: "",
	}, err
}

func (r *mutationResolver) SetupCluster(ctx context.Context, clusterID repos.ID, address string, listenPort int, netInterface string) (*model.Cluster, error) {
	clusterEntity, err := r.Domain.SetupCluster(ctx, clusterID, address, uint16(listenPort), netInterface)
	if err != nil {
		return nil, err
	}
	return &model.Cluster{
		ID:       clusterEntity.Id,
		Name:     clusterEntity.Name,
		Endpoint: clusterEntity.Address,
	}, err
}

func (r *queryResolver) ListClusters(ctx context.Context) ([]*model.Cluster, error) {
	return make([]*model.Cluster, 0), nil
}

func (r *queryResolver) GetCluster(ctx context.Context, clusterID repos.ID) (*model.Cluster, error) {
	clusterEntity, err := r.Domain.GetCluster(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	return &model.Cluster{
		ID:       clusterEntity.Id,
		Name:     clusterEntity.Name,
		Endpoint: clusterEntity.Address,
	}, err
}

func (r *queryResolver) ListUserDevices(ctx context.Context, userID repos.ID) ([]*model.Device, error) {
	devices, err := r.Domain.ListUserDevices(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := make([]*model.Device, 0)
	for _, device := range devices {
		resp = append(resp, &model.Device{
			ID:            device.Id,
			Name:          device.Name,
			Configuration: "",
		})
	}
	return resp, err
}

func (r *queryResolver) GetDevice(ctx context.Context, deviceID repos.ID) (*model.Device, error) {
	device, err := r.Domain.GetDevice(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	return &model.Device{
		ID:            device.Id,
		Name:          device.Name,
		Configuration: "",
	}, err
}

func (r *userResolver) Devices(ctx context.Context, obj *model.User) ([]*model.Device, error) {
	devices, err := r.Domain.ListUserDevices(ctx, obj.ID)
	if err != nil {
		return nil, err
	}
	resp := make([]*model.Device, 0)
	for _, device := range devices {
		resp = append(resp, &model.Device{
			ID:            device.Id,
			Name:          device.Name,
			Configuration: "",
		})
	}
	fmt.Println(devices)
	return resp, err
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
