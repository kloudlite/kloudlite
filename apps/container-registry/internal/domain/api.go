package domain

import (
	"context"

	"kloudlite.io/apps/container-registry/internal/domain/entities"
	"kloudlite.io/pkg/repos"
)

func NewRegistryContext(parent context.Context, userId repos.ID, accountName string) RegistryContext {
	return RegistryContext{
		Context:     parent,
		UserId:      userId,
		AccountName: accountName,
	}
}

type Domain interface {
	// registry
	ListRepositories(ctx RegistryContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Repository], error)
	CreateRepository(ctx RegistryContext, repoName string) error
	DeleteRepository(ctx RegistryContext, repoName string) error

	// tags
	ListRepositoryTags(ctx RegistryContext, repoName string, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Tag], error)
	DeleteRepositoryTag(ctx RegistryContext, repoName string, digest string) error

	// credential
	ListCredentials(ctx RegistryContext, search map[string]repos.MatchFilter, pagination repos.CursorPagination) (*repos.PaginatedRecord[*entities.Credential], error)
	CreateCredential(ctx RegistryContext, credential entities.Credential) error
	DeleteCredential(ctx RegistryContext, userName string) error

	ProcessEvents(ctx context.Context, events []entities.Event) error

	GetToken(ctx RegistryContext, username string) (string, error)
	GetTokenKey(ctx context.Context, username string, accountname string) (string, error)
}
