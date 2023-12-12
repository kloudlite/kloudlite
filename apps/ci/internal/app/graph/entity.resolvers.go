package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"

	"github.com/kloudlite/api/apps/ci/internal/app/graph/generated"
	"github.com/kloudlite/api/apps/ci/internal/app/graph/model"
	"github.com/kloudlite/api/pkg/repos"
)

func (r *entityResolver) FindAppByID(ctx context.Context, id repos.ID) (*model.App, error) {
	return &model.App{
		ID: id,
	}, nil
}

// Entity returns generated.EntityResolver implementation.
func (r *Resolver) Entity() generated.EntityResolver { return &entityResolver{r} }

type entityResolver struct{ *Resolver }
