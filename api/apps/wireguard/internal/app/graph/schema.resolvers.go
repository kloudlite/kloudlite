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
		UserID:        device.UserId,
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
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ListDevices(ctx context.Context) ([]*model.Device, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetDevice(ctx context.Context, deviceID repos.ID) (*model.Device, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
