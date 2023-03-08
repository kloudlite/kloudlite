package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"kloudlite.io/apps/infra/internal/app/graph/generated"
	"kloudlite.io/apps/infra/internal/app/graph/model"
	"kloudlite.io/apps/infra/internal/domain/entities"
)

func (r *cloudProviderResolver) Status(ctx context.Context, obj *entities.CloudProvider) (*model.Status, error) {
	if obj == nil {
		return nil, nil
	}
	return toModelStatus(obj.Status)
}

// CloudProvider returns generated.CloudProviderResolver implementation.
func (r *Resolver) CloudProvider() generated.CloudProviderResolver { return &cloudProviderResolver{r} }

type cloudProviderResolver struct{ *Resolver }
