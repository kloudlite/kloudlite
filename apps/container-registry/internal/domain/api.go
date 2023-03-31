package domain

import (
	"context"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/pkg/harbor"
	"kloudlite.io/pkg/repos"
)

func NewRegistryContext(parent context.Context, userId repos.ID, accountName string) RegistryContext {
	return RegistryContext{
		Context:     parent,
		userId:      userId,
		accountName: accountName,
	}
}

type Domain interface {
	GetHarborImages(ctx RegistryContext) ([]harbor.Repository, error)
	GetImageTags(ctx RegistryContext, repoName string) ([]harbor.ImageTag, error)
	GetHarborRobots(ctx RegistryContext) ([]harbor.Robot, error)
	DeleteHarborRobot(ctx RegistryContext, robotId int) error
	CreateHarborRobot(ctx RegistryContext, name string, description *string, readOnly bool) (*harbor.Robot, error)

	CreateHarborProject(ctx RegistryContext) (*entities.HarborProject, error)
	GetHarborCredentials(ctx RegistryContext) (*entities.HarborCredentials, error)
}
