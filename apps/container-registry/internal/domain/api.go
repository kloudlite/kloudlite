package domain

import (
	"context"
	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/pkg/harbor"
	"kloudlite.io/pkg/repos"
	opHarbor "github.com/kloudlite/operator/pkg/harbor"
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
	GetRepoArtifacts(ctx RegistryContext, repoName string) ([]harbor.Artifact, error)

	// query:harbor robots
	ListHarborRobots(ctx RegistryContext) ([]*entities.HarborRobotUser, error)

	// mutation:harbor robots
	CreateHarborRobot(ctx RegistryContext, hru *entities.HarborRobotUser) (*entities.HarborRobotUser, error)
	DeleteHarborRobot(ctx RegistryContext, robotId int) error
	UpdateHarborRobot(ctx RegistryContext, name string, permissions []opHarbor.Permission) (*entities.HarborRobotUser, error)
	ReSyncHarborRobot(ctx RegistryContext, name string) error

	// CreateHarborProject(ctx RegistryContext) (*entities.HarborProject, error)
	GetHarborCredentials(ctx RegistryContext) (*entities.HarborCredentials, error)
}
