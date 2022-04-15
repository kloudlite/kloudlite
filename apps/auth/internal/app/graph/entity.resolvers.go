package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"kloudlite.io/apps/auth/internal/app/graph/generated"
	"kloudlite.io/apps/auth/internal/app/graph/model"
	"kloudlite.io/pkg/repos"
)

func (r *entityResolver) FindUserByID(ctx context.Context, id repos.ID) (*model.User, error) {
	userEntity, err := r.d.GetUserById(ctx, id)
	return userModelFromEntity(userEntity), err
}

// Entity returns generated.EntityResolver implementation.
func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
