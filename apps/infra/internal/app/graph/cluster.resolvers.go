package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/app/graph/model"
	"kloudlite.io/apps/infra/internal/domain/entities"
	fn "kloudlite.io/pkg/functions"
)

func (r *clusterResolver) Spec(ctx context.Context, obj *entities.Cluster) (*model.ClusterSpec, error) {
	var m model.ClusterSpec
	if err := fn.JsonConversion(obj.Spec, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *clusterInResolver) Spec(ctx context.Context, obj *entities.Cluster, data *model.ClusterSpecIn) error {
	if obj == nil {
		return nil
	}
	return fn.JsonConversion(data, &obj.Spec)
}

// Cluster returns generated.ClusterResolver implementation.
func (r *Resolver) Cluster() generated.ClusterResolver { return &clusterResolver{r} }

// ClusterIn returns generated.ClusterInResolver implementation.
func (r *Resolver) ClusterIn() generated.ClusterInResolver { return &clusterInResolver{r} }

type clusterResolver struct{ *Resolver }
type clusterInResolver struct{ *Resolver }
