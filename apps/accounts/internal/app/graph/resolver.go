package graph

import (
	"context"
	"fmt"
	"github.com/kloudlite/api/apps/accounts/internal/domain"
	"github.com/kloudlite/api/common"
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
	sess, ok := ctx.Value("kloudlite-user-session").(common.AuthSession)
	if !ok {
		return domain.UserContext{}, fmt.Errorf("`kloudlite-user-session` not set in request context")
	}

	return domain.UserContext{
		Context:   ctx,
		UserId:    sess.UserId,
		UserName:  sess.UserName,
		UserEmail: sess.UserEmail,
	}, nil
}
