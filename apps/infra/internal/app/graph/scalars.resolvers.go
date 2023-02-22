package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	operator1 "github.com/kloudlite/cluster-operator/lib/operator"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/app/graph/model"
)

func (r *metadataResolver) Labels(ctx context.Context, obj *v1.ObjectMeta) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *statusResolver) Checks(ctx context.Context, obj *operator1.Status) ([]*model.Check, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *statusResolver) DisplayVars(ctx context.Context, obj *operator1.Status) (map[string]interface{}, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *metadataInResolver) Labels(ctx context.Context, obj *v1.ObjectMeta, data map[string]interface{}) error {
	panic(fmt.Errorf("not implemented"))
}

// Metadata returns generated.MetadataResolver implementation.
func (r *Resolver) Metadata() generated.MetadataResolver { return &metadataResolver{r} }

// Status returns generated.StatusResolver implementation.
func (r *Resolver) Status() generated.StatusResolver { return &statusResolver{r} }

// MetadataIn returns generated.MetadataInResolver implementation.
func (r *Resolver) MetadataIn() generated.MetadataInResolver { return &metadataInResolver{r} }

type metadataResolver struct{ *Resolver }
type statusResolver struct{ *Resolver }
type metadataInResolver struct{ *Resolver }
