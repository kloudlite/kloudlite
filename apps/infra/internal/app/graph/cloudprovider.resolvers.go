package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/app/graph/model"
	"kloudlite.io/apps/infra/internal/domain/entities"
	fn "kloudlite.io/pkg/functions"
)

func (r *cloudProviderResolver) SyncStatus(ctx context.Context, obj *entities.CloudProvider) (*entities.SyncStatus, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *cloudProviderResolver) Spec(ctx context.Context, obj *entities.CloudProvider) (*model.CloudProviderSpec, error) {
	var m model.CloudProviderSpec
	if err := fn.JsonConversion(obj.Spec, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *cloudProviderInResolver) Spec(ctx context.Context, obj *entities.CloudProvider, data *model.CloudProviderSpecIn) error {
	if obj == nil {
		return nil
	}
	return fn.JsonConversion(data, &obj.Spec)
}

// CloudProvider returns generated.CloudProviderResolver implementation.
func (r *Resolver) CloudProvider() generated.CloudProviderResolver { return &cloudProviderResolver{r} }

// CloudProviderIn returns generated.CloudProviderInResolver implementation.
func (r *Resolver) CloudProviderIn() generated.CloudProviderInResolver {
	return &cloudProviderInResolver{r}
}

type cloudProviderResolver struct{ *Resolver }
type cloudProviderInResolver struct{ *Resolver }
