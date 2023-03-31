package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	generated1 "kloudlite.io/apps/container-registry/internal/app/graph/generated"
	"kloudlite.io/pkg/harbor"
)

func (r *mutationResolver) CrDeleteRobot(ctx context.Context, robotID int) (bool, error) {
	err := r.Domain.DeleteHarborRobot(toRegistryContext(ctx), robotID)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *mutationResolver) CrCreateRobot(ctx context.Context, name string, description *string, readOnly bool) (*harbor.Robot, error) {
	return r.Domain.CreateHarborRobot(toRegistryContext(ctx), name, description, readOnly)
}

func (r *queryResolver) CrListRepos(ctx context.Context) ([]*harbor.Repository, error) {
	images, err := r.Domain.GetHarborImages(toRegistryContext(ctx))
	if err != nil {
		return nil, err
	}
	repositories := make([]*harbor.Repository, len(images))
	for i, image := range images {
		repositories[i] = &image
	}
	return repositories, nil
}

func (r *queryResolver) CrListRobots(ctx context.Context) ([]*harbor.Robot, error) {
	robots, err := r.Domain.GetHarborRobots(toRegistryContext(ctx))
	if err != nil {
		return nil, err
	}
	robotsReturn := make([]*harbor.Robot, len(robots))
	for i, robot := range robots {
		robotsReturn[i] = &robot
	}
	return robotsReturn, nil
}

// Mutation returns generated1.MutationResolver implementation.
func (r *Resolver) Mutation() generated1.MutationResolver { return &mutationResolver{r} }

// Query returns generated1.QueryResolver implementation.
func (r *Resolver) Query() generated1.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
