package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/app/graph/model"
	"kloudlite.io/apps/infra/internal/domain/entities"
)

func (r *clusterResolver) Metadata(ctx context.Context, obj *entities.Cluster) (*v1.ObjectMeta, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *clusterInResolver) Metadata(ctx context.Context, obj *entities.Cluster, data *v1.ObjectMeta) error {
	panic(fmt.Errorf("not implemented"))
}

func (r *clusterInResolver) Spec(ctx context.Context, obj *entities.Cluster, data *model.ClusterSpecIn) error {
	panic(fmt.Errorf("not implemented"))
}

// Cluster returns generated.ClusterResolver implementation.
func (r *Resolver) Cluster() generated.ClusterResolver { return &clusterResolver{r} }

// ClusterIn returns generated.ClusterInResolver implementation.
func (r *Resolver) ClusterIn() generated.ClusterInResolver { return &clusterInResolver{r} }

type clusterResolver struct{ *Resolver }
type clusterInResolver struct{ *Resolver }
