package graph

import (
	"context"
	"fmt"
	"kloudlite.io/pkg/repos"

	"kloudlite.io/apps/accounts/internal/domain"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	domain domain.Domain
}

func NewResolver(domain domain.Domain) *Resolver {
	return &Resolver{
		domain: domain,
	}
}

func toUserContext(ctx context.Context) (domain.UserContext, error) {
	userId, ok := ctx.Value("kloudlite-user-id").(string)
	if !ok {
		return domain.UserContext{}, fmt.Errorf("`kloudlite-user-id` not set in request context")
	}

	userEmail, ok := ctx.Value("kloudlite-user-email").(string)
	if !ok {
		return domain.UserContext{}, fmt.Errorf("`kloudlite-user-email` not set in request context")
	}

	return domain.UserContext{Context: ctx, UserId: repos.ID(userId), UserEmail: userEmail}, nil
}
