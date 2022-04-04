package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"kloudlite.io/pkg/repos"

	"kloudlite.io/apps/wireguard/internal/app/graph/generated"
	"kloudlite.io/apps/wireguard/internal/app/graph/model"
	"kloudlite.io/apps/wireguard/internal/domain/entities"
)

func (r *mutationResolver) CreateCluster(ctx context.Context, name string) (*model.Cluster, error) {

	cluster, e := r.Domain.CreateCluster(ctx, entities.Cluster{
		Name: name,
	})

	if e != nil {
		return nil, e
	}

	return &model.Cluster{
		ID:       string(cluster.Id),
		Name:     cluster.Name,
		Endpoint: cluster.Address,
	}, nil
}

func (r *mutationResolver) AddDevice(ctx context.Context, clusterID string, userID string, name string) (*model.Device, error) {
	device, err := r.Domain.AddDevice(ctx, name, repos.ID(clusterID), repos.ID(userID))
	return &model.Device{
		ID:            string(device.Id),
		UserID:        string(device.UserId),
		Name:          device.Name,
		Configuration: "",
	}, err
}

func (r *queryResolver) ListClusters(ctx context.Context) ([]*model.Cluster, error) {
	return make([]*model.Cluster, 0), nil
}

func (r *queryResolver) GetCluster(ctx context.Context, clusterID string) (*model.Cluster, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) ListDevices(ctx context.Context) ([]*model.Device, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) GetDevice(ctx context.Context, deviceID string) (*model.Device, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
