package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"

	"kloudlite.io/apps/console/internal/app/graph/generated"
	"kloudlite.io/apps/console/internal/app/graph/model"
	"kloudlite.io/apps/console/internal/domain/entities"
	fn "kloudlite.io/pkg/functions"
)

func (r *projectResolver) Spec(ctx context.Context, obj *entities.Project) (*model.ProjectSpec, error) {
	if obj == nil {
		return nil, nil
	}
	var m model.ProjectSpec
	if err := fn.JsonConversion(obj.Spec, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *projectInResolver) Spec(ctx context.Context, obj *entities.Project, data *model.ProjectSpecIn) error {
	panic(fmt.Errorf("not implemented"))
}

// Project returns generated.ProjectResolver implementation.
func (r *Resolver) Project() generated.ProjectResolver { return &projectResolver{r} }

// ProjectIn returns generated.ProjectInResolver implementation.
func (r *Resolver) ProjectIn() generated.ProjectInResolver { return &projectInResolver{r} }

type projectResolver struct{ *Resolver }
type projectInResolver struct{ *Resolver }
