package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"github.com/kloudlite/cluster-operator/apis/infra/v1"
	v11 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/app/graph/model"
	"kloudlite.io/apps/infra/internal/domain/entities"
)

func (r *cloudProviderResolver) Metadata(ctx context.Context, obj *entities.CloudProvider) (*v11.ObjectMeta, error) {
	return &obj.ObjectMeta, nil
}

func (r *cloudProviderSpecResolver) ProviderSecret(ctx context.Context, obj *v1.CloudProviderSpec) (*model.CloudProviderSpecProviderSecret, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *cloudProviderInResolver) Metadata(ctx context.Context, obj *entities.CloudProvider, data *v11.ObjectMeta) error {
	*data = obj.ObjectMeta
	return nil
}

func (r *cloudProviderInResolver) Spec(ctx context.Context, obj *entities.CloudProvider, data *model.CloudProviderSpecIn) error {
	panic(fmt.Errorf("not implemented"))
}

// CloudProvider returns generated.CloudProviderResolver implementation.
func (r *Resolver) CloudProvider() generated.CloudProviderResolver { return &cloudProviderResolver{r} }

// CloudProviderSpec returns generated.CloudProviderSpecResolver implementation.
func (r *Resolver) CloudProviderSpec() generated.CloudProviderSpecResolver {
	return &cloudProviderSpecResolver{r}
}

// CloudProviderIn returns generated.CloudProviderInResolver implementation.
func (r *Resolver) CloudProviderIn() generated.CloudProviderInResolver {
	return &cloudProviderInResolver{r}
}

type cloudProviderResolver struct{ *Resolver }
type cloudProviderSpecResolver struct{ *Resolver }
type cloudProviderInResolver struct{ *Resolver }
