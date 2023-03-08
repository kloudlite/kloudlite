package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/app/graph/model"
	"kloudlite.io/apps/infra/internal/domain/entities"
)

func (r *clusterResolver) Status(ctx context.Context, obj *entities.Cluster) (*model.Status, error) {
	if obj == nil {
		return nil, nil
	}
	return toModelStatus(obj.Status)
}

// Cluster returns generated.ClusterResolver implementation.
func (r *Resolver) Cluster() generated.ClusterResolver { return &clusterResolver{r} }

type clusterResolver struct{ *Resolver }
