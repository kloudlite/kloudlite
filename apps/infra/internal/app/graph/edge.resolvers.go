package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	v11 "github.com/kloudlite/cluster-operator/apis/infra/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/app/graph/model"
	"kloudlite.io/apps/infra/internal/domain/entities"
)

func (r *edgeResolver) Metadata(ctx context.Context, obj *entities.Edge) (*v1.ObjectMeta, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *edgeSpecResolver) Pools(ctx context.Context, obj *v11.EdgeSpec) (*model.EdgeSpecPools, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *edgeInResolver) Metadata(ctx context.Context, obj *entities.Edge, data *v1.ObjectMeta) error {
	panic(fmt.Errorf("not implemented"))
}

func (r *edgeInResolver) Spec(ctx context.Context, obj *entities.Edge, data *model.EdgeSpecIn) error {
	panic(fmt.Errorf("not implemented"))
}

// Edge returns generated.EdgeResolver implementation.
func (r *Resolver) Edge() generated.EdgeResolver { return &edgeResolver{r} }

// EdgeSpec returns generated.EdgeSpecResolver implementation.
func (r *Resolver) EdgeSpec() generated.EdgeSpecResolver { return &edgeSpecResolver{r} }

// EdgeIn returns generated.EdgeInResolver implementation.
func (r *Resolver) EdgeIn() generated.EdgeInResolver { return &edgeInResolver{r} }

type edgeResolver struct{ *Resolver }
type edgeSpecResolver struct{ *Resolver }
type edgeInResolver struct{ *Resolver }
